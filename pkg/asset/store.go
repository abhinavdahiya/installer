package asset

import (
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Store is a store for the states of assets.
type Store interface {
	// Fetch retrieves the state of the given asset, generating it and its
	// dependencies if necessary.
	Fetch(Asset) error

	// Destroy removes the asset from all its internal state and also from
	// disk if possible.
	Destroy(Asset) error
}

// assetSource indicates from where the asset was fetched
type assetSource int

const (
	// unfetched indicates that the asset has not been fetched
	unfetched assetSource = iota
	// generatedSource indicates that the asset was generated
	generatedSource
	// onDiskSource indicates that the asset was fetched from disk
	onDiskSource
	// stateFileSource indicates that the asset was fetched from the state file
	stateFileSource
)

type assetState struct {
	// asset is the asset.
	// If the asset has not been fetched, then this will be nil.
	asset Asset
	// source is the source from which the asset was fetched
	source assetSource
	// anyParentsDirty is true if any of the parents of the asset are dirty
	anyParentsDirty bool
	// presentOnDisk is true if the asset in on-disk. This is set whether the
	// asset is sourced from on-disk or not. It is used in purging consumed assets.
	presentOnDisk bool
}

// DiskStore is the implementation of Store.
type DiskStore struct {
	directory string

	assets    map[reflect.Type]*assetState
	stateFile *stateFile
}

// NewStore returns an asset store that implements the Store interface.
func NewStore(dir string) (Store, error) {
	store := &DiskStore{
		directory: dir,
		assets:    map[reflect.Type]*assetState{},

		stateFile: &stateFile{},
	}

	if err := store.stateFile.load(filepath.Join(store.directory, stateFileName)); err != nil {
		return nil, err
	}
	return store, nil
}

// Fetch retrieves the state of the given asset, generating it and its
// dependencies if necessary.
func (s *DiskStore) Fetch(asset Asset) error {
	if err := s.fetch(asset, ""); err != nil {
		return err
	}
	if err := s.save(); err != nil {
		return errors.Wrapf(err, "failed to save state")
	}
	return errors.Wrapf(s.purge(asset), "failed to purge asset")
}

// Destroy removes the asset from all its internal state and also from
// disk if possible.
func (s *DiskStore) Destroy(asset Asset) error {
	if sa, ok := s.assets[reflect.TypeOf(asset)]; ok {
		reflect.ValueOf(asset).Elem().Set(reflect.ValueOf(sa.asset).Elem())
	} else if s.stateFile.exists(asset) {
		if _, err := asset.Load(s.stateFile.fileFetcher(asset)); err != nil {
			return err
		}
	} else {
		// nothing to do
		return nil
	}

	if err := deleteAssetFromDisk(asset, s.directory); err != nil {
		return err
	}
	delete(s.assets, reflect.TypeOf(asset))
	delete(s.stateFile.Contents, assetToStateKey(asset))
	return s.save()
}

func (s *DiskStore) save() error {
	var assets []Asset
	for _, v := range s.assets {
		if v.source == unfetched {
			continue
		}
		assets = append(assets, v.asset)
	}
	return s.stateFile.save(filepath.Join(s.directory, stateFileName), assets)
}

// fetch populates the given asset, generating it and its dependencies if
// necessary, and returns whether or not the asset had to be regenerated and
// any errors.
func (s *DiskStore) fetch(asset Asset, indent string) error {
	logrus.Debugf("%sFetching %q...", indent, asset.Name())

	assetState, ok := s.assets[reflect.TypeOf(asset)]
	if !ok {
		if _, err := s.load(asset, ""); err != nil {
			return err
		}
		assetState = s.assets[reflect.TypeOf(asset)]
	}

	// Return immediately if the asset has been fetched before,
	// this is because we are doing a depth-first-search, it's guaranteed
	// that we always fetch the parent before children, so we don't need
	// to worry about invalidating anything in the cache.
	if assetState.source != unfetched {
		logrus.Debugf("%sReusing previously-fetched %q", indent, asset.Name())
		reflect.ValueOf(asset).Elem().Set(reflect.ValueOf(assetState.asset).Elem())
		return nil
	}

	// Re-generate the asset
	dependencies := asset.Dependencies()
	parents := make(Parents, len(dependencies))
	for _, d := range dependencies {
		if err := s.fetch(d, increaseIndent(indent)); err != nil {
			return errors.Wrapf(err, "failed to fetch dependency of %q", asset.Name())
		}
		parents.Add(d)
	}
	logrus.Debugf("%sGenerating %q...", indent, asset.Name())
	if err := asset.Generate(parents); err != nil {
		return errors.Wrapf(err, "failed to generate asset %q", asset.Name())
	}
	assetState.asset = asset
	assetState.source = generatedSource
	return nil
}

// load loads the asset and all of its ancestors from on-disk and the state file.
func (s *DiskStore) load(asset Asset, indent string) (*assetState, error) {
	logrus.Debugf("%sLoading %q...", indent, asset.Name())

	// Stop descent if the asset has already been loaded.
	if state, ok := s.assets[reflect.TypeOf(asset)]; ok {
		return state, nil
	}

	// Load dependencies from on-disk.
	anyParentsDirty := false
	for _, d := range asset.Dependencies() {
		state, err := s.load(d, increaseIndent(indent))
		if err != nil {
			return nil, err
		}
		if state.anyParentsDirty || state.source == onDiskSource {
			anyParentsDirty = true
		}
	}

	// Try to load from on-disk.
	onDiskAsset := reflect.New(reflect.TypeOf(asset).Elem()).Interface().(Asset)
	foundOnDisk, err := onDiskAsset.Load(&diskFileFetcher{s.directory})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load asset %q", asset.Name())
	}

	// Try to load from state file.
	var (
		stateFileAsset         Asset
		foundInStateFile       bool
		onDiskMatchesStateFile bool
	)
	// Do not need to bother with loading from state file if any of the parents
	// are dirty because the asset must be re-generated in this case.
	if !anyParentsDirty {
		foundInStateFile = s.stateFile.exists(asset)
		if foundInStateFile {
			stateFileAsset = reflect.New(reflect.TypeOf(asset).Elem()).Interface().(Asset)
			if _, err := stateFileAsset.Load(s.stateFile.fileFetcher(stateFileAsset)); err != nil {
				return nil, errors.Wrapf(err, "failed to load asset %q from state file", asset.Name())
			}
		}

		if foundOnDisk && foundInStateFile {
			logrus.Debugf("%sLoading %q from both state file and target directory", indent, asset.Name())

			// If the on-disk asset is the same as the one in the state file, there
			// is no need to consider the one on disk and to mark the asset dirty.
			onDiskMatchesStateFile = reflect.DeepEqual(onDiskAsset, stateFileAsset)
			if onDiskMatchesStateFile {
				logrus.Debugf("%sOn-disk %q matches asset in state file", indent, asset.Name())
			}
		}
	}

	var (
		assetToStore Asset
		source       assetSource
	)
	switch {
	// A parent is dirty. The asset must be re-generated.
	case anyParentsDirty:
		if foundOnDisk {
			logrus.Warningf("%sDiscarding the %q that was provided in the target directory because its dependencies are dirty and it needs to be regenerated", indent, asset.Name())
		}
		source = unfetched
	// The asset is on disk and that differs from what is in the source file.
	// The asset is sourced from on disk.
	case foundOnDisk && !onDiskMatchesStateFile:
		logrus.Debugf("%sUsing %q loaded from target directory", indent, asset.Name())
		assetToStore = onDiskAsset
		source = onDiskSource
	// The asset is in the state file. The asset is sourced from state file.
	case foundInStateFile:
		logrus.Debugf("%sUsing %q loaded from state file", indent, asset.Name())
		assetToStore = stateFileAsset
		source = stateFileSource
	// There is no existing source for the asset. The asset will be generated.
	default:
		source = unfetched
	}

	state := &assetState{
		asset:           assetToStore,
		source:          source,
		anyParentsDirty: anyParentsDirty,
		presentOnDisk:   foundOnDisk,
	}
	s.assets[reflect.TypeOf(asset)] = state
	return state, nil
}

// purge deletes the on-disk assets that are consumed already.
// E.g., install-config.yml will be deleted after fetching 'manifests'.
// The target asset is excluded.
func (s *DiskStore) purge(excluded Asset) error {
	for _, assetState := range s.assets {
		if !assetState.presentOnDisk {
			continue
		}
		if reflect.TypeOf(assetState.asset) == reflect.TypeOf(excluded) {
			continue
		}
		logrus.Infof("Consuming %q from target directory", assetState.asset.Name())
		if err := deleteAssetFromDisk(assetState.asset, s.directory); err != nil {
			return err
		}
		assetState.presentOnDisk = false
	}
	return nil
}

func increaseIndent(indent string) string {
	return indent + "  "
}

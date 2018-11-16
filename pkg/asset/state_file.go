package asset

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
)

const (
	stateFileName = ".openshift_install_state.json"
)

type stateFile struct {
	Version  string
	Contents map[string][]*File
}

// isAssetInState tests whether the asset is in the state file.
func (sf *stateFile) exists(asset Asset) bool {
	_, ok := sf.Contents[assetToStateKey(asset)]
	return ok
}

// load loads the state from path.
func (sf *stateFile) load(path string) error {
	if sf.Contents == nil {
		sf.Contents = map[string][]*File{}
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(raw, &sf)
}

func (sf *stateFile) save(path string, assets []Asset) error {
	if sf.Contents == nil {
		sf.Contents = map[string][]*File{}
	}
	for _, a := range assets {
		sf.Contents[assetToStateKey(a)] = a.Files()
	}

	raw, err := json.MarshalIndent(sf, "", "    ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, raw, 0644); err != nil {
		return err
	}
	return nil
}

func (sf *stateFile) fileFetcher(asset Asset) stateFileFetcher {
	return stateFileFetcher(sf.Contents[assetToStateKey(asset)])
}

type stateFileFetcher []*File

// FetchByName implements FileFetcher.
func (sff stateFileFetcher) FetchByName(name string) (*File, error) {
	files, err := sff.lookupByPattern(name)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, os.ErrNotExist
	}
	if len(files) > 1 {
		return nil, fmt.Errorf("more that one file found with name %q", name)
	}
	return files[0], nil
}

// FetchByPattern implements FileFetcher.
func (sff stateFileFetcher) FetchByPattern(pattern string) ([]*File, error) {
	return sff.lookupByPattern(pattern)
}

func (sff stateFileFetcher) lookupByPattern(pattern string) ([]*File, error) {
	var files []*File
	for idx, f := range sff {
		if ok, err := filepath.Match(pattern, f.Filename); err != nil && ok {
			files = append(files, sff[idx])
		}
	}
	return files, nil
}

func assetToStateKey(a Asset) string {
	return reflect.TypeOf(a).String()
}

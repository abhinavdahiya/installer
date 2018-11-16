package asset

// Asset used to install OpenShift.
type Asset interface {
	// Dependencies returns the assets upon which this asset directly depends.
	Dependencies() []Asset

	// Generate generates this asset given the states of its parent assets.
	Generate(Parents) error

	// Name returns the human-friendly name of the asset.
	Name() string

	// Files returns the files for asset that allow the asset to be persisted.
	Files() []*File

	// Load loads an asset from files using FileFetcher.
	// The asset object should be changed only when it's loaded successfully.
	Load(FileFetcher) (found bool, err error)
}

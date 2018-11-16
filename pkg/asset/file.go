package asset

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// File is a file for an Asset.
type File struct {
	// Filename is the name of the file.
	Filename string
	// Data is the contents of the file.
	Data []byte
}

// PersistToFile writes all of the files of the specified asset into the specified
// directory.
func PersistToFile(asset Asset, directory string) error {
	for _, f := range asset.Files() {
		path := filepath.Join(directory, f.Filename)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return errors.Wrap(err, "failed to create dir")
		}
		if err := ioutil.WriteFile(path, f.Data, 0644); err != nil {
			return errors.Wrap(err, "failed to write file")
		}
	}
	return nil
}

// deleteAssetFromDisk removes all the files for asset from disk.
// this is function is not safe for calling concurrently on the same directory.
func deleteAssetFromDisk(asset Asset, directory string) error {
	logrus.Debugf("Purging asset %q from disk", asset.Name())
	for _, f := range asset.Files() {
		path := filepath.Join(directory, f.Filename)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to remove file")
		}

		dir := filepath.Dir(path)
		ok, err := isDirEmpty(dir)
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to read directory")
		}
		if ok {
			if err := os.Remove(dir); err != nil {
				return errors.Wrap(err, "failed to remove directory")
			}
		}
	}
	return nil
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

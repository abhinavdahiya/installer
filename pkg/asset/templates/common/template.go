package common

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"

	"github.com/openshift/installer/pkg/asset"
)

// Template allows assets to generate a text/template.Template from .
type Template struct {
	file *asset.File
	tmpl *template.Template
}

// Generate loads the contents of the template from uri and compiles it to `text/template.Template`.
func (t *Template) Generate(uri, dst string) error {
	raw, err := getFileContents(uri)
	if err != nil {
		return errors.Wrapf(err, "failed to load uri %q", uri)
	}
	tmpl, err := template.New(filepath.Base(dst)).Parse(string(raw))
	if err != nil {
		return errors.Wrapf(err, "parse to template failed for uri %q", uri)
	}
	t.tmpl = tmpl
	t.file = &asset.File{
		Filename: dst,
		Data:     raw,
	}
	return nil
}

// Files implements `Asset.Files`.
func (t *Template) Files() []*asset.File {
	return []*asset.File{t.file}
}

// Load loads the template from dst using the fetcher.
// This implements `Asset.Load`.
func (t *Template) Load(dst string, fetcher asset.FileFetcher) (bool, error) {
	file, err := fetcher.FetchByName(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	t.file = file

	tmpl, err := template.New(filepath.Base(dst)).Parse(string(file.Data))
	if err != nil {
		return false, err
	}
	t.tmpl = tmpl
	return true, nil
}

package bootkube

import (
	"path/filepath"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/templates/common"
)

const (
	openshiftWebConsoleNamespaceFileName = "03-openshift-web-console-namespace.yaml"
)

var _ asset.Asset = (*OpenshiftWebConsoleNamespace)(nil)

// OpenshiftWebConsoleNamespace is the constant to represent contents of Openshift_WebConsoleNamespace.yaml file
type OpenshiftWebConsoleNamespace struct {
	common.Template
}

// Dependencies returns all of the dependencies directly needed by the asset
func (t *OpenshiftWebConsoleNamespace) Dependencies() []asset.Asset {
	return []asset.Asset{}
}

// Name returns the human-friendly name of the asset.
func (t *OpenshiftWebConsoleNamespace) Name() string {
	return "OpenshiftWebConsoleNamespace"
}

// Generate generates the actual files by this asset
func (t *OpenshiftWebConsoleNamespace) Generate(parents asset.Parents) error {
	return t.Template.Generate(
		filepath.Join(bootkubeDataDir, openshiftWebConsoleNamespaceFileName),
		filepath.Join(common.TemplateDir, openshiftWebConsoleNamespaceFileName),
	)
}

// Load returns the asset from disk.
func (t *OpenshiftWebConsoleNamespace) Load(f asset.FileFetcher) (bool, error) {
	return t.Template.Load(filepath.Join(common.TemplateDir, openshiftWebConsoleNamespaceFileName), f)
}

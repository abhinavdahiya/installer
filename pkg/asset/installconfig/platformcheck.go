package installconfig

import (
	"fmt"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/types/aws"
	"github.com/openshift/installer/pkg/types/azure"
	"github.com/openshift/installer/pkg/types/libvirt"
	"github.com/openshift/installer/pkg/types/none"
	"github.com/openshift/installer/pkg/types/openstack"
	"github.com/openshift/installer/pkg/types/vsphere"
)

// PlatformCheck is an asset that checks the platform configuration.
type PlatformCheck struct {
}

var _ asset.Asset = (*PlatformCheck)(nil)

// Dependencies returns the dependencies for PlatformCheck
func (a *PlatformCheck) Dependencies() []asset.Asset {
	return []asset.Asset{
		&InstallConfig{},
		&PlatformCredsCheck{},
	}
}

// Generate queries for input from the user.
func (a *PlatformCheck) Generate(dependencies asset.Parents) error {
	ic := &InstallConfig{}
	dependencies.Get(ic)

	var err error
	platform := ic.Config.Platform.Name()
	switch platform {
	case aws.Name, azure.Name, libvirt.Name, none.Name, vsphere.Name:
		// no platform checks.
	case openstack.Name:
	default:
		err = fmt.Errorf("unknown platform type %q", platform)
	}

	return err
}

// Name returns the human-friendly name of the asset.
func (a *PlatformCheck) Name() string {
	return "Platform Check"
}

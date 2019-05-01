package installconfig

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"

	"github.com/openshift/installer/pkg/asset"
	awsconfig "github.com/openshift/installer/pkg/asset/installconfig/aws"
	"github.com/openshift/installer/pkg/types/aws"
	awsvalidation "github.com/openshift/installer/pkg/types/aws/validation"
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

	platform := ic.Config.Platform.Name()
	switch platform {
	case azure.Name, libvirt.Name, none.Name, vsphere.Name:
		// no platform checks.
	case openstack.Name:
	case aws.Name:
		ssn, err := awsconfig.GetSession()
		if err != nil {
			return errors.Wrap(err, "creating AWS session")
		}
		return awsvalidation.ValidateInstallConfig(ic.Config, ec2.New(ssn)).ToAggregate()
	default:
		return fmt.Errorf("unknown platform type %q", platform)
	}
	return nil
}

// Name returns the human-friendly name of the asset.
func (a *PlatformCheck) Name() string {
	return "Platform Check"
}

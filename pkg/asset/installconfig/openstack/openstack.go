// Package openstack collects OpenStack-specific configuration.
package openstack

import (
	"github.com/pkg/errors"
	survey "gopkg.in/AlecAivazis/survey.v1"

	"github.com/openshift/installer/pkg/asset/userprovided"
	"github.com/openshift/installer/pkg/types/openstack"
)

const (
	defaultVPCCIDR = "10.0.0.0/16"
)

// Platform collects OpenStack-specific configuration.
func Platform() (*openstack.Platform, error) {
	region, err := userprovided.Generate(
		"OpenStack Region",
		&survey.Question{
			Prompt: &survey.Input{
				Message: "Region",
				Help:    "The OpenStack region to be used for installation.",
				Default: "regionOne",
			},
			Validate: survey.ComposeValidators(survey.Required, func(ans interface{}) error {
				//value := ans.(string)
				//FIXME(shardy) add some validation here
				return nil
			}),
		},
		"OPENSHIFT_INSTALL_OPENSTACK_REGION",
	)
	if err != nil {
		return nil, err
	}

	image, err := userprovided.Generate(
		"OpenStack Image",
		&survey.Question{
			Prompt: &survey.Input{
				Message: "Image",
				Help:    "The OpenStack image to be used for installation.",
				Default: "rhcos",
			},
			Validate: survey.ComposeValidators(survey.Required, func(ans interface{}) error {
				//value := ans.(string)
				//FIXME(shardy) add some validation here
				return nil
			}),
		},
		"OPENSHIFT_INSTALL_OPENSTACK_IMAGE",
	)
	if err != nil {
		return nil, err
	}

	cloud, err := userprovided.Generate(
		"OpenStack Cloud",
		&survey.Question{
			//TODO(russellb) - We could open clouds.yaml here and read the list of defined clouds
			//and then use survey.Select to let the user choose one.
			Prompt: &survey.Input{
				Message: "Cloud",
				Help:    "The OpenStack cloud name from clouds.yaml.",
			},
			Validate: survey.ComposeValidators(survey.Required, func(ans interface{}) error {
				//value := ans.(string)
				//FIXME(russellb) add some validation here
				return nil
			}),
		},
		"OPENSHIFT_INSTALL_OPENSTACK_CLOUD",
	)
	if err != nil {
		return nil, err
	}

	extNet, err := userprovided.Generate(
		"OpenStack External Network",
		&survey.Question{
			Prompt: &survey.Input{
				Message: "ExternalNetwork",
				Help:    "The OpenStack external network to be used for installation.",
			},
			Validate: survey.ComposeValidators(survey.Required, func(ans interface{}) error {
				//value := ans.(string)
				//FIXME(shadower) add some validation here
				return nil
			}),
		},
		"OPENSHIFT_INSTALL_OPENSTACK_EXTERNAL_NETWORK",
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Marshal %s platform", openstack.Name)
	}

	return &openstack.Platform{
		NetworkCIDRBlock: defaultVPCCIDR,
		Region:           region,
		BaseImage:        image,
		Cloud:            cloud,
		ExternalNetwork:  extNet,
	}, nil
}

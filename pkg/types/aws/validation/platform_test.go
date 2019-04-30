package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/openshift/installer/pkg/types/aws"
)

func TestValidatePlatform(t *testing.T) {
	cases := []struct {
		name     string
		platform *aws.Platform
		valid    bool
	}{
		{
			name: "minimal",
			platform: &aws.Platform{
				Region: "us-east-1",
			},
			valid: true,
		},
		{
			name: "invalid region",
			platform: &aws.Platform{
				Region: "bad-region",
			},
			valid: false,
		},
		{
			name: "valid machine pool",
			platform: &aws.Platform{
				Region:                 "us-east-1",
				DefaultMachinePlatform: &aws.MachinePool{},
			},
			valid: true,
		},
		{
			name: "invalid machine pool",
			platform: &aws.Platform{
				Region: "us-east-1",
				DefaultMachinePlatform: &aws.MachinePool{
					EC2RootVolume: aws.EC2RootVolume{
						IOPS: -10,
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid vpc, public subnet not provided",
			platform: &aws.Platform{
				Region:         "us-east-1",
				VPC:            "vpc-custom",
				PrivateSubnets: []string{"subnet-private-1", "subnet-private-2"},
			},
			valid: false,
		},
		{
			name: "invalid vpc, private subnet not provided",
			platform: &aws.Platform{
				Region:        "us-east-1",
				VPC:           "vpc-custom",
				PublicSubnets: []string{"subnet-public-1", "subnet-public-2"},
			},
			valid: false,
		},
		{
			name: "invalid public subnet",
			platform: &aws.Platform{
				Region:        "us-east-1",
				PublicSubnets: []string{"subnet-public-1", "subnet-public-2"},
			},
			valid: false,
		},
		{
			name: "invalid private subnet",
			platform: &aws.Platform{
				Region:         "us-east-1",
				PrivateSubnets: []string{"subnet-private-1", "subnet-private-2"},
			},
			valid: false,
		},
		{
			name: "unequal private and public subnet",
			platform: &aws.Platform{
				Region:         "us-east-1",
				VPC:            "vpc-custom",
				PrivateSubnets: []string{"subnet-private-1", "subnet-private-2"},
				PublicSubnets:  []string{"subnet-public-1"},
			},
			valid: false,
		},
		{
			name: "valid byo net",
			platform: &aws.Platform{
				Region:         "us-east-1",
				VPC:            "vpc-custom",
				PublicSubnets:  []string{"subnet-public-1", "subnet-public-2"},
				PrivateSubnets: []string{"subnet-private-1", "subnet-private-2"},
			},
			valid: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePlatform(tc.platform, field.NewPath("test-path")).ToAggregate()
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

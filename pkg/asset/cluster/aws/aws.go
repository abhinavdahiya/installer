// Package aws extracts AWS metadata from install configurations.
package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	awsconfig "github.com/openshift/installer/pkg/asset/installconfig/aws"
	"github.com/openshift/installer/pkg/types"
	awstypes "github.com/openshift/installer/pkg/types/aws"
)

// Metadata converts an install configuration to AWS metadata.
func Metadata(clusterID, infraID string, config *types.InstallConfig) *awstypes.Metadata {
	return &awstypes.Metadata{
		Region: config.Platform.AWS.Region,
		Identifier: []map[string]string{{
			fmt.Sprintf("kubernetes.io/cluster/%s", infraID): "owned",
		}, {
			"openshiftClusterID": clusterID,
		}},
	}
}

// TagSubnets adds tag key=value to subnets.
func TagSubnets(subnets []string, key, value string) error {
	ssn, err := awsconfig.GetSession()
	if err != nil {
		return err
	}
	req := &ec2.CreateTagsInput{
		Resources: aws.StringSlice(subnets),
		Tags: []*ec2.Tag{{
			Key:   aws.String(key),
			Value: aws.String(value),
		}},
	}
	_, err = ec2.New(ssn).CreateTags(req)
	if err != nil {
		return err
	}
	return nil
}

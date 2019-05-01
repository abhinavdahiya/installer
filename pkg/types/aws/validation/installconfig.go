package validation

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/openshift/installer/pkg/types"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const privateELBTag = "kubernetes.io/role/internal-elb"

// ValidateInstallConfig validates the config with respective to aws requirements.
func ValidateInstallConfig(config *types.InstallConfig, client *ec2.EC2) field.ErrorList {
	allErrs := field.ErrorList{}
	p := config.Platform.AWS
	pfield := field.NewPath("platform").Child("aws")
	if len(p.VPC) > 0 {
		if err := validateVPC(p.VPC, client); err != nil {
			allErrs = append(allErrs, field.Invalid(pfield.Child("vpc"), p.VPC, err.Error()))
		}

		if err := validatePrivateSubnets(p.PrivateSubnets, client); err != nil {
			allErrs = append(allErrs, field.Invalid(pfield.Child("privateSubnets"), p.PrivateSubnets, err.Error()))
		}

		if err := validatePublicSubnetsAZs(p.PublicSubnets, p.PrivateSubnets, client); err != nil {
			allErrs = append(allErrs, field.Invalid(pfield.Child("publicSubnets"), p.PublicSubnets, err.Error()))
		}
	}
	return allErrs
}

// validateVPC ensure that the vpc (id)
// Has enableDnsSupport and enableDnsHostnames attributes enabled
func validateVPC(id string, client *ec2.EC2) error {
	att, err := awsVpcDescribeVpcAttribute(id, "enableDnsSupport", client)
	if err != nil {
		return errors.Wrapf(err, "failed to get vpc attribute enableDnsSupport for %s", id)
	}
	if !aws.BoolValue(att.EnableDnsSupport.Value) {
		return errors.New("enableDnsSupport must be enabled")
	}

	att, err = awsVpcDescribeVpcAttribute(id, "enableDnsHostnames", client)
	if err != nil {
		return errors.Wrapf(err, "failed to get vpc attribute enableDnsHostnames for %s", id)
	}
	if !aws.BoolValue(att.EnableDnsHostnames.Value) {
		return errors.New("enableDnsHostnames must be enabled")
	}

	return nil
}

// validatePublicSubnetsAZs ensures that public subnets cover at least the availability zones
// covered by private subnets.
func validatePublicSubnetsAZs(public []string, private []string, client *ec2.EC2) error {
	privateSet := sets.NewString(private...)
	publicSet := sets.NewString(public...)
	req := &ec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice(append(public, private...)),
	}
	resp, err := client.DescribeSubnets(req)
	if err != nil {
		return err
	}
	privateAzs := sets.NewString()
	publicAzs := sets.NewString()
	for _, subnet := range resp.Subnets {
		if privateSet.Has(aws.StringValue(subnet.SubnetId)) {
			privateAzs.Insert(aws.StringValue(subnet.AvailabilityZone))
		}
		if publicSet.Has(aws.StringValue(subnet.SubnetId)) {
			publicAzs.Insert(aws.StringValue(subnet.AvailabilityZone))
		}
	}
	if diff := privateAzs.Difference(publicAzs); diff.Len() > 0 {
		return fmt.Errorf("No public subnets were provided in zones %s", diff.List())
	}
	return nil
}

// validatePrivateSubnets ensure that all the private subnets have the tag
// `kubernetes.io/role/internal-elb`.
func validatePrivateSubnets(ids []string, client *ec2.EC2) error {
	req := &ec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice(ids),
	}
	resp, err := client.DescribeSubnets(req)
	if err != nil {
		return err
	}
	var invalidSubnets []string
	for _, subnet := range resp.Subnets {
		found := false
		for _, t := range subnet.Tags {
			if k := aws.StringValue(t.Key); k == privateELBTag {
				found = true
				break
			}
		}
		if !found {
			invalidSubnets = append(invalidSubnets, aws.StringValue(subnet.SubnetId))
		}
	}
	if len(invalidSubnets) > 0 {
		sort.Strings(invalidSubnets)
		return fmt.Errorf("missing %q tag from private subnets %s", privateELBTag, invalidSubnets)
	}
	return nil
}

func awsVpcDescribeVpcAttribute(vpcID string, att string, client *ec2.EC2) (*ec2.DescribeVpcAttributeOutput, error) {
	describeAttrOpts := &ec2.DescribeVpcAttributeInput{
		Attribute: aws.String(att),
		VpcId:     aws.String(vpcID),
	}
	resp, err := client.DescribeVpcAttribute(describeAttrOpts)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

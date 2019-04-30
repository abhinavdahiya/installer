package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsutil "github.com/openshift/installer/pkg/asset/installconfig/aws"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

// AvailabilityZones retrieves a list of availability zones for the given region
//  and subnets.
// If no subnets are provided, it returns all available AZs for the region.
func AvailabilityZones(region string, subnets ...string) ([]string, error) {
	ec2Client, err := ec2Client(region)
	if err != nil {
		return nil, err
	}

	if len(subnets) > 0 {
		zones, err := fetchAvailabilityZonesForSubnets(ec2Client, region, subnets)
		if err != nil {
			return nil, errors.Wrapf(err, "failed fetch for (%s, %v)", region, subnets)
		}
		return zones, nil
	}
	zones, err := fetchAvailabilityZones(ec2Client, region)
	if err != nil {
		return nil, errors.Wrapf(err, "failed fetch for %s", region)
	}
	return zones, nil
}

func ec2Client(region string) (*ec2.EC2, error) {
	ssn, err := awsutil.GetSession()
	if err != nil {
		return nil, err
	}

	client := ec2.New(ssn, aws.NewConfig().WithRegion(region))
	return client, nil
}

func fetchAvailabilityZones(client *ec2.EC2, region string) ([]string, error) {
	req := &ec2.DescribeAvailabilityZonesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("region-name"),
				Values: []*string{aws.String(region)},
			},
			{
				Name:   aws.String("state"),
				Values: []*string{aws.String("available")},
			},
		},
	}
	resp, err := client.DescribeAvailabilityZones(req)
	if err != nil {
		return nil, err
	}
	zones := []string{}
	for _, zone := range resp.AvailabilityZones {
		zones = append(zones, *zone.ZoneName)
	}
	return zones, nil
}

func fetchAvailabilityZonesForSubnets(client *ec2.EC2, region string, subnets []string) ([]string, error) {
	sIDs := make([]*string, len(subnets))
	for i := range subnets {
		sIDs[i] = aws.String(subnets[i])
	}
	resp, err := client.DescribeSubnets(&ec2.DescribeSubnetsInput{SubnetIds: sIDs})
	if err != nil {
		return nil, err
	}
	zones := []string{}
	for _, subnet := range resp.Subnets {
		zones = append(zones, aws.StringValue(subnet.AvailabilityZone))
	}
	return sets.NewString(zones...).List(), nil
}

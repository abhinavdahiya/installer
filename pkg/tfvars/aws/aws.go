// Package aws contains AWS-specific Terraform-variable logic.
package aws

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/installer/pkg/types/aws/defaults"
	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api-provider-aws/pkg/apis/awsproviderconfig/v1beta1"
)

type config struct {
	AMI                     string            `json:"aws_ami"`
	BootstrapInstanceType   string            `json:"aws_bootstrap_instance_type"`
	ExtraTags               map[string]string `json:"aws_extra_tags,omitempty"`
	IOPS                    int64             `json:"aws_master_root_volume_iops"`
	MasterAvailabilityZones []string          `json:"aws_master_availability_zones"`
	MasterInstanceType      string            `json:"aws_master_instance_type"`
	PrivateSubnets          []string          `json:"aws_private_subnets,omitempty"`
	PublicSubnets           []string          `json:"aws_public_subnets,omitempty"`
	Region                  string            `json:"aws_region"`
	Size                    int64             `json:"aws_master_root_volume_size"`
	Type                    string            `json:"aws_master_root_volume_type"`
	VPCID                   string            `json:"aws_vpc_id,omitempty"`
	WorkerAvailabilityZones []string          `json:"aws_worker_availability_zones"`
}

// TFVars generates AWS-specific Terraform variables launching the cluster.
// vpcID, publicSubnets, privateSubnets can be empty.
func TFVars(vpcID string, publicSubnets []string, privateSubnets []string,
	masterConfigs []*v1beta1.AWSMachineProviderConfig, workerConfigs []*v1beta1.AWSMachineProviderConfig) ([]byte, error) {
	masterConfig := masterConfigs[0]

	tags := make(map[string]string, len(masterConfig.Tags))
	for _, tag := range masterConfig.Tags {
		tags[tag.Name] = tag.Value
	}

	masterAvailabilityZones := make([]string, len(masterConfigs))
	for i, c := range masterConfigs {
		masterAvailabilityZones[i] = c.Placement.AvailabilityZone
	}

	exists := struct{}{}
	availabilityZoneMap := map[string]struct{}{}
	for _, c := range workerConfigs {
		availabilityZoneMap[c.Placement.AvailabilityZone] = exists
	}
	workerAvailabilityZones := make([]string, 0, len(availabilityZoneMap))
	for zone := range availabilityZoneMap {
		workerAvailabilityZones = append(workerAvailabilityZones, zone)
	}

	if len(masterConfig.BlockDevices) == 0 {
		return nil, errors.New("block device slice cannot be empty")
	}

	rootVolume := masterConfig.BlockDevices[0]
	if rootVolume.EBS == nil {
		return nil, errors.New("EBS information must be configured for the root volume")
	}

	if rootVolume.EBS.VolumeType == nil {
		return nil, errors.New("EBS volume type must be configured for the root volume")
	}

	if rootVolume.EBS.VolumeSize == nil {
		return nil, errors.New("EBS volume size must be configured for the root volume")
	}

	if *rootVolume.EBS.VolumeType == "io1" && rootVolume.EBS.Iops == nil {
		return nil, errors.New("EBS IOPS must be configured for the io1 root volume")
	}

	instanceClass := defaults.InstanceClass(masterConfig.Placement.Region)

	cfg := &config{
		AMI: *masterConfig.AMI.ID,
		BootstrapInstanceType:   fmt.Sprintf("%s.large", instanceClass),
		ExtraTags:               tags,
		MasterAvailabilityZones: masterAvailabilityZones,
		MasterInstanceType:      masterConfig.InstanceType,
		PrivateSubnets:          privateSubnets,
		PublicSubnets:           publicSubnets,
		Region:                  masterConfig.Placement.Region,
		Size:                    *rootVolume.EBS.VolumeSize,
		Type:                    *rootVolume.EBS.VolumeType,
		VPCID:                   vpcID,
		WorkerAvailabilityZones: workerAvailabilityZones,
	}

	if rootVolume.EBS.Iops != nil {
		cfg.IOPS = *rootVolume.EBS.Iops
	}

	return json.MarshalIndent(cfg, "", "  ")
}

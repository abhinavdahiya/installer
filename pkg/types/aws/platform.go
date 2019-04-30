package aws

// Platform stores all the global configuration that all machinesets
// use.
type Platform struct {
	// Region specifies the AWS region where the cluster will be created.
	Region string `json:"region"`

	// VPC specifies the VPC where the cluster will be created.
	// If not set, a new VPC will be created for the cluster.
	// +optional
	VPC string `json:"vpc,omitempty"`

	// PublicSubnets specifies the list of subnets where the public resources for the cluster will be created.
	// This list of subnets must be non-zero only when VPC field is set.
	// https://aws.amazon.com/premiumsupport/knowledge-center/public-load-balancer-private-ec2/
	// +optional
	PublicSubnets []string `json:"publicSubnets,omitempty"`

	// PrivateSubnets specifies the list of subnets where the private resources for the cluster will be created.
	// This list of subnets must be non-zero only when VPC field is set.
	// +optional
	PrivateSubnets []string `json:"privateSubnets,omitempty"`

	// UserTags specifies additional tags for AWS resources created for the cluster.
	// +optional
	UserTags map[string]string `json:"userTags,omitempty"`

	// DefaultMachinePlatform is the default configuration used when
	// installing on AWS for machine pools which do not define their own
	// platform configuration.
	// +optional
	DefaultMachinePlatform *MachinePool `json:"defaultMachinePlatform,omitempty"`
}

# Canonical internal state definitions for this module.
# read only: only locals and data source definitions allowed. No resources or module blocks in this file

// Only reference data sources which are guaranteed to exist at any time (above) in this locals{} block
locals {
  // How many AZs to create subnets in
  new_az_count = "${length(var.availability_zones)}"

  // When referencing the _ids arrays or data source arrays via count = , always use the *_count variable rather than taking the length of the list
  private_subnet_ids = ["${split(",", var.vpc_id == "" ? join(",", aws_subnet.private_subnet.*.id) :  join(",", data.aws_subnet.cluster_private.*.id))}"]

  private_subnet_azs = ["${distinct(split(",", var.vpc_id == "" ? join(",", aws_subnet.private_subnet.*.availability_zone) :  join(",", data.aws_subnet.cluster_private.*.availability_zone)))}"]
  public_subnet_ids  = ["${split(",", var.vpc_id == "" ? join(",", aws_subnet.public_subnet.*.id) :  join(",", data.aws_subnet.cluster_public.*.id))}"]

  public_subnet_azs = ["${distinct(split(",", var.vpc_id == "" ? join(",", aws_subnet.public_subnet.*.availability_zone) :  join(",", data.aws_subnet.cluster_public.*.availability_zone)))}"]
}

# all data sources should be input variable-agnostic and used as canonical source for querying "state of resources" and building outputs
# (ie: we don't want "aws.new_vpc" and "data.aws_vpc.cluster_vpc", just "data.aws_vpc.cluster_vpc" used everwhere).

data "aws_vpc" "cluster_vpc" {
  # The join() hack is required because currently the ternary operator
  # evaluates the expressions on both branches of the condition before
  # returning a value. When providing and external VPC, the template VPC
  # resource gets a count of zero which triggers an evaluation error.
  #
  # This is tracked upstream: https://github.com/hashicorp/hil/issues/50
  #
  id = "${var.vpc_id == "" ? join(" ", aws_vpc.new_vpc.*.id) : var.vpc_id }"
}

data "aws_subnet" "cluster_public" {
  count = "${var.vpc_id == "" ? 0 : length(var.public_subnets)}"
  id    = "${var.public_subnets[count.index]}"
}

data "aws_subnet" "cluster_private" {
  count = "${var.vpc_id == "" ? 0 : length(var.private_subnets)}"
  id    = "${var.private_subnets[count.index]}"
}

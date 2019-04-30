variable "vpc_id" {
  type    = "string"
  default = ""

  description = <<EOF
  The VPC for the cluster.
  If empty, new VPC will be created for the cluster.
  EOF
}

variable "public_subnets" {
  type    = "list"
  default = []

  description = <<EOF
  The public subnets for the cluster.
  If empty, new public subnets will be created in the availability_zones for the cluster.
  EOF
}

variable "private_subnets" {
  type    = "list"
  default = []

  description = <<EOF
  The private subnets for the cluster.
  If empty, new public subnets will be created in the availability_zones for the cluster.
  EOF
}

variable "availability_zones" {
  type    = "list"
  default = []

  description = <<EOF
  The availability zones in which to provision subnets.
  This is only used when public_subnets and private_subnets are empty to create new subnets for the cluster.
  EOF
}

variable "cidr_block" {
  type = "string"
}

variable "cluster_id" {
  type = "string"
}

variable "region" {
  type        = "string"
  description = "The target AWS region for the cluster."
}

variable "tags" {
  type        = "map"
  default     = {}
  description = "AWS tags to be applied to created resources."
}

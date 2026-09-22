variable "region" {
  description = "OCI region"
  type        = string
  default     = "us-sanjose-1"
}

variable "tenancy_ocid" {
  description = "Tenancy OCID (also used as root compartment)"
  type        = string
}

variable "compartment_ocid" {
  description = "Compartment to create resources in (usually same as tenancy for root)"
  type        = string
}

variable "ssh_public_key_path" {
  description = "Path to the SSH public key used for instance access"
  type        = string
  default     = "~/.ssh/id_ed25519_oracle.pub"
}

variable "instance_name" {
  description = "Compute instance name"
  type        = string
  default     = "ebpf-observer"
}

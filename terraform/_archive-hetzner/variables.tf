variable "hcloud_token" {
  description = "Hetzner Cloud API token (Read & Write)"
  type        = string
  sensitive   = true
}

variable "ssh_public_key_path" {
  description = "Path to SSH public key uploaded to the server"
  type        = string
  default     = "~/.ssh/id_ed25519_oracle.pub"
}

variable "server_name" {
  description = "Cloud server name"
  type        = string
  default     = "ebpf-observer"
}

variable "server_type" {
  description = "Hetzner server type. CAX11 = 2 vCPU / 4 GB arm64 (€3.79/mo)"
  type        = string
  default     = "cax11"
}

variable "location" {
  description = "Hetzner datacenter location. ash = Ashburn VA, closest US"
  type        = string
  default     = "fsn1"
}

variable "image" {
  description = "Base OS image"
  type        = string
  default     = "ubuntu-24.04"
}

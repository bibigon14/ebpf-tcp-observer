output "instance_public_ip" {
  description = "Public IPv4 of the eBPF observer instance"
  value       = oci_core_instance.observer.public_ip
}

output "instance_private_ip" {
  description = "Private IPv4 of the eBPF observer instance (inside VCN)"
  value       = oci_core_instance.observer.private_ip
}

output "ssh_command" {
  description = "Ready-to-paste SSH command"
  value       = "ssh -i ~/.ssh/id_ed25519_oracle ubuntu@${oci_core_instance.observer.public_ip}"
}

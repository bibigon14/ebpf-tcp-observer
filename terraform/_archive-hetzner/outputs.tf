output "server_public_ipv4" {
  description = "Public IPv4 of the observer server"
  value       = hcloud_server.observer.ipv4_address
}

output "server_public_ipv6" {
  description = "Public IPv6 of the observer server"
  value       = hcloud_server.observer.ipv6_address
}

output "ssh_command" {
  description = "Ready-to-paste SSH command"
  value       = "ssh -i ~/.ssh/id_ed25519_oracle root@${hcloud_server.observer.ipv4_address}"
}

# ebpf-tcp-observer

An eBPF-based agent that observes TCP-level events (retransmits, connect latency)
on a Linux host and exports them as Prometheus metrics.

Runs as a single Go binary using CO-RE (Compile Once, Run Everywhere) via
`cilium/ebpf`, no BCC or kernel headers at runtime.

Deployed as a systemd service on a public-facing Oracle Cloud VM, with metrics
streamed to a homelab Prometheus over WireGuard for correlation with existing
service SLOs.

## Status

Work in progress. See [docs/roadmap.md](docs/roadmap.md).

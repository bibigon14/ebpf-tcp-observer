# Roadmap

## v0.1 - Proof of concept

Goal: end-to-end pipeline from kernel event to Grafana panel.

- [ ] Provision Oracle Cloud Ampere A1 arm64 instance, Ubuntu 24.04
- [ ] WireGuard tunnel Oracle ↔ homelab Pi (bidirectional)
- [ ] Homelab Prometheus scrape config for the Oracle instance
- [ ] Minimal Go binary using cilium/ebpf that loads a single kprobe on
      `tcp_retransmit_skb` and increments a counter per event
- [ ] `/metrics` HTTP endpoint exposing `tcp_retransmits_total{saddr,daddr}`
- [ ] systemd unit, `CAP_BPF` + `CAP_PERFMON`, no root
- [ ] Grafana panel: top talkers by retransmit rate
- [ ] README with architecture diagram and one-command reproduction

## v0.2 - Connect latency

- [ ] Second kprobe pair: `tcp_v4_connect` (start) + `tcp_rcv_established` (finish)
- [ ] BPF map holding start timestamp keyed by socket pointer
- [ ] Histogram metric `tcp_connect_duration_seconds` with process labels
      (comm, pid) resolved from userspace
- [ ] Grafana panel: p50/p95/p99 connect latency by process

## v0.3 - Correlation with victim SLOs

Tie observed network events back to the existing homelab SLO framework:

- [ ] Add a burn-generator script on the Oracle instance that periodically
      degrades network conditions (packet loss via `tc netem`) against
      one of the homelab services
- [ ] Show the resulting cascade in Grafana: raw kernel events
      (`tcp_retransmits_total`) → blackbox probe SLO burn
      (`slo:probe:success_rate_5m`) → chaos-scheduler-operator aborting
      its next run because the SLO guardrail tripped
- [ ] Write it up as a case study - the value of kernel-level observability
      is proving root cause instead of guessing at it

## Non-goals (for now)

- Kubernetes pod-awareness (cgroup → pod resolution). Interesting, but
  it's a second project's worth of scope; keep this one focused on
  host-level network observation.
- eBPF-based enforcement (dropping packets, blocking connections).
  This is an observer, not an enforcer.
- Multi-node deployment. Single edge node is enough to demonstrate the
  pattern; scaling is a productization concern.

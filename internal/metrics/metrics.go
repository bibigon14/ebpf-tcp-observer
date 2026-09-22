// Package metrics owns the Prometheus registry and counter/gauge definitions
// used across the observer. Kept separate from collectors so that a stub
// collector (v0) and the eBPF collector (v1+) can populate the same
// well-known metric without touching the HTTP handler.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Registry is a dedicated registry, so we do not leak Go runtime metrics
// unless the operator explicitly wires them in. Keeps /metrics output tight.
var Registry = prometheus.NewRegistry()

var factory = promauto.With(Registry)

// TCPRetransmits counts kernel-level TCP retransmit events observed by the
// agent. In v0 this is incremented by a synthetic ticker; from v1 it is
// backed by a kprobe on tcp_retransmit_skb.
var TCPRetransmits = factory.NewCounter(prometheus.CounterOpts{
	Name: "ebpf_tcp_retransmits_total",
	Help: "Total TCP retransmit events observed by the eBPF agent.",
})

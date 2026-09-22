// Package collector produces metric samples. v0 ships a stub that increments
// a counter on a ticker; v1 will replace this with a real kprobe attached to
// tcp_retransmit_skb via cilium/ebpf.
package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/bibigon14/ebpf-tcp-observer/internal/metrics"
)

// RunStub increments the retransmit counter every interval until ctx is done.
// It exists so the transport path (WireGuard, scrape, dashboard) can be
// validated end-to-end before the eBPF probe lands.
func RunStub(ctx context.Context, interval time.Duration, log *slog.Logger) {
	log.Info("stub collector starting", "interval", interval)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info("stub collector stopping")
			return
		case <-t.C:
			metrics.TCPRetransmits.Inc()
		}
	}
}

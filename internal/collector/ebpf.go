// eBPF collector: loads the compiled kprobe program into the kernel,
// attaches it to tcp_retransmit_skb, then polls the shared array map
// on a ticker and mirrors the kernel counter into the Prometheus
// counter exported by internal/metrics.
//
// Two counters are involved deliberately:
//   - The BPF array holds the *absolute* number of retransmit events
//     since program load. It is a u64 owned by the kernel.
//   - metrics.TCPRetransmits is a Prometheus counter, which only
//     accepts .Add(delta) with a non-negative delta.
// The collector remembers the last observed absolute value and adds
// the difference. This handles the map being monotonic and the
// prometheus type expecting deltas.
//
// The program is intentionally minimal for v0.1: single global counter,
// no per-{src,dst} labels. Labels land in v0.2 via an LPM trie map keyed
// on the socket's remote address.

package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"

	"github.com/bibigon14/ebpf-tcp-observer/internal/metrics"
)

// RunEBPF loads the eBPF objects, attaches the kprobe, and blocks until
// ctx is done. All eBPF resources are released on return.
func RunEBPF(ctx context.Context, pollInterval time.Duration, log *slog.Logger) error {
	// Older kernels require RLIMIT_MEMLOCK to be raised before eBPF maps
	// can be created. Kernel 5.11+ removed the requirement, but leaving
	// this in keeps the collector portable across the kernels a public
	// edge host might run.
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("remove memlock: %w", err)
	}

	// Load the ELF that bpf2go embedded into the binary and materialise
	// its maps and programs in the kernel. objs.Close() releases both.
	var objs tcpRetransmitObjects
	if err := loadTcpRetransmitObjects(&objs, nil); err != nil {
		return fmt.Errorf("load bpf objects: %w", err)
	}
	defer objs.Close()

	// Attach the program to tcp_retransmit_skb. cilium/ebpf resolves the
	// symbol via /proc/kallsyms; if the kernel does not export it (built
	// without tracing, CONFIG_KPROBES=n, or the function was inlined),
	// this returns an error that surfaces up.
	kp, err := link.Kprobe("tcp_retransmit_skb", objs.KprobeTcpRetransmitSkb, nil)
	if err != nil {
		return fmt.Errorf("attach kprobe tcp_retransmit_skb: %w", err)
	}
	defer kp.Close()

	log.Info("ebpf kprobe attached", "symbol", "tcp_retransmit_skb", "poll_interval", pollInterval)

	// The map holds one u64 at key=0. See tcp_retransmit.c for the
	// rationale behind BPF_MAP_TYPE_ARRAY over a hash for a single
	// global counter.
	var (
		key      uint32
		absolute uint64
		last     uint64
	)

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("ebpf collector stopping")
			return nil
		case <-t.C:
			if err := objs.RetransmitCount.Lookup(&key, &absolute); err != nil {
				if errors.Is(err, ebpf.ErrKeyNotExist) {
					// key=0 is created at map init; missing means the map
					// was replaced under us. Log and keep going.
					log.Warn("retransmit_count key missing")
					continue
				}
				log.Error("map lookup failed", "err", err)
				continue
			}

			if absolute < last {
				// Only reason this happens in practice is program reload
				// resetting the map. Do not emit a negative delta.
				log.Warn("counter went backwards, resetting baseline",
					"last", last, "absolute", absolute)
				last = absolute
				continue
			}

			delta := absolute - last
			if delta > 0 {
				metrics.TCPRetransmits.Add(float64(delta))
				last = absolute
			}
		}
	}
}

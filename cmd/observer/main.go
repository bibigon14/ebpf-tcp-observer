// Command observer exposes eBPF-derived TCP metrics as a Prometheus endpoint.
//
// By default it attaches a kprobe to tcp_retransmit_skb and mirrors the
// kernel counter into a Prometheus counter. Pass -collector=stub to fall
// back to a synthetic ticker; useful for debugging the transport path
// when eBPF is unavailable (missing CAP_BPF, kprobe not exported, etc).
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bibigon14/ebpf-tcp-observer/internal/collector"
	"github.com/bibigon14/ebpf-tcp-observer/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	addr := flag.String("listen", "10.0.0.4:9100", "metrics listen address (host:port)")
	mode := flag.String("collector", "ebpf", "collector to use: ebpf | stub")
	pollStr := flag.String("poll-interval", "1s", "how often to sample the eBPF map (ebpf mode)")
	tickStr := flag.String("stub-interval", "10s", "stub collector tick interval (stub mode)")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start the chosen collector. Failure in the eBPF path is fatal: we do
	// not silently degrade to stub, because that would hide a real
	// observability outage behind synthetic data.
	switch *mode {
	case "ebpf":
		poll, err := time.ParseDuration(*pollStr)
		if err != nil {
			log.Error("invalid poll-interval", "value", *pollStr, "err", err)
			os.Exit(2)
		}
		go func() {
			if err := collector.RunEBPF(ctx, poll, log); err != nil {
				log.Error("ebpf collector failed", "err", err)
				cancel()
			}
		}()

	case "stub":
		tick, err := time.ParseDuration(*tickStr)
		if err != nil {
			log.Error("invalid stub-interval", "value", *tickStr, "err", err)
			os.Exit(2)
		}
		go collector.RunStub(ctx, tick, log)

	default:
		log.Error("unknown collector mode", "mode", *mode, "valid", "ebpf|stub")
		os.Exit(2)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("http server listening", "addr", *addr, "collector", *mode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server crashed", "err", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown failed", "err", err)
	}
	log.Info("bye")
}

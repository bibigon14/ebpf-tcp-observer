// Command observer exposes eBPF-derived TCP metrics as a Prometheus endpoint.
// v0 wires a stub collector so the pipeline can be validated before the eBPF
// probes are implemented.
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
	tickStr := flag.String("stub-interval", "10s", "stub collector tick interval")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	tick, err := time.ParseDuration(*tickStr)
	if err != nil {
		log.Error("invalid stub-interval", "value", *tickStr, "err", err)
		os.Exit(2)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go collector.RunStub(ctx, tick, log)

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
		log.Info("http server listening", "addr", *addr)
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

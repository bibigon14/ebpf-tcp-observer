// Package collector generates eBPF Go bindings via cilium/ebpf's bpf2go.
//
// Run `go generate ./internal/collector/` to compile bpf/tcp_retransmit.c
// with clang and emit ARM64/x86_64 Go source files carrying the compiled
// object as a byte slice. The generated files are checked in so consumers
// of this repo do not need clang installed to `go build`.
package collector

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -target native -type u64 tcpRetransmit ../../bpf/tcp_retransmit.c -- -I../../bpf/headers

# 1. Use cilium/ebpf, not BCC

Date: 2026-09-20
Status: Accepted

## Context

Two mainstream toolchains exist for writing eBPF programs that ship as
loadable Go binaries: **BCC** (BPF Compiler Collection, Python/C
bindings, requires clang + kernel headers on the target) and
**cilium/ebpf** (pure Go library, uses BTF and CO-RE for
Compile-Once-Run-Everywhere semantics).

The target for this project is a single-binary systemd service that
lands on a fresh Ubuntu 24.04 host and starts working. No package
install choreography, no kernel header matching per host, no runtime
compile step. Portability matters because the same binary should
eventually run on the arm64 Oracle Cloud instance and on any x86_64
demo host during interviews.

## Decision

Use `cilium/ebpf` with CO-RE.

## Consequences

**Positive**

- Single static Go binary. No clang, no headers, no BCC runtime deps.
- CO-RE handles per-kernel struct layout differences via BTF, so the
  same binary works across kernels that expose BTF (Ubuntu 20.04+,
  most modern distros).
- Native Go for the userspace half - existing tooling (structured
  logs, Prometheus client, testing) works without shims.
- Cross-compile between amd64 and arm64 is trivial; the BPF object
  is architecture-neutral.

**Negative**

- CO-RE requires `/sys/kernel/btf/vmlinux` on the target kernel. Not
  all distributions ship it (notably Raspberry Pi OS does not; that's
  why the compute for this project sits on Ubuntu Cloud instead of
  the homelab Pi).
- Less mature examples than BCC - most public eBPF tutorials are
  BCC-first. Reading them means translating C-with-Python-glue into
  cilium/ebpf's Go API.
- Requires a working toolchain to generate BPF objects (clang for the
  BPF-C source, then `go:generate` to bind them). Not a runtime cost,
  but a build-time one.

## Alternatives considered

- **BCC.** Rejected for the runtime dependency footprint - defeats
  the "single binary systemd service" goal.
- **libbpf-go directly.** More minimal than cilium/ebpf but the
  ergonomics for maps, perf buffers, and program loading are worse;
  cilium/ebpf wraps libbpf-go concepts in idiomatic Go.
- **Tetragon or Pixie.** Full observability platforms, not libraries.
  Right answer for a real deployment, wrong answer for a portfolio
  project that wants to show the primitives.

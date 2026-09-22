// SPDX-License-Identifier: GPL-2.0
//
// tcp_retransmit.c - counts kernel-level TCP retransmit events.
//
// Attaches a kprobe to tcp_retransmit_skb(), the kernel function invoked
// when the TCP stack decides to retransmit a segment (RTO fired, dup-ACK
// threshold reached, tail loss probe, etc). Every invocation increments
// the single counter held in `retransmit_count`.
//
// The map is an array of one u64, which is the simplest structure the
// verifier accepts for a global counter. Reads from userspace never
// contend with the kernel writer because __sync_fetch_and_add is atomic.
//
// Built for CO-RE (Compile Once, Run Everywhere) - only vmlinux.h symbols
// are used, no kernel-version-specific offsets are hardcoded.

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

char LICENSE[] SEC("license") = "GPL";

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 1);
} retransmit_count SEC(".maps");

SEC("kprobe/tcp_retransmit_skb")
int BPF_KPROBE(kprobe_tcp_retransmit_skb, struct sock *sk)
{
    __u32 key = 0;
    __u64 *value = bpf_map_lookup_elem(&retransmit_count, &key);
    if (value)
        __sync_fetch_and_add(value, 1);
    return 0;
}

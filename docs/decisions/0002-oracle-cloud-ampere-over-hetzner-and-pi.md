# 0002 - Oracle Cloud Ampere A1 over Hetzner Cloud and Raspberry Pi

Date: 2026-09-22
Status: Accepted

## Context

The eBPF agent needs a Linux host that is:

1. arm64 (matches Ampere in production edge deployments; cross-arch story via
   CO-RE remains, but a real arm64 target is more compelling for the portfolio).
2. Running a kernel with BTF (/sys/kernel/btf/vmlinux) exposed, so CO-RE can
   work without shipping distro-specific headers.
3. Always-on with a stable public IP so the homelab Prometheus can scrape it
   over a WireGuard tunnel.
4. Cheap enough that leaving it running as a portfolio artifact does not turn
   into a recurring monthly cost.

Three candidate hosts were evaluated: Raspberry Pi in the existing homelab,
a Hetzner Cloud CAX arm64 VM, and an Oracle Cloud Ampere A1 flex VM.

## Options considered

### Option A - Raspberry Pi in the homelab

Pros: zero incremental cost, already provisioned, familiar operating model,
the eBPF agent could observe real traffic of the bots and services already
running on the Pi.

Cons ruled it out: the Pi runs Raspberry Pi OS with kernel 6.18.29+rpt-rpi-2712,
which ships without /sys/kernel/btf/vmlinux and without a /boot/config-*
file. CO-RE requires BTF, so the only paths forward were rebuilding the
kernel with CONFIG_DEBUG_INFO_BTF or migrating the Pi to Ubuntu Server 24.04
(which does ship BTF). Both are invasive changes to a homelab that is
already running production workloads (bots, k3s, observability stack).
Not worth the disruption for what is a portfolio project.

### Option B - Hetzner Cloud CAX11 arm64

Pros: EUR 3.79/mo for 2 vCPU / 4 GB arm64 (Ampere Altra), Ubuntu 24.04
with BTF out of the box, clean provider API, fast Terraform provider.

Cons ruled it out at provisioning time: CAX arm64 shapes are EU-only
(Falkenstein, Nuremberg, Helsinki). No US region for arm64 as of
September 2026. That is livable - the agent talks to homelab over
WireGuard, so latency is not critical. But the Hetzner Cloud API also
returned "unsupported location for server type" for every EU location
attempted, and querying /v1/datacenters showed all EU DCs reporting
zero orderable server types at that moment. Whether this was capacity
throttling for new accounts or a wider outage, the UX confirmed it:
CAX11 was flagged "Limited availability" in the console. Not a base
we can rely on for a portfolio artifact that needs to keep running.

### Option C - Oracle Cloud Ampere A1

Pros: Ampere Altra arm64, Ubuntu 24.04 Minimal with BTF, and a
generous Always Free allowance (1500 OCPU-hours + 9000 GB-hours per
month post-June-2026 reduction) that covers a 2 OCPU / 12 GB flex
instance running 24/7 with room to spare.

Cons: the Always Free arm64 lottery is real. First attempt on a fresh
free-tier account failed with "Out of host capacity" in us-sanjose-1
after ~90 minutes of UI wrangling (the Create Instance wizard's public
IPv4 toggle silently would not enable, VCN Wizard button was hidden,
and switching to Terraform did not help capacity). No path to change
home region or subscribe to another region on a brand-new tenancy.

## Decision

Oracle Cloud Ampere A1, on a Pay As You Go account.

The PAYG upgrade unblocked capacity: paid tenants sit ahead of free-tier
lottery for arm64 allocation, and the Always Free allowance is preserved
after the upgrade. Provisioning succeeded in the first apply after the
upgrade (35 seconds), same region, same shape, same Terraform code.

Effective monthly cost: ~$0.30 for boot volume storage (~47 GB). The 2 OCPU
+ 12 GB compute stays within Always Free allowance. Oracle's $300 credit
for 30 days covers even the storage in the first month.

## Consequences

Positive:

- Real arm64 target host with BTF, matching what the portfolio narrative
  says about edge deployments on Ampere.
- Effectively $0/mo running cost - won't turn into recurring maintenance
  or a subscription to cancel.
- Terraform code is portable across free-tier and PAYG accounts; only
  region choice depends on tenancy home region.
- The Pi stays untouched, its k3s and bots keep running unchanged.

Negative:

- Oracle Cloud has a steeper UI learning curve than Hetzner. Console
  navigation for VCN/subnet/security list is less obvious.
- PAYG requires a credit card and a $100 authorization hold at signup.
  For a portfolio project this is acceptable overhead but is worth
  naming.
- The Ampere A1 shape family may be deprecated in the future (Oracle
  has been shifting toward A2 in some regions). If that happens, the
  Terraform code has one shape name to change.

Reverted alternatives (Pi, Hetzner) are documented in terraform/_archive-*/
so the failure paths can be reviewed if the situation changes (Pi
migrates to Ubuntu Server, or Hetzner capacity opens up).

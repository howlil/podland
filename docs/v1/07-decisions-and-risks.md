# 07 — ADR Summary, Constraints & Open Questions

## Accepted architectural decisions

| ADR | Decision | Why |
|---|---|---|
| ADR-001 | Single-server modular monolith Go | Small footprint, manageable failure domains |
| ADR-002 | React + TypeScript + Bun build; static embedded UI | No JS server required in production |
| ADR-003 | SQLite WAL + `database/sql`, raw SQL | In-process metadata, no DB server overhead |
| ADR-004 | Podman as **only** V1 runtime | Lower potential idle footprint; preserve runtime port for future |
| ADR-005 | Caddy separate process | Dynamic routing persists when control plane restarts |
| ADR-006 | Cloudflare Tunnel default ingress | Works behind NAT/private datacenter; wildcard public hostnames |
| ADR-007 | GitHub App source + Dockerfile-first auto-deploy | No PAT in normal UX; push-to-deploy |
| ADR-008 | Readiness-gated promotion + durable worker | Failed releases should not replace healthy ones |
| ADR-009 | Repository boundaries but only SQLite adapter V1 | Make PostgreSQL migration feasible without needless work now |

## Minimum supported VM (engineering test target)

- **2 vCPU / 2 GB RAM / 40 GB SSD**, 20 Mbps connection and 512 GB monthly egress (network assumptions from target VM offering).
- Linux x86-64; first validated distribution must be pinned during installer work (e.g., Ubuntu LTS).
- Start with **1 concurrent build max**, configurable container CPU/memory/disk/log limits; admit/reject based on headroom.
- Target control-plane idle **≤256 MiB** (Go+Caddy, include cloudflared when selected); **unverified** and must be measured as PSS/RSS plus host memory.
- OS, Podman/storage, Buildah/build cache, user containers, and OS page cache are *additional* consumers.
- State metrics/retention policy must have hard caps; clean stopped containers, unused image/build cache carefully—never persistent volumes.

## Unresolved decisions (must validate before dependent milestone)

| ID | Question | Validation gate |
|---|---|---|
| OPEN-01 | Cloudflare connection UX: supported OAuth/delegation vs scoped API token / tunnel credential? | Provider API capability + least-privilege permissions spike before M4 |
| OPEN-02 | Podman rootless vs rootful on service host (network, systemd, volumes, reboot)? | Test start/restart, Caddy networking and permission model before M1 final |
| OPEN-03 | Podman build path: Podman API vs Buildah CLI, image pinning/logs? | M1 spike on VM minimum |
| OPEN-04 | Pure Go SQLite driver choice and CGO implications? | Benchmark startup/RAM and migration/test behavior during M0 |
| OPEN-05 | Bun native React dev/build sufficient vs Vite fallback? | React routing, HMR, production bundle, tests during M0 |
| OPEN-06 | Caddy ↔ Cloudflare origin, hostname forwarding, TLS mode, wildcard certificate needs? | End-to-end smoke before M4 release |
| OPEN-07 | How to handle shared writable volumes during overlap? | M2 design: controlled downtime or explicit rejection |

## Architectural risks

- **Docker/Podman socket compromise:** isolate Podland control-plane, permission-lock Unix sockets; treat it as host-level privileged control; no hostile multi-tenancy V1.
- **OOM/build exhaustion:** resource admission; one build; test that heavy app cannot make dashboard unusable under chosen cgroup/systemd limits.
- **Disk exhaustion:** logs and image caches grow; set retention, disk watermark, explicit recovery flows.
- **SQLite recovery:** WAL on local filesystem; online backup with checkpoint-safe API; include encryption key in disaster recovery procedure.
- **Leaked credentials:** GitHub App narrow permissions; webhook HMAC validation; encrypted env secrets; never print tokens in logs.
- **Caddy route/metadata divergence:** declarative desired routing + startup reconciliation; route changes atomic where possible.
- **PostgreSQL future:** repository ports reduce coupling, but schema/SQL/transactions and data migration still require work.

## Change control

Accepted ADR changed → update this doc, impacted docs, API and tests in same change. Pending decisions are **not** features already supported. Avoid benchmark claims without measured test context and reproducible workload.

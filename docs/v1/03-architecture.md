# 03 — Technical Architecture & Module Boundaries

## Chosen stack

| Boundary | Choice |
|---|---|
| UI | React + TypeScript; Tailwind CSS; native React state + `fetch`; Bun native build/dev (uji kompatibilitas) |
| Control plane | Go, `net/http`, modular monolith, internal background worker |
| Metadata | SQLite WAL, Go `database/sql`, raw SQL, versioned migrations, **single local filesystem** |
| App runtime | Podman (satu adapter V1), systemd/service management sesuai hasil spike |
| Image build | Podman/Buildah Dockerfile-first, satu job build pada satu waktu secara default |
| Routing | Caddy sebagai proses terpisah, private admin API, no access to Podman socket |
| External ingress | Cloudflare Tunnel primary; wildcard hostnames; public HTTPS at Cloudflare |
| Protocol | REST+OpenAPI; Server-Sent Events untuk logs/progress |
| Observability | Go `slog`, Podman logs/inspect/stats; no Prometheus server V1 |
| Distribution | Go binary meng-embed static frontend, installer/service unit; Bun tidak berjalan production |

## Request topology

```text
Browser ─HTTPS─> Cloudflare Edge ─tunnel─> cloudflared
                                          │
                                          ▼
                                        Caddy
                                       /     \
                             Podland Go      User app containers
                            (API + UI)         (Podman)
                               │                   ▲
                         SQLite WAL                │
                         Job Worker ──runtime adapter─┘
```

Podland backend dan Caddy harus tetap *independently restartable*. Caddy route tetap melayani aplikasi meski Go API restart. Tunnel mode menggunakan TLS publik di Cloudflare, dan koneksi internal dilindungi oleh jaringan/akses origin yang tepat.

## Module layout (target, dibuat bertahap)

```text
apps/
  api/
    cmd/podland/main.go
    internal/
      modules/
        identity/        projects/       applications/
        deployments/     integrations/   networking/
        services/        observability/
      infrastructure/
        database/sqlite/
        runtime/podman/
        proxy/caddy/
        providers/cloudflare/
        providers/github/
      platform/config/
      httpapi/
    migrations/
    go.mod
  web/src/
    features/   components/   routes/   lib/api/
contracts/openapi.yaml
Taskfile.yml
AGENTS.md
```

**Module responsibilities:**

- **Identity:** bootstrap/login/session, owner-only authorization.
- **Projects/Applications:** app metadata, repo/config, desired state.
- **Integrations:** GitHub App, Cloudflare credentials; never container lifecycle.
- **Deployments:** validate → build → create/start → verify → activate → rollback.
- **Networking:** domain model dan route intents; Caddy/Cloudflare details di adapter.
- **Services:** managed DB/container templates dan persistence.
- **Observability:** expose status/logs/host metrics tanpa heavy backend.

## Dependency rule

`HTTP handler → module service → module-owned ports → infrastructure adapters`

- Business modules tidak mengimpor Podman SDK, Caddy client, SQLite-specific queries atau Cloudflare API.
- Repository interface dibuat **per kebutuhan consumer**, tidak satu generic CRUD repository.
- SQL dan transaction boundary tetap jelas; jangan pura-pura SQL SQLite = PostgreSQL.
- Tidak membuat `docker/`, `postgres/`, `kubernetes/`, `proxmox/` adapters sebelum ada use case nyata.
- Integration wiring hanya di composition root (`cmd/podland`).

Contoh kontrak *minimal*, final signature mengikuti use case:

```go
type Runtime interface {
    Create(ctx context.Context, spec ContainerSpec) (ContainerID, error)
    Start(ctx context.Context, id ContainerID) error
    Inspect(ctx context.Context, id ContainerID) (ContainerState, error)
    Stop(ctx context.Context, id ContainerID) error
    Remove(ctx context.Context, id ContainerID) error
}
```

Jangan menganggap runtime swap hanya mengganti adapter: Proxmox dan multi-node juga membutuhkan scheduler, networking dan storage orchestration baru.

## SQLite → PostgreSQL migration boundary

Repositori SQLite di V1. Bila pindah PostgreSQL: buat adapter/schema/migrations baru, uji transaksi/locking dan concurrency, lalu export-transform-import-validate data. Trigger perpindahan: multiple control-plane writers, kebutuhan HA, atau measured write contention; **bukan** semata jumlah app.

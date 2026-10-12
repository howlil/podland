# Podland V1 — Documentation

> Status: **Draft v1.0** · Updated: 2026-10-10 · Product: lightweight self-hosted PaaS · Audience: solo developer / homelab

**Misi:** deploy dan kelola aplikasi pada satu Linux VM melalui dashboard, dengan onboarding singkat, GitHub auto-deploy, domain otomatis, dan biaya resource control plane yang rendah.

## Baca berurutan

| Dokumen | Fokus |
|---|---|
| [01 — Product & UX](docs/01-product-and-ux.md) | Target user, MVP scope, onboarding singkat |
| [02 — Requirements](docs/02-requirements.md) | FR, rules, acceptance criteria |
| [03 — Architecture](docs/03-architecture.md) | Stack, modul, interfaces, runtime topology |
| [04 — Data & API](docs/04-data-and-api.md) | Entities, persistence, REST contract |
| [05 — Deployment lifecycle](docs/05-deployment-lifecycle.md) | Durable job, state machine, rollback |
| [06 — Delivery plan](docs/06-delivery-plan.md) | M0–M5, TDD, vertical slices |
| [07 — Decisions & risks](docs/07-decisions-and-risks.md) | ADR, hardware constraints, open questions |

**Hierarchy:** Product constraints → User flows → API/Domain contract → Infrastructure adapters → Implementation/tests.

## Keputusan terkunci

- **V1 single-server**, Podman sebagai satu-satunya container runtime yang diimplementasikan.
- **Go modular monolith** + `net/http`, SQLite/WAL, React + TypeScript + Bun (build-time), Caddy, Cloudflare Tunnel sebagai ingress default.
- **GitHub-first**: OAuth untuk owner login; GitHub App untuk repository + webhook; Dockerfile-first build.
- **Automatic deployment** on push ke branch terpilih, **health-gated promotion**, mempertahankan versi lama hingga versi baru siap jika dapat dilakukan dengan aman.
- **Wildcard subdomain** otomatis saat Cloudflare routing terhubung.
- **2 vCPU / 2 GB RAM / 40 GB disk** sebagai VM minimum yang diuji, bukan jaminan untuk workload berat.

**Belum final:** mekanisme otorisasi Cloudflare, mode Podman rootless/rootful, dan angka penggunaan RAM aktual. Tercatat di [07 — Decisions & risks](docs/07-decisions-and-risks.md).

## Aturan pengembangan

1. Satu feature = **contract → BE/FE paralel → integration → E2E**.
2. Red → Green → Refactor. Jangan menambah adapter masa depan sebelum diperlukan.
3. Setiap slice harus berjalan dari UI sampai runtime; hindari backend-only backlog panjang.
4. Update docs/ADR bila keputusan berubah. API canonical nantinya berada di `contracts/openapi.yaml`.

Dokumentasi ini adalah **rancangan**, bukan klaim fitur sudah diimplementasikan.

# 06 — Implementation Plan (TDD + Vertical Slices)

## Development cadence

Setiap fitur: **small spec/acceptance → OpenAPI → BE dan FE paralel → integrate → E2E → merge**. Tidak ada fase membangun seluruh backend terlebih dahulu. Gunakan Red → Green → Refactor.

| Milestone | Output | Exit condition |
|---|---|---|
| M0 — Foundation | Go API, React/Bun UI shell, SQLite migration, Taskfile, CI/test, OpenAPI seed | UI ↔ `/healthz` berjalan; production binary embed assets; tests pass |
| M1 — Deployment Core | Podman adapter, create app, deploy OCI image, durable queue, status/logs | Docker-compatible OCI image deployed via UI; restart and recovery tested |
| M2 — Routing & Health | Caddy adapter, HTTP readiness, route promotion, rollback proof | Failed release tidak memutus release lama bila resource cukup |
| M3 — GitHub Source | GitHub App, Dockerfile build, branch webhook, auto deployment | Push configured branch memicu build dan health-gated deployment |
| M4 — Installer & Onboarding | VM installer, owner bootstrap, GitHub OAuth, Cloudflare setup, wildcard domains | Fresh VM install → connect → deploy via browser |
| M5 — V1 Hardening | Secrets/env, persistent volumes, managed DB templates, metrics, failure/disk/security tests | Semua V1 release gate di 02 requirements lulus |

**Milestone M1 memakai deploy OCI image sebagai technical vertical slice**, tetapi UX produk V1 final adalah GitHub/Dockerfile-first; tidak ada kontradiksi karena urutan engineering ≠ urutan wizard.

## M0 task breakdown

- **M0-T01:** minimal monorepo `apps/api`, `apps/web`, `contracts`, `docs`; pin Go/Bun versions, lint/format baseline.
- **M0-T02:** Go `net/http` server, config, `/healthz`, graceful shutdown; tests dahulu.
- **M0-T03:** Bun+React shell, API fetch, health indicator; UI/component test.
- **M0-T04:** SQLite WAL opening, migrations, repository testing setup; persistence/restart tests.
- **M0-T05:** OpenAPI 3.1 seed contract, error schema, TypeScript API types approach, Taskfile.
- **M0-T06:** Build static UI, `go:embed` into binary, CI, fresh-run docs and E2E smoke.

## First coding task (M0-T01)

1. Inspect existing repo first (`git status`, files, Go/Bun availability), **do not overwrite existing work**.
2. Create only directories needed for Go API + Bun UI bootstrap. Empty future modules not required.
3. Add `README.md` and Taskfile commands (`dev:api`, `dev:web`, `test`, `build`) as each command becomes usable.
4. Definition of done: clean bootstrap from fresh clone, reproducible commands, no Node/Bun process required in production architecture.

## CI gates

Go tests, Bun/TypeScript checks, OpenAPI validation, Go build, frontend build, integration tests on runner equipped with Podman (once M1 starts), Playwright smoke (once shell is ready). Separate rootless/rootful platform tests after runtime decision.

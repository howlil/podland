# 02 — V1 Requirements & Acceptance Criteria

## Functional requirements

| ID | Requirement | Acceptance check |
|---|---|---|
| FR-01 | Owner bootstrap dan GitHub OAuth login | Hanya owner terikat pertama dapat mengakses dashboard; bootstrap token single-use/expiring |
| FR-02 | Cloudflare connection + wildcard zone | Domain otomatis dapat mencapai Caddy via tunnel setelah setup valid |
| FR-03 | GitHub App repository connection | User memilih repo/branch tanpa menempel PAT; permission minimum |
| FR-04 | Project & app CRUD | App selalu milik tepat satu project; config tersimpan setelah restart |
| FR-05 | GitHub Dockerfile build | Push ke configured branch menghasilkan job durable, image, log build |
| FR-06 | Docker image deploy (secondary path) | Pull pinned reference, create container, cek readiness |
| FR-07 | HTTP readiness check | Path configurable; timeout/retry terukur; failure tidak mengaktifkan route baru |
| FR-08 | Domain routing | Generated hostname unik, route menunjuk deployment aktif saja |
| FR-09 | Environment & secrets | Update diterapkan pada deployment berikut; secrets tidak direturn plaintext |
| FR-10 | Logs/status/history | User bisa melacak run, image revision, error dan runtime status |
| FR-11 | Actions & rollback | Restart/redeploy/rollback memunculkan operation record; idempotent |
| FR-12 | Persistent volumes | Restart/redeploy tidak menghapus data; delete data perlu konfirmasi eksplisit |
| FR-13 | Managed service | PostgreSQL, MySQL, Redis bisa di-deploy secara internal dengan volume bila relevan |
| FR-14 | Basic metrics | Host CPU, RAM, disk, container usage tanpa metrics database khusus |
| FR-15 | Recovery | Backend restart/reboot memulihkan job/state dan tidak menghentikan app sehat |

## Business rules

1. V1 = **single trusted owner + single VM**; apps user bukan tenant hostile.
2. Hanya satu operasi mutasi deployment aktif per app. Requests duplicate didedup via operation/idempotency key.
3. Setiap deployment mengacu immutable `revision` + immutable image digest, tidak mutable `latest` sebagai identitas rollback.
4. **Ready ≠ running**. Routing baru dialihkan hanya setelah readiness pass.
5. Deployment gagal mempertahankan route sebelumnya bila release lama masih layak dan resource memungkinkan.
6. Domain/hostname unik; env var keys unik per app; referensi secrets tidak bocor ke logs/API.
7. Volume lifecycle independen; garbage collector tidak boleh menghapus data persistent tanpa approval eksplisit.
8. Webhook GitHub harus signature-validated, idempotent, hanya branch yang dipilih.
9. Jika Cloudflare belum siap, dashboard tetap bisa diakses via bootstrap/local channel; deploy publik memberi status konfigurasi yang jelas.
10. Resource admission untuk build/containers mencegah overload jika memungkinkan; ditolak dengan error yang dapat ditindaklanjuti.

## Non-functional requirements

- **Target host minimum:** 2 vCPU, 2 GB RAM, 40 GB SSD; koneksi 20 Mbps, kuota egress 512 GB/bulan (asumsi paket VM yang ditargetkan).
- **Lightweight:** target awal API+Caddy(+cloudflared sesuai mode) idle ≤256 MiB, *hipotesis, bukan SLA*. Total Docker/Podman, kernel dan OS diukur terpisah.
- **Secure by default:** session cookie Secure/HttpOnly/SameSite, CSRF protection, socket private, least-privilege integration tokens, encrypted secrets at rest.
- **Reliable:** durable job queue; reconciliation saat backend start; request deployment asynchronous.
- **Observable:** error berkode, structured logging, retention/rotation dibatasi.
- **Testable:** contract/API tests, module tests, Podman+Caddy integration tests, E2E first deployment.

## Release gate

- Fresh-install di VM 2GB sukses; onboarding jelas dan tanpa password app account.
- Build GitHub Dockerfile, auto-deploy push, domain HTTPS, readiness promotion sukses.
- Simulasi failed build, failed readiness, backend crash, VM reboot, rollback sukses.
- Volume dan konfigurasi tetap ada setelah redeploy/restart.
- RAM/disk/CPU diukur dengan workload representatif, hasil dicatat; jika target idle meleset, ada keputusan eksplisit.
- Ancaman rootless/runtime privilege, webhook replay, secrets leak, dan disk exhaustion memiliki tests/control.

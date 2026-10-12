# 01 — Product Definition & User Experience

## Target dan positioning

- **Primary:** solo developer, indie hacker, homelab owner yang mempunyai Linux VM kecil.
- **Secondary:** small development team yang berbagi VM trusted, **tanpa** isolasi multi-tenant hostile.
- **Problem:** deploy Dockerfile-based app butuh SSH, container tooling, DNS, HTTPS, logs, dan upgrade manual.
- **Promise V1:** hubungkan server + Cloudflare + GitHub sekali, lalu deploy dan kelola aplikasi dengan dashboard.

## Minimum setup flow

```text
Install Podland on VM
  → Buka bootstrap link di browser
  → Klaim owner (login GitHub)
  → Hubungkan Cloudflare zone + tunnel
  → Hubungkan GitHub App
  → Dashboard ready
```

UI setup ditargetkan **maksimal tiga tahap utama** (claim, Cloudflare, GitHub); langkah instalasi dilakukan di terminal. Jika provider izin membutuhkan langkah ekstra, jelaskan tepat di layar, jangan sembunyikan kompleksitas.

### First deployment

```text
Select GitHub repository
  → pilih branch + Dockerfile (default auto-detect)
  → pilih internal HTTP port + readiness path (default `/`)
  → Deploy
  → Build → Start → Check readiness → Switch route
  → app-<short-id>.apps.example.com (generated, unique)
```

Auto domain tidak boleh bergantung pada user mengisi DNS satu per satu. Domain assignment harus stabil setelah app dibuat; gunakan suffix ID untuk mencegah collision. `apps.example.com` hanya contoh, bukan domain yang disediakan Podland.

### Regular lifecycle

- Push ke branch yang dikonfigurasi memicu deployment baru; hanya versi sehat yang dipromosikan.
- User dapat melihat current status, logs, deployment history, restart, redeploy, rollback.
- Environment variables dan persistent volumes terpisah dari lifecycle container.

## Scope V1

**Core:** bootstrap & auth owner, project/app CRUD, GitHub App source, Dockerfile build, auto-deploy branch, optional Docker image source, health-gated deployment, automatic wildcard route, Cloudflare-first ingress, env/secrets, volumes, logs/history, rollback, basic host metrics, managed PostgreSQL/MySQL/Redis sebagai service single-node.

**Non-goals:** Proxmox, multi-node orchestration, Kubernetes, HA control plane, enterprise RBAC/SSO, multi-tenancy billing, autoscale, marketplace, arbitrary public TCP/UDP ingress, managed backup/PITR.

## Success criteria

Pengguna dengan VM sesuai minimum bisa install Podland, login, menghubungkan Cloudflare dan GitHub, deploy satu app ringan yang mempunyai HTTP readiness, memperoleh domain HTTPS, dan mengelola deploy failure tanpa SSH manual setelah onboarding.

**Not guaranteed:** zero-downtime universal, build proyek berat di RAM 2 GB, sandbox kuat untuk kode tidak tepercaya.

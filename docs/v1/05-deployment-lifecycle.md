# 05 — Deployment Lifecycle & Recovery

## State machine

```text
QUEUED → PREPARING → BUILDING* → CREATING → STARTING → VERIFYING → ACTIVATING → SUCCEEDED
          │             │           │            │            │             │
          └─────────────┴───────────┴────────────┴────────────┴─────────────┴→ FAILED
```

`BUILDING*` dilewati untuk Docker/OCI image source. `CANCELED` hanya bila job dapat dihentikan secara aman; bukan otomatis setiap state.

## Release algorithm

1. Validate resource budget, config, source, port, writable volume semantics; persist immutable revision.
2. Acquire per-app operation lock via durable SQLite transaction; enqueue operation.
3. Build via Podman/Buildah dari GitHub Dockerfile atau pull image; pin immutable digest.
4. Create new container with predictable Podland labels; attach isolated Podman network.
5. Start, inspect; perform **Podland-managed HTTP readiness** ke path configurable (default `/`) pada internal port. Gunakan total timeout + interval/retry limit.
6. Hanya setelah READY: update Caddy route secara atomik/transactional bila didukung, verify intended config; set active deployment.
7. Stop/cleanup old container **hanya setelah** switch sukses dan retention policy membolehkan; preserve known-good old image untuk rollback.

**Safety:** jika build/start/check gagal, route lama tetap aktif; operasi gagal dan detail error muncul di UI. Jika route switch gagal, rollback route intent. Jangan menghapus running old container lebih dulu. Sinkronisasi DB↔Caddy menggunakan reconciliation untuk kegagalan di antara dua operasi.

## Nuansa availability

Create-before-switch memerlukan resource cukup untuk **dua versi hidup sementara**. Pada 2 GB, jika RAM kurang atau volume bersifat shared-writable, Podland harus menggunakan deployment dengan downtime yang dijelaskan di UI atau menolak operasi jika tidak aman. **Tidak ada janji universal zero-downtime.**

## Durable worker

- SQLite menyimpan operation status dan lease/claim info; worker adalah goroutine dalam Go binary.
- Semua step harus idempotent atau mendeteksi artefak existing menggunakan labels dan IDs.
- Saat startup: scan jobs belum terminal, inspect Podman containers, bandingkan active deployment + Caddy routes, reconcile sebelum menerima kerja konflik.
- Simpan log build/deploy dengan bounded retention dan sanitized secrets.
- GitHub webhook signature verified + delivery dedup; hanya configured branch yang di-queue.

## Rollback

Rollback = buat **deployment baru** yang mengacu revision dan image digest sebelumnya. Jalankan readiness dan promotion sama seperti deploy normal; jangan mengedit record history lama.

## Minimal integration failure scenarios

- Invalid Dockerfile/build fails → old app remains available.
- HTTP readiness fails/timeouts → route not switched.
- Podland backend killed mid deploy → reconcile safely after restart.
- VM reboots → managed app restarts sesuai policy, dashboard recovers metadata.
- Caddy config rejected → previous routing preserved.
- Disk full / RAM shortage → readable error, no data destruction.
- Redeploy app with volume → persistent data still present.

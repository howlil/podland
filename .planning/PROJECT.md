# Podland

## What This Is

Podland is a multi-tenant PaaS (Platform as a Service) for students, built on a single bare-metal server managed by Proxmox. Users can deploy containerized applications ("VMs") with automatic resource allocation, domain setup via Cloudflare, and built-in observability. Authentication requires GitHub OAuth with @student.unand.ac.id email verification.

## Core Value

Students can deploy and run applications with zero DevOps knowledge — just create a "VM", get auto-configured domain and tunnel, and focus on code.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] User authentication via GitHub OAuth with @student.unand.ac.id verification
- [ ] Role-based access control (superadmin, internal/SI, external)
- [ ] User can create "VM" (container) with CPU/RAM/OS selection (Ubuntu/Debian)
- [ ] Automatic Cloudflare DNS and Tunnel setup for each VM
- [ ] VM runs on k3s with Docker containers
- [ ] Resource quotas per role (conservative: Internal 1 CPU/2GB RAM, External 0.5 CPU/1GB RAM)
- [ ] Grafana/Prometheus/Loki dashboard for VM metrics
- [ ] Auto-delete VM after 2 days of combined idle (no HTTP traffic + no active process + no user login)
- [ ] Terraform IoC for VM provisioning
- [ ] Monorepo structure with Turborepo
- [ ] Frontend: TanStack Start, Tailwind v4, Zustand, Axios
- [ ] Backend: Go services
- [ ] Helm package manager for k3s deployments
- [ ] Storage: local for OS/temp, local-lvm for persistent volumes

### Out of Scope

- Real VM (qemu/kvm) — using Docker containers with resource limits instead
- Dedicated resources — shared resource model only
- Multi-server cluster — single server deployment initially
- Mobile app — web dashboard only

## Context

**Infrastructure:**
- Single bare-metal server managed by Proxmox
- Private network access, exposed via Cloudflare Tunnel
- k3s cluster for container orchestration
- Storage: local + local-lvm

**User Model:**
- Target: ~500 student users
- Internal: Students from SI UNAND (identified by NIM prefix "1152")
- External: Students outside SI UNAND
- Superadmin: Platform administrators

**NIM Structure:**
- Format: `YY####` where YY = year, #### = department code
- SI department code: 1152
- Example: 221152XX = SI student, class of 2022

## Constraints

- **Tech Stack**: Go backend, React frontend (TanStack ecosystem) — user preference
- **Infrastructure**: Single server, Proxmox-managed, Cloudflare-dependent
- **Authentication**: GitHub OAuth only, @student.unand.ac.id required
- **Resource Model**: Shared resources with conservative quotas
- **Timeline**: Side project (no hard deadline)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| k3s over Docker native | "Cloud native" requirement, better multi-tenant isolation, mature observability ecosystem | — Pending |
| Container-as-VM abstraction | Shared resource model, 500 users target, simpler than real VMs | — Pending |
| Combined idle detection | Comprehensive idle detection (HTTP + process + login) prevents resource waste | — Pending |
| Conservative quotas | Limited server resources, 500 users shared pool | — Pending |

---
*Last updated: 2026-03-25 after initial questioning*

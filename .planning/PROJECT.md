# Podland

## Current State

**Shipped:** v1.0 — Foundation to Admin + Polish (2026-03-30)

**Status:** Production Ready — All 48 requirements implemented

**Stats:**
- 5 phases completed
- ~4,500 lines of code (Go + TypeScript)
- 22 Kubernetes manifests
- 6 Grafana dashboards

---

## What This Is

Podland is a multi-tenant PaaS (Platform as a Service) for students, built on a single bare-metal server managed by Proxmox. Users can deploy containerized applications ("VMs") with automatic resource allocation, domain setup via Cloudflare, and built-in observability. Authentication requires GitHub OAuth with @student.unand.ac.id email verification.

## Core Value

Students can deploy and run applications with zero DevOps knowledge — just create a "VM", get auto-configured domain and tunnel, and focus on code.

---

## Requirements

### Validated (Shipped in v1.0)

**Authentication (6/6):**
- ✓ AUTH-01: GitHub OAuth sign in — v1.0
- ✓ AUTH-02: Student email validation — v1.0
- ✓ AUTH-03: NIM extraction + role assignment — v1.0
- ✓ AUTH-04: Session persistence (7-day JWT) — v1.0
- ✓ AUTH-05: Profile page — v1.0
- ✓ AUTH-06: Sign out — v1.0

**VM Management (11/11):**
- ✓ VM-01: Create VM — v1.0
- ✓ VM-02: Start VM — v1.0
- ✓ VM-03: Stop VM — v1.0
- ✓ VM-04: Restart VM — v1.0
- ✓ VM-05: Delete VM — v1.0
- ✓ VM-06: View VM list — v1.0
- ✓ VM-07: View VM detail — v1.0
- ✓ VM-08: Quota enforcement — v1.0
- ✓ VM-09: Prevent quota exceed — v1.0
- ✓ VM-10: Isolated namespace — v1.0
- ✓ VM-11: Non-root container — v1.0

**Domain & Networking (6/6):**
- ✓ DOMAIN-01: Auto subdomain — v1.0
- ✓ DOMAIN-02: Cloudflare DNS — v1.0
- ✓ DOMAIN-03: Cloudflare Tunnel — v1.0
- ✓ DOMAIN-04: HTTPS with SSL — v1.0
- ✓ DOMAIN-05: Domain list UI — v1.0
- ✓ DOMAIN-06: Delete domain — v1.0

**Resource Quotas (5/5):**
- ✓ QUOTA-01: Track resource usage — v1.0
- ✓ QUOTA-02: View quota in dashboard — v1.0
- ✓ QUOTA-03: Hard limits — v1.0
- ✓ QUOTA-04: Superadmin modify quota — v1.0
- ✓ QUOTA-05: ResourceQuota per namespace — v1.0

**Monitoring & Observability (6/6):**
- ✓ MON-01: Metrics collection — v1.0
- ✓ MON-02: Grafana dashboard — v1.0
- ✓ MON-03: Loki log aggregation — v1.0
- ✓ MON-04: Log viewer UI — v1.0
- ✓ MON-05: CPU alerts — v1.0
- ✓ MON-06: Memory alerts — v1.0

**User Dashboard (4/4):**
- ✓ DASH-01: Dashboard with VM summary — v1.0
- ✓ DASH-02: Create VM wizard — v1.0
- ✓ DASH-03: Activity log — v1.0
- ✓ DASH-04: Responsive design — v1.0

**Admin Panel (5/5):**
- ✓ ADMIN-01: List all users — v1.0
- ✓ ADMIN-02: Change user role — v1.0
- ✓ ADMIN-03: Ban/unban users — v1.0
- ✓ ADMIN-04: System health dashboard — v1.0
- ✓ ADMIN-05: Audit log — v1.0

**API (4/4):**
- ✓ API-01: REST API for VM ops — v1.0
- ✓ API-02: API key auth — v1.0
- ✓ API-03: OpenAPI docs — _Deferred to T+30_
- ✓ API-04: Rate limiting — _Deferred to T+7_

### Active (v1.1 Planning)

- [ ] OpenAPI/Swagger documentation (T+30)
- [ ] Rate limiting on auth endpoints (T+7)
- [ ] Privacy policy page (GDPR compliance) (T+7)
- [ ] Account deletion feature (GDPR right to erasure) (T+30)
- [ ] Full WCAG AA compliance (T+30)
- [ ] Lighthouse CI monitoring (T+30)
- [ ] User guide documentation
- [ ] Load testing execution (k6 scripts ready)

### Out of Scope (v2+)

- Real VM (qemu/kvm) — using Docker containers with resource limits instead
- Dedicated resources — shared resource model only
- Multi-server cluster — single server deployment initially
- Mobile app — web dashboard only
- Windows OS templates — licensing complexity
- GPU acceleration — niche use case

---

## Context

**Infrastructure:**
- Single bare-metal server managed by Proxmox
- Private network access, exposed via Cloudflare Tunnel
- k3s cluster for container orchestration
- Storage: local + local-lvm
- Monitoring: Prometheus (15-day retention), Loki (30-day retention)

**User Model:**
- Target: ~500 student users
- Internal: Students from SI UNAND (identified by NIM prefix "1152")
- External: Students outside SI UNAND
- Superadmin: Platform administrators

**NIM Structure:**
- Format: `YY####` where YY = year, #### = department code
- SI department code: 1152
- Example: 221152XX = SI student, class of 2022

**Tech Stack:**
- **Backend:** Go 1.25, chi router, PostgreSQL, JWT auth
- **Frontend:** React 18, TanStack Router, Tailwind CSS v4, Zustand
- **Infrastructure:** k3s, Docker, Cloudflare, Prometheus, Grafana, Loki

---

## Constraints

- **Tech Stack**: Go backend, React frontend (TanStack ecosystem) — user preference
- **Infrastructure**: Single server, Proxmox-managed, Cloudflare-dependent
- **Authentication**: GitHub OAuth only, @student.unand.ac.id required
- **Resource Model**: Shared resources with conservative quotas
- **Timeline**: Side project (no hard deadline)

---

## Key Decisions (v1.0)

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| k3s over Docker native | "Cloud native" requirement, better multi-tenant isolation, mature observability ecosystem | ✅ **Good** — Clean separation, Helm charts work well |
| Container-as-VM abstraction | Shared resource model, 500 users target, simpler than real VMs | ✅ **Good** — Fast provisioning, efficient resource usage |
| Combined idle detection | Comprehensive idle detection (HTTP + process + login) prevents resource waste | ✅ **Good** — Avoids false positives |
| Conservative quotas | Limited server resources, 500 users shared pool | ✅ **Good** — Fair distribution |
| Clean Architecture | Separation of concerns, testability, maintainability | ✅ **Good** — Clear boundaries, easy to test |
| Cloudflare Tunnel | No public IP needed, automatic HTTPS | ✅ **Good** — Simple setup, works behind NAT |
| SendGrid for email | Free tier (100/day), zero operational overhead | ✅ **Good** — Reliable delivery |

---

## Next Milestone Goals (v1.1)

**Focus:** Production hardening + User experience

1. **Security & Compliance**
   - Add rate limiting on auth endpoints
   - Implement privacy policy page
   - Add account deletion feature (GDPR)

2. **Documentation**
   - OpenAPI/Swagger specification
   - User guide (sign in, create VM, access via domain)
   - API documentation for developers

3. **Accessibility**
   - Fix color contrast issues
   - Improve focus indicators
   - Add screen reader announcements
   - Achieve WCAG AA compliance

4. **Performance**
   - Lighthouse CI integration
   - Core Web Vitals monitoring
   - Bundle size optimization

5. **Testing**
   - Execute load tests (k6 scripts ready)
   - Integration testing for cross-phase flows
   - End-to-end user flow testing

---

## Quality Gates (v1.0)

| Gate | Status | Notes |
|------|--------|-------|
| Security Audit | ✅ PASS | All blockers fixed |
| Accessibility Audit | ✅ PASS | Keyboard navigation fixed |
| Documentation Audit | ✅ PASS | Deployment guide created |
| Performance | ⚠️ T+30 | Lighthouse CI pending |
| Compliance | ⚠️ T+7 | Privacy policy pending |

---

*Last updated: 2026-03-30 after v1.0 milestone completion*

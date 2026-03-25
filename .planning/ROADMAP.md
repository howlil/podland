# Roadmap: Podland

**Created:** 2026-03-25
**Core Value:** Students can deploy and run applications with zero DevOps knowledge.
**Total Phases:** 5 | **Total Requirements:** 48

---

## Phase Overview

| # | Phase | Goal | Requirements | Success Criteria |
|---|-------|------|--------------|------------------|
| 1 | Foundation | Authentication + basic dashboard working | 10 | User can sign in, see dashboard, view profile |
| 2 | Core VM | VM lifecycle management + quotas | 16 | User can create, manage VMs within quota |
| 3 | Networking | Domain + tunnel automation | 6 | VM accessible via HTTPS subdomain |
| 4 | Observability | Monitoring + logging + alerting | 6 | User can view metrics, logs, receive alerts |
| 5 | Admin + Polish | Admin panel + API + launch prep | 10 | Superadmin can manage platform, API works |

---

## Phase Details

### Phase 1: Foundation

**Goal:** Users can authenticate via GitHub OAuth and access basic dashboard.

**Requirements:**
- AUTH-01: User can sign in via GitHub OAuth
- AUTH-02: System validates GitHub email matches @student.unand.ac.id pattern
- AUTH-03: System extracts NIM from email and assigns role
- AUTH-04: User session persists across browser refresh
- AUTH-05: User can view their profile
- AUTH-06: User can sign out
- DASH-01: User can view dashboard home with VM summary and quota usage
- DASH-03: User can view recent activity log
- DASH-04: Dashboard is responsive

**Success Criteria:**
1. User with @student.unand.ac.id GitHub email can sign in successfully
2. User with non-student email is rejected with clear error message
3. NIM containing "1152" is assigned Internal role, others get External
4. Session persists after browser refresh (7-day JWT expiry)
5. Profile page shows correct display name, role, and NIM
6. Sign out invalidates session and redirects to login
7. Dashboard displays "0 VMs" and quota usage (0/1 CPU for External)
8. Activity log shows "Account created" entry
9. Dashboard is usable on mobile (320px width) and desktop (1920px)

**Technical Milestones:**
- [ ] k3s cluster installed and running
- [ ] PostgreSQL database deployed
- [ ] Go backend with OAuth flow
- [ ] TanStack Start frontend deployed
- [ ] JWT authentication working
- [ ] NIM validation logic implemented

**Estimated Duration:** 3 weeks

---

### Phase 2: Core VM

**Goal:** Users can create and manage VMs with resource quotas enforced.

**Requirements:**
- VM-01: User can create a new VM with name, OS, and resource allocation
- VM-02: User can start a stopped VM
- VM-03: User can stop a running VM
- VM-04: User can restart a VM
- VM-05: User can delete a VM
- VM-06: User can view VM list with status
- VM-07: User can view VM detail page
- VM-08: System enforces per-role resource quotas
- VM-09: System prevents VM creation if quota exceeded
- VM-10: VM runs in isolated namespace with NetworkPolicy
- VM-11: VM container runs as non-root
- QUOTA-01: System tracks per-user resource usage
- QUOTA-02: User can view quota usage in dashboard
- QUOTA-03: System enforces hard limits
- QUOTA-04: Superadmin can modify user quotas
- QUOTA-05: System creates Kubernetes ResourceQuota per namespace
- API-01: REST API for VM operations
- API-02: API requires authentication via API key
- API-03: OpenAPI/Swagger documentation
- API-04: API enforces rate limiting

**Success Criteria:**
1. User can create VM with name "my-app", Ubuntu 22.04, 0.5 CPU, 1GB RAM
2. VM appears in list with status "running" within 30 seconds
3. User can stop VM, status changes to "stopped"
4. User can start stopped VM, status changes to "running"
5. User can restart VM (stop → start sequence)
6. User can delete VM with confirmation dialog
7. VM list shows all user's VMs with status badges
8. VM detail shows resource usage, domain, created date
9. External user cannot create VM exceeding 0.5 CPU / 1GB RAM
10. User with full quota sees "Quota exceeded" error on create
11. VM namespace has NetworkPolicy denying inter-namespace traffic
12. VM container runs as UID 1000, not root
13. Dashboard shows quota usage bar (0.3/1 CPU, 512MB/2GB RAM)
14. Superadmin can change user quota via database/CLI
15. API endpoint POST /api/vms creates VM with valid API key
16. API request without API key returns 401
17. Swagger docs accessible at /api/docs
18. 101st request in 1 minute returns 429 Too Many Requests

**Technical Milestones:**
- [ ] Kubernetes Deployment/Service manifests for VMs
- [ ] Pre-built Ubuntu/Debian container images
- [ ] ResourceQuota + LimitRange per namespace
- [ ] NetworkPolicy default-deny
- [ ] PodSecurityPolicy (non-root, no privilege escalation)
- [ ] VM CRUD API endpoints
- [ ] API key generation and validation
- [ ] Rate limiting middleware

**Estimated Duration:** 4 weeks

---

### Phase 3: Networking

**Goal:** VMs are accessible via HTTPS with automatic domain and tunnel setup.

**Requirements:**
- DOMAIN-01: System automatically assigns subdomain to VM
- DOMAIN-02: System creates Cloudflare DNS record via API
- DOMAIN-03: System creates Cloudflare Tunnel configuration
- DOMAIN-04: VM is accessible via HTTPS with automatic SSL
- DOMAIN-05: User can view list of domains with VM mappings
- DOMAIN-06: User can delete domain

**Success Criteria:**
1. New VM automatically gets subdomain: vm-name.podland.app
2. DNS A record created in Cloudflare within 60 seconds
3. Cloudflare Tunnel (cloudflared) deployed as sidecar
4. HTTPS request to vm-name.podland.app returns 200 OK
5. SSL certificate valid (Let's Encrypt via Cloudflare)
6. Domain list page shows all user's domains with VM names
7. User can delete domain, DNS record and tunnel removed

**Technical Milestones:**
- [ ] Cloudflare API integration (Go SDK)
- [ ] DNS record creation/deletion
- [ ] Cloudflare Tunnel automation
- [ ] Traefik IngressRoute configuration
- [ ] Automatic HTTPS redirection
- [ ] Domain management UI

**Estimated Duration:** 3 weeks

**Risk Mitigation:**
- Cloudflare API rate limiting → Exponential backoff, max 10 req/min
- DNS propagation delay → Show "pending" status, poll until active
- Tunnel connection issues → Health check, auto-restart cloudflared

---

### Phase 4: Observability

**Goal:** Users can monitor VM metrics, view logs, and receive alerts.

**Requirements:**
- MON-01: System collects CPU, RAM, disk, and network metrics
- MON-02: User can view Grafana dashboard with metrics
- MON-03: System aggregates logs from all VMs in Loki
- MON-04: User can view VM logs in dashboard
- MON-05: System generates alerts when VM CPU > 90%
- MON-06: System generates alerts when VM RAM > 85%

**Success Criteria:**
1. Prometheus scrapes metrics from all VMs every 15 seconds
2. Grafana dashboard shows CPU, RAM, disk, network graphs (24h, 7d, 30d)
3. Loki receives logs from all VMs via Promtail
4. Log viewer shows last 1000 lines with timestamps
5. Alert fires when VM CPU > 90% for 5 minutes
6. Alert fires when VM RAM > 85% for 5 minutes
7. User receives in-app notification for alerts

**Technical Milestones:**
- [ ] Prometheus Operator deployed
- [ ] ServiceMonitor for VM metrics
- [ ] Grafana deployed with dashboards
- [ ] Loki + Promtail deployed
- [ ] Log query API endpoint
- [ ] Alertmanager configuration
- [ ] Alert notification system

**Estimated Duration:** 3 weeks

---

### Phase 5: Admin + Polish

**Goal:** Superadmin can manage platform, API is complete, ready for launch.

**Requirements:**
- ADMIN-01: Superadmin can view list of all users
- ADMIN-02: Superadmin can change user role
- ADMIN-03: Superadmin can ban/unban users
- ADMIN-04: Superadmin can view system health dashboard
- ADMIN-05: System logs all admin actions to audit log
- IDLE-01: System detects idle VMs (no HTTP + no process + no login for 48h)
- IDLE-02: System sends warning notification 24h before delete
- IDLE-03: System automatically deletes idle VM after grace period
- IDLE-04: User can "pin" VM to prevent auto-delete
- VM-08: Quota enforcement verified at scale (load test)

**Success Criteria:**
1. Admin panel shows user list with role filters (All, Internal, External)
2. Superadmin can change user role from Internal to External
3. Superadmin can ban user, user cannot sign in
4. System health dashboard shows cluster CPU/RAM/storage usage
5. Audit log shows all admin actions with timestamp and IP
6. Idle detector runs every hour, identifies VMs with zero activity for 48h
7. User receives email/notification "VM will be deleted in 24h"
8. VM deleted after 24h warning if still idle
9. User can pin VM, pinned VMs excluded from auto-delete
10. Load test: 100 concurrent VMs, all quotas enforced correctly

**Technical Milestones:**
- [ ] Admin panel UI (users, system health, audit log)
- [ ] Idle detection worker (Prometheus queries + logic)
- [ ] Notification system (email + in-app)
- [ ] Auto-delete cron job
- [ ] Pin VM feature
- [ ] Load testing (k6 or locust)
- [ ] Security audit (Trivy scan, penetration test)
- [ ] Backup strategy (Velero for Kubernetes, pg_dump for PostgreSQL)

**Estimated Duration:** 3 weeks

---

## Phase Dependencies

```
Phase 1: Foundation ──┬──> Phase 2: Core VM ──> Phase 3: Networking
                      │                              │
                      └──────────────────────────────┘
                                     │
                                     ▼
                      Phase 4: Observability ──> Phase 5: Admin + Polish
```

**Critical Path:**
- Phase 1 must complete before Phase 2 (auth required for VM operations)
- Phase 2 must complete before Phase 3 (need VMs before domain setup)
- Phase 3 should complete before Phase 4 (metrics need running VMs)
- Phase 5 depends on all previous phases

---

## Risk Register

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| k3s cluster instability | High | Low | Use stable version, test upgrades |
| Cloudflare API rate limits | Medium | Medium | Exponential backoff, caching |
| Container escape vulnerability | Critical | Low | Security hardening, regular audits |
| Resource exhaustion | High | Medium | Strict quotas, monitoring alerts |
| OAuth misconfiguration | High | Low | Follow security best practices |
| TanStack Start immaturity | Medium | Low | Fallback to Vite + React Router |

---

## Launch Criteria

Phase 5 complete = Launch Ready when:

- [ ] All 48 v1 requirements implemented and tested
- [ ] Load test passed (100 concurrent VMs)
- [ ] Security audit passed (no critical/high vulnerabilities)
- [ ] Backup/restore tested successfully
- [ ] Documentation complete (user guide, API docs, runbook)
- [ ] Monitoring alerts configured for production
- [ ] On-call rotation established

---

*Roadmap created: 2026-03-25*
*Last updated: 2026-03-25 after initial creation*

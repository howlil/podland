# Research Summary: Podland PaaS

**Synthesized:** 2026-03-25
**Domain:** Student PaaS Platform
**Target:** ~500 students (UNAND)

---

## Executive Summary

Podland is feasible with the proposed stack. Key architectural decisions validated:

✅ **k3s over Docker native** — Correct choice for "cloud native" requirement
✅ **Container-as-VM** — Appropriate for shared-resource model
✅ **Conservative quotas** — Necessary for 500-user shared pool
✅ **Modular monolith** — Right starting point, extract to microservices later

---

## Stack Recommendation

### Core Stack (Validated)

| Layer | Technology | Confidence |
|-------|------------|------------|
| **Orchestration** | k3s 1.29+ | ✅ High — CNCF certified, production-ready |
| **Backend** | Go 1.21+ | ✅ High — Excellent k3s ecosystem fit |
| **Frontend** | TanStack Start + Tailwind v4 | ✅ High — Modern, type-safe, good DX |
| **Database** | PostgreSQL 16 | ✅ High — Battle-tested, JSONB flexibility |
| **Monitoring** | Prometheus + Grafana + Loki | ✅ High — Industry standard |
| **Networking** | Cloudflare DNS + Tunnel | ✅ High — Well-documented API |

### Stack Risks

| Risk | Mitigation |
|------|------------|
| TanStack Start is new (1.0) | Have fallback: Vite + React Router |
| Single-server limitation | Design for multi-node from start |
| Cloudflare dependency | Abstract Cloudflare API, can swap providers |

---

## Feature Priorities

### P0 (Must Have — Launch Blockers)

1. **GitHub OAuth + NIM validation** — Auth is foundational
2. **VM create/start/stop/delete** — Core value proposition
3. **Resource quotas (k3s native)** — Multi-tenant safety
4. **Cloudflare DNS + Tunnel automation** — User-facing domains
5. **Basic dashboard** — User interaction point

### P1 (Should Have — Launch Ready)

6. **Grafana metrics dashboard** — User visibility
7. **Idle detection (basic)** — Resource cleanup
8. **Admin panel (basic)** — Platform management
9. **Audit logging** — Security compliance

### P2 (Nice to Have — Post-Launch)

10. **Console/terminal access** — High complexity, defer
11. **Custom domains** — Can use subdomains initially
12. **Advanced idle detection** — Start with HTTP-only
13. **API + CLI** — Power users, can add later

---

## Architecture Validation

### Decision: k3s ✓

**Pros:**
- 500MB RAM footprint (vs 2GB+ for full K8s)
- CNCF certified (not a fork)
- Built-in Traefik, Metrics Server, Local Path Provisioner
- Single binary, easy upgrades
- Production-proven (100K+ nodes)

**Cons:**
- SQLite default (switch to etcd for HA later)
- Some enterprise features removed (but not needed)

**Verdict:** Correct choice for single-server PaaS.

---

### Decision: Container-as-VM ✓

**Implementation:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-vm
  namespace: user-123
spec:
  template:
    spec:
      containers:
      - name: vm
        image: podland/ubuntu-2204  # Pre-built base image
        resources:
          limits:
            cpu: "1"
            memory: 2Gi
        securityContext:
          runAsNonRoot: true
          allowPrivilegeEscalation: false
```

**Trade-offs:**
- ✅ Fast startup (<10s vs >30s for real VM)
- ✅ Efficient resource utilization
- ❌ Not true isolation (kernel shared)
- ❌ Can't run custom kernels

**Verdict:** Acceptable for student PaaS use case.

---

### Decision: Combined Idle Detection ✓

**Signals:**
1. HTTP requests (Prometheus: `http_requests_total`)
2. Process activity (containerd metrics)
3. User login (database: `last_login_at`)

**Logic:**
```
IF (http_requests == 0 AND process_idle == true AND user_login > 48h)
  THEN mark_for_deletion
  SEND warning_notification
  WAIT 24h
  DELETE vm
```

**Verdict:** Comprehensive, prevents false positives.

---

## Quota Recommendation

### Final Quota Settings

| Resource | Superadmin | Internal (SI) | External |
|----------|------------|---------------|----------|
| **CPU** | Unlimited | 1 core | 0.5 core |
| **RAM** | Unlimited | 2 GB | 1 GB |
| **Storage** | Unlimited | 10 GB | 5 GB |
| **VMs** | Unlimited | 5 VMs | 2 VMs |
| **Priority** | Highest | High | Normal |
| **Auto-delete** | Never | 2 days idle | 2 days idle |

**Rationale:**
- Conservative quotas prevent resource exhaustion
- SI students (internal) get 2x external (aligns with project goals)
- Superadmin unlimited for testing and emergencies

---

## Storage Strategy

### Recommendation: Dual Storage Class

```yaml
# Fast, ephemeral (OS, temp, cache)
storageClass: local
  - Container root filesystem
  - Application binaries
  - Temporary files

# Persistent, survives restarts (user data)
storageClass: local-lvm
  - Database files
  - User uploads
  - Configuration
```

**Why:**
- `local`: Fast (direct disk access), but deleted with container
- `local-lvm`: Survives container restarts, can backup

**Implementation:**
- Use Kubernetes PersistentVolumeClaims
- Mount to `/data` in containers
- Backup local-lvm regularly (Velero)

---

## Security Priorities

### Non-Negotiable (Phase 2)

1. **No privileged containers** — Ever
2. **Run as non-root** — Enforce via PodSecurityPolicy
3. **NetworkPolicy per namespace** — Deny inter-tenant traffic
4. **Resource limits enforced** — Prevent exhaustion
5. **Audit logging** — All admin actions logged

### Important (Phase 3-4)

6. **Secrets encryption** — sops + age
7. **Image scanning** — Trivy in CI
8. **Rate limiting** — API endpoints
9. **OAuth state validation** — Prevent CSRF

---

## Common Pitfalls (Top 5)

1. **Resource exhaustion** — Prevent with ResourceQuota
2. **Container escape** — Prevent with securityContext
3. **Noisy neighbor** — Prevent with NetworkPolicy + storage isolation
4. **Idle false positives** — Prevent with combined signals
5. **Cloudflare rate limits** — Prevent with exponential backoff

---

## Build Order (Validated)

```
Phase 1: Foundation (Weeks 1-3)
├── k3s cluster setup
├── Go backend skeleton
├── GitHub OAuth + NIM validation
└── Basic dashboard (TanStack Start)

Phase 2: Core VM (Weeks 4-7)
├── VM create/start/stop/delete
├── Resource quotas (k3s)
├── Security hardening (non-root, NetworkPolicy)
└── PostgreSQL integration

Phase 3: Networking (Weeks 8-10)
├── Cloudflare DNS API
├── Cloudflare Tunnel automation
├── Domain management UI
└── SSL/TLS automation

Phase 4: Observability (Weeks 11-13)
├── Prometheus + Grafana
├── Loki logging
├── Alerting (Alertmanager)
└── User-facing dashboards

Phase 5: Advanced (Weeks 14-16)
├── Idle detection + auto-delete
├── Admin panel
├── API + webhooks
└── Polish + launch
```

**Total:** 16 weeks (4 months) to launch-ready

---

## Go/No-Go Decision

### ✅ GO — Proceed with current plan

**Confidence Factors:**
- Stack is proven (k3s, Go, TanStack all production-ready)
- Architecture is appropriate (modular monolith → microservices)
- Feature scope is achievable (16 weeks realistic)
- Risks are known and mitigated

**Success Conditions:**
- Single server can handle 500 users with conservative quotas
- Cloudflare Tunnel provides reliable exposure
- k3s resource isolation prevents noisy neighbors

**Fallback Options:**
- If k3s too complex → Docker Compose + Traefik (lose some features)
- If TanStack Start immature → Vite + React Router (more config)
- If Cloudflare limiting → Self-hosted WireGuard + VPS

---

## Next Steps

1. **Create REQUIREMENTS.md** — Translate features into checkable requirements
2. **Create ROADMAP.md** — Phase breakdown with success criteria
3. **Start Phase 1** — k3s setup + OAuth implementation

---

*Research conducted: 2026-03-25*
*Sources: 15+ articles, docs, case studies*
*Confidence: High*

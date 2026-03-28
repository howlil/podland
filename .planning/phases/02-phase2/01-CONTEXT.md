# Phase 2 Context: Core VM

**Phase:** 2 of 5
**Goal:** Users can create and manage VMs with resource quotas enforced
**Requirements:** 16 (VM-01 through VM-11, QUOTA-01 through QUOTA-05, API-01 through API-04)
**Status:** Context gathered — ready for research and planning

---

## Prior Context (From Phase 1 + Research)

### Architecture Decisions (Locked)

| Decision | Value | Rationale |
|----------|-------|-----------|
| Orchestration | k3s | Cloud native, production-ready, 500MB RAM footprint |
| Backend | Go 1.25+ | Excellent k3s ecosystem, type-safe, performant |
| Frontend | React + TanStack Router + Tailwind v4 | Modern DX, type-safe routing |
| Database | PostgreSQL 15 | Battle-tested, JSONB flexibility |
| VM Abstraction | Docker containers with resource limits | Shared resource model, fast startup |
| Quotas | Role-based (Internal: 4 CPU/8GB/100GB, External: 0.5 CPU/1GB/10GB) | Diverse tiers for different workloads |

### Existing Infrastructure (Phase 1)

```
podland/
├── apps/
│   ├── backend/          # Go service (OAuth, JWT, sessions working)
│   └── frontend/         # React + TanStack Router (dashboard working)
├── infra/
│   ├── docker-compose/   # Local dev (postgres + backend + frontend)
│   └── k3s/              # Production deployment
│       ├── namespace.yaml
│       ├── postgres.yaml
│       ├── backend.yaml
│       ├── frontend.yaml
│       └── secrets/
│           └── local-secrets.yaml
└── packages/
```

### Reusable Patterns (From Phase 1)

- **Database:** PostgreSQL with connection pooling, migrations via SQL
- **Auth:** JWT + refresh tokens, HTTP-only cookies
- **API:** RESTful endpoints with middleware auth
- **k3s:** Namespace per environment, StatefulSet for stateful services

---

## Phase 2 Decisions (Implementation Details)

### 1. VM Abstraction & Kubernetes Resources

| Question | Decision | Rationale |
|----------|----------|-----------|
| **Kubernetes resource** | Deployment per VM | Automatic restart, health checks, rolling updates |
| **OS images** | Pre-built images (`podland/ubuntu-2204`, `podland/debian-12`) | Fast startup (<10s), consistent base |
| **Storage** | PVC per VM | Data survives restart, isolated per VM |
| **Networking** | Ingress (Traefik) with dynamic subdomain routing | Single entry point, Phase 3 ready |

**Implementation pattern:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vm-{vm-id}
  namespace: user-{user-id}
spec:
  template:
    spec:
      containers:
      - name: vm
        image: podland/ubuntu-2204:latest
        resources:
          limits:
            cpu: "0.5"
            memory: 1Gi
        volumeMounts:
        - name: vm-storage
          mountPath: /data
      volumes:
      - name: vm-storage
        persistentVolumeClaim:
          claimName: vm-{vm-id}-pvc
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: vm-{vm-id}-ingress
  namespace: user-{user-id}
spec:
  rules:
  - host: {vm-name}.podland.app
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: vm-{vm-id}-service
            port:
              number: 80
```

---

### 2. Quota Enforcement Strategy

| Question | Decision | Rationale |
|----------|----------|-----------|
| **Tracking** | Hybrid: DB tracks intended usage, k3s ResourceQuota enforces | Fast UI queries + hard enforcement |
| **Concurrency** | Database transaction with `SELECT FOR UPDATE` | Prevents race conditions |
| **Limits** | 4 limits: CPU, RAM, storage, VM count | Comprehensive resource control |
| **Superadmin edit** | Direct database edit (documented, audit via triggers) | Simple for Phase 2, UI in Phase 5 |

**Quota limits by role:**

| Resource | Internal (SI) | External |
|----------|---------------|----------|
| CPU | 4.0 cores | 0.5 core |
| RAM | 8 GB | 1 GB |
| Storage | 100 GB | 10 GB |
| VM Count | 5 VMs | 2 VMs |

**Database schema additions:**
```sql
-- user_quotas table
CREATE TABLE user_quotas (
  user_id UUID PRIMARY KEY REFERENCES users(id),
  cpu_limit DECIMAL(4,2) NOT NULL DEFAULT 0.5,
  ram_limit BIGINT NOT NULL DEFAULT 1073741824,   -- 1GB in bytes
  storage_limit BIGINT NOT NULL DEFAULT 10737418240,  -- 10GB in bytes
  vm_count_limit INTEGER NOT NULL DEFAULT 2,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- user_quota_usage table (for tracking)
CREATE TABLE user_quota_usage (
  user_id UUID PRIMARY KEY REFERENCES users(id),
  cpu_used DECIMAL(4,2) NOT NULL DEFAULT 0,
  ram_used BIGINT NOT NULL DEFAULT 0,
  storage_used BIGINT NOT NULL DEFAULT 0,
  vm_count INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- vms table
CREATE TABLE vms (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  name VARCHAR(100) NOT NULL,
  os VARCHAR(50) NOT NULL DEFAULT 'ubuntu-2204',
  tier VARCHAR(20) NOT NULL,
  cpu DECIMAL(4,2) NOT NULL,
  ram BIGINT NOT NULL,
  storage BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  k8s_namespace VARCHAR(100),
  k8s_deployment VARCHAR(100),
  k8s_service VARCHAR(100),
  k8s_pvc VARCHAR(100),
  domain VARCHAR(255),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  started_at TIMESTAMP,
  stopped_at TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX idx_vms_user_id ON vms(user_id);
CREATE INDEX idx_vms_status ON vms(status);

-- Quota check with locking (prevents race conditions)
BEGIN;
SELECT cpu_limit, ram_limit, storage_limit FROM user_quotas WHERE user_id = $1 FOR UPDATE;
-- Check if new VM fits within quota
-- Update usage
UPDATE user_quota_usage SET cpu_used = cpu_used + $1, ram_used = ram_used + $2, storage_used = storage_used + $3 WHERE user_id = $4;
COMMIT;
```

**Default quotas on user creation:**
```sql
-- External users (default)
INSERT INTO user_quotas (user_id, cpu_limit, ram_limit, storage_limit, vm_count_limit)
VALUES ($1, 0.5, 1073741824, 10737418240, 2);

-- Internal users (SI UNAND)
INSERT INTO user_quotas (user_id, cpu_limit, ram_limit, storage_limit, vm_count_limit)
VALUES ($1, 4.0, 8589934592, 107374182400, 5);
```

**Tier definitions (for validation):**
```sql
-- tiers lookup table
CREATE TABLE tiers (
  name VARCHAR(20) PRIMARY KEY,
  cpu DECIMAL(4,2) NOT NULL,
  ram BIGINT NOT NULL,
  storage BIGINT NOT NULL,
  min_role VARCHAR(20) NOT NULL DEFAULT 'external'  -- 'external' or 'internal'
);

-- Insert default tiers
INSERT INTO tiers (name, cpu, ram, storage, min_role) VALUES
  ('nano', 0.25, 536870912, 5368709120, 'external'),    -- 512MB, 5GB
  ('micro', 0.5, 1073741824, 10737418240, 'external'),  -- 1GB, 10GB
  ('small', 1.0, 2147483648, 21474836480, 'internal'),  -- 2GB, 20GB
  ('medium', 2.0, 4294967296, 42949672960, 'internal'), -- 4GB, 40GB
  ('large', 4.0, 8589934592, 85899345920, 'internal'),  -- 8GB, 80GB
  ('xlarge', 4.0, 8589934592, 107374182400, 'internal'); -- 8GB, 100GB
```

**k3s ResourceQuota example:**
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: user-quota
  namespace: user-{user-id}
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "4"
    limits.memory: 8Gi
    persistentvolumeclaims: "5"
    requests.storage: 100Gi
```

---

### 3. VM Lifecycle API Design

| Question | Decision | Rationale |
|----------|----------|-----------|
| **API style** | RESTful resource-based (`/api/vms`, `/api/vms/{id}/start`) | Standard, matches Phase 1 patterns |
| **Async handling** | Polling with status field (2s interval) | Simple, reliable, no WebSocket complexity |
| **VM states** | `pending`, `running`, `stopped`, `error` | Covers all cases, simple mental model |
| **Resource selection** | Predefined tiers (nano, micro, small, medium, large, xlarge, xxlarge) | VPS-like UX, diverse options for different workloads |

**REST API endpoints:**

```
POST   /api/vms              # Create VM (returns 202 Accepted + VM ID)
GET    /api/vms              # List user's VMs
GET    /api/vms/{id}         # Get VM details
POST   /api/vms/{id}/start   # Start VM (returns 202 Accepted)
POST   /api/vms/{id}/stop    # Stop VM (returns 202 Accepted)
POST   /api/vms/{id}/restart # Restart VM (returns 202 Accepted)
DELETE /api/vms/{id}         # Delete VM (returns 202 Accepted)
```

**VM resource model:**
```json
// POST /api/vms request
{
  "name": "my-app",
  "os": "ubuntu-2204",
  "tier": "small"
}

// VM response (all states)
{
  "id": "vm-123",
  "name": "my-app",
  "os": "ubuntu-2204",
  "tier": "small",
  "cpu": 0.5,
  "ram": 1073741824,
  "storage": 5368709120,
  "status": "running",
  "domain": "my-app.podland.app",
  "created_at": "2026-03-27T10:00:00Z",
  "updated_at": "2026-03-27T10:00:30Z"
}

// Status transitions
pending → running (on create/start)
running → stopped (on stop)
running → error (on failure)
stopped → pending (on start)
stopped → error (on failure)
```

**Resource tiers (VPS-style):**

| Tier | CPU | RAM | Storage | Monthly Cost* | Available To |
|------|-----|-----|---------|---------------|--------------|
| `nano` | 0.25 | 512 MB | 5 GB | Free | External only |
| `micro` | 0.5 | 1 GB | 10 GB | Free | Both roles |
| `small` | 1.0 | 2 GB | 20 GB | Free | Both roles |
| `medium` | 2.0 | 4 GB | 40 GB | Free | Internal only |
| `large` | 4.0 | 8 GB | 80 GB | Free | Internal only |
| `xlarge` | 4.0 | 8 GB | 100 GB | Free | Internal only (max) |

*Free tier for students (covered by conservative quota model)

**Quota enforcement by role:**

| Role | Max CPU | Max RAM | Max Storage | Max VMs | Tiers Available |
|------|---------|---------|-------------|---------|-----------------|
| External | 0.5 CPU | 1 GB | 10 GB | 2 | nano, micro |
| Internal | 4.0 CPU | 8 GB | 100 GB | 5 | All tiers |

**Notes:**
- External users can only create VMs up to `micro` tier (0.5 CPU, 1GB RAM, 10GB storage)
- Internal users can create VMs up to `xlarge` tier (4 CPU, 8GB RAM, 100GB storage)
- Users can mix tiers (e.g., 2x `nano` + 1x `micro` for External, as long as total ≤ quota)
- Storage is per-VM PVC, backed by local-lvm storage class

**Frontend polling pattern:**
```typescript
// Create VM
const response = await api.post('/api/vms', { name, os, tier })
const vmId = response.data.id

// Poll for status
const pollInterval = setInterval(async () => {
  const vm = await api.get(`/api/vms/${vmId}`)
  if (vm.data.status !== 'pending') {
    clearInterval(pollInterval)
    // VM ready
  }
}, 2000)
```

---

### 4. Security Hardening

| Question | Decision | Rationale |
|----------|----------|-----------|
| **Pod security** | PodSecurityStandard at `restricted` level | Built into k3s, no extra components |
| **Container context** | Non-root, no privilege escalation, drop ALL capabilities | Defense in depth for multi-tenant |
| **Network policy** | Default deny all (allow backend + DNS + external) | Maximum isolation |
| **ServiceAccount** | One per user namespace | Good balance of isolation and manageability |

**PodSecurityStandard enforcement:**
```yaml
# namespace.yaml (add label for restricted PSS)
apiVersion: v1
kind: Namespace
metadata:
  name: user-{user-id}
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

**Container security context:**
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: false  # Users need to write files
  capabilities:
    drop:
      - ALL
```

**NetworkPolicy (default deny + exceptions):**
```yaml
# Default deny all
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny
  namespace: user-{user-id}
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
# Allow ingress from backend
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-from-backend
  namespace: user-{user-id}
spec:
  podSelector:
    matchLabels:
      app: vm
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: podland
    - podSelector:
        matchLabels:
          app: podland-backend
  policyTypes:
  - Ingress
---
# Allow egress to DNS and external (for apt updates, etc.)
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-dns-and-external
  namespace: user-{user-id}
spec:
  podSelector:
    matchLabels:
      app: vm
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: UDP
      port: 53
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
        except:
        - 10.0.0.0/8
        - 172.16.0.0/12
        - 192.168.0.0/16
    ports:
    - protocol: TCP
      port: 80
    - protocol: TCP
      port: 443
  policyTypes:
  - Egress
```

---

## Deferred Ideas (For Later Phases)

These were considered but explicitly deferred to keep Phase 2 focused:

| Idea | Deferred To | Reason |
|------|-------------|--------|
| Custom CPU/RAM (beyond tiers) | Phase 5 | Tiers cover 90% of use cases, simpler for Phase 2 |
| Admin panel for quota management | Phase 5 | Direct DB edit works for now |
| WebSocket for real-time VM status | Phase 4 | Polling is sufficient for Phase 2 |
| Windows OS templates | Phase 5 (or v2) | Licensing complexity, resource-heavy |
| Console/terminal access | Phase 5 | WebSocket complexity, can defer |
| VM cloning/templates | Phase 5 | Nice-to-have, not core to Phase 2 |
| Paid tiers / billing | Phase 5 or v2 | Free tier for students initially |
| More granular tiers (xxlarge+) | Phase 5 | 6 tiers cover most use cases for now |

---

## Code Context (What Exists)

### Backend (Phase 1)

```
apps/backend/
├── cmd/
│   └── main.go           # Server setup, routes
├── handlers/
│   ├── auth.go           # OAuth, login, logout
│   ├── users.go          # Profile, quota display
│   ├── activity.go       # Activity log
│   └── health.go         # Health check
├── internal/
│   ├── auth/
│   │   ├── jwt.go        # JWT generation/validation
│   │   ├── oauth.go      # GitHub OAuth
│   │   └── session.go    # Session management
│   ├── database/
│   │   ├── database.go   # Connection pool
│   │   ├── queries.go    # SQL queries
│   │   └── types.go      # DB types (User, Session)
│   └── config/
│       └── config.go     # Config loading
└── middleware/
    └── middleware.go     # Auth middleware, CORS
```

### Frontend (Phase 1)

```
apps/frontend/
├── src/
│   ├── routes/
│   │   ├── __root.tsx
│   │   ├── index.tsx              # Login page
│   │   ├── dashboard/
│   │   │   ├── index.tsx          # Dashboard home
│   │   │   └── profile.tsx        # User profile
│   │   └── auth/
│   │       ├── callback.tsx       # OAuth callback
│   │       └── welcome.tsx        # First-time user
│   ├── components/
│   │   ├── layout/
│   │   │   └── DashboardLayout.tsx
│   │   └── dashboard/
│   │       ├── QuotaUsage.tsx
│   │       └── ActivityLog.tsx
│   └── lib/
│       ├── api.ts                 # Axios client
│       └── auth.ts                # Auth hooks
```

### Infrastructure (Phase 1)

```
infra/
├── docker-compose/
│   ├── docker-compose.yml   # All services (postgres + backend + frontend)
│   └── .env.example
└── k3s/
    ├── namespace.yaml
    ├── postgres.yaml        # StatefulSet + PVC + Service
    ├── backend.yaml         # Deployment + Service
    ├── frontend.yaml        # Deployment + Service + PVC
    └── secrets/
        └── local-secrets.yaml
```

---

## Research Questions (For Next Step)

These questions need research before planning:

1. **k3s client-go patterns** — What's the Go library pattern for creating Deployments, PVCs, Services programmatically?

2. **Pre-built OS images** — What's the best way to build and host Ubuntu 22.04 and Debian 12 container images? Use official images or build custom?

3. **Ingress controller** — k3s comes with Traefik, but do we need to configure it specially for dynamic subdomain routing?

4. **ResourceQuota behavior** — How does k3s ResourceQuota interact with PVC storage limits? Any gotchas?

5. **PodSecurityStandard** — What specific restrictions does `restricted` level enforce? Will our VM containers pass validation?

---

## Next Steps

**Ready for:** `/ez:research-phase 2`

Research should answer the 5 questions above, then planning can proceed with full context.

---

*Context gathered: 2026-03-27*
*Decisions locked for Phase 2 implementation*

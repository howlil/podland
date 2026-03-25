# Architecture Research: PaaS Platform

## Recommended Architecture Pattern

### Modular Monolith → Microservices Evolution

**Phase 1: Modular Monolith** (Recommended for launch)
- Single codebase, clearly separated modules
- Easier to develop, test, deploy
- Lower operational complexity for team of 1-3

**Phase 2: Microservices** (When scale demands)
- Extract high-load modules (VM management, auth)
- Independent scaling per service
- Adds operational complexity (service discovery, tracing)

**Decision:** Start with **Modular Monolith** in Go, extract to microservices when:
- Team grows beyond 5 developers
- Specific modules need independent scaling
- Clear service boundaries emerge from usage patterns

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Cloudflare                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │  DNS Mgmt   │  │   Tunnel    │  │  SSL/TLS Termination    │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ HTTPS
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    k3s Cluster (Single Server)                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    Ingress (Traefik)                      │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         ▼                    ▼                    ▼             │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐       │
│  │   Podland   │     │  Monitoring │     │   User VMs  │       │
│  │   Backend   │     │   Stack     │     │ (Containers)│       │
│  │   (Go)      │     │  (Grafana)  │     │  Namespace  │       │
│  └─────────────┘     └─────────────┘     └─────────────┘       │
│         │                                      │                │
│         ▼                                      ▼                │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐       │
│  │  PostgreSQL │     │  Prometheus │     │  Persistent │       │
│  │   (State)   │     │   Metrics   │     │  Volumes    │       │
│  └─────────────┘     └─────────────┘     └─────────────┘       │
│                                                                  │
│  Storage Classes: local (fast), local-lvm (persistent)          │
└─────────────────────────────────────────────────────────────────┘
```

## Component Architecture

### Backend Modules (Go)

```
cmd/
├── api/           # Main API server
├── worker/        # Background jobs (idle detection, cleanup)
└── migrate/       # Database migrations

internal/
├── auth/          # Authentication (GitHub OAuth, JWT, RBAC)
├── vm/            # VM lifecycle (create, start, stop, delete)
├── domain/        # Cloudflare DNS + Tunnel management
├── quota/         # Resource quota enforcement
├── metrics/       # Prometheus metrics, Grafana integration
├── idle/          # Idle detection logic (combined signals)
├── database/      # PostgreSQL connection, repositories
└── config/        # Configuration management
```

### Frontend Structure (TanStack Start)

```
apps/web/
├── app/
│   ├── __root.tsx
│   ├── index.tsx              # Dashboard home
│   ├── auth/
│   │   ├── login.tsx          # GitHub OAuth callback
│   │   └── verify.tsx         # Email verification
│   ├── vms/
│   │   ├── index.tsx          # VM list
│   │   ├── create.tsx         # Create VM wizard
│   │   └── $vmId.tsx          # VM detail + console
│   ├── domains/
│   │   ├── index.tsx          # Domain management
│   │   └── setup.tsx          # Domain setup wizard
│   ├── monitoring/
│   │   └── $vmId.tsx          # Grafana embed + custom charts
│   ├── admin/                 # Superadmin only
│   │   ├── users.tsx
│   │   ├── system.tsx
│   │   └── audit.tsx
│   └── settings/
│       └── profile.tsx
├── components/
├── lib/
└── styles/
```

## Kubernetes Namespace Strategy

### Multi-Tenant Isolation

```
Namespace: podland-system
  - Core services (Traefik, Metrics Server, etc.)
  - Platform components (backend, monitoring)

Namespace: podland-monitoring
  - Prometheus
  - Grafana
  - Loki + Promtail
  - Alertmanager

Namespace: user-{userId}
  - User VMs (containers with resource limits)
  - Per-namespace ResourceQuota
  - NetworkPolicy for isolation
```

### Resource Quota Example

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: user-quota
  namespace: user-123
spec:
  hard:
    requests.cpu: "1"        # Internal role
    requests.memory: 2Gi
    limits.cpu: "1"
    limits.memory: 2Gi
    persistentvolumeclaims: "5"
    requests.storage: 10Gi
```

## Data Flow

### VM Creation Flow

```
User → Dashboard → Create VM → Backend API
                              │
                              ▼
                    Validate Quota → Reject if exceeded
                              │
                              ▼
                    Create Namespace (if first VM)
                              │
                              ▼
                    Create Deployment (container)
                              │
                              ▼
                    Create Service (internal)
                              │
                              ▼
                    Create Ingress (Traefik)
                              │
                              ▼
                    Cloudflare DNS API → A/CNAME record
                              │
                              ▼
                    Cloudflare Tunnel → cloudflared sidecar
                              │
                              ▼
                    Return VM URL to User
```

### Idle Detection Flow

```
┌─────────────────┐
│  Prometheus     │─── HTTP requests ──┐
│  (metrics)      │                    │
└─────────────────┘                    │
                                       ▼
┌─────────────────┐           ┌─────────────────┐
│  Backend API    │─── User login ──→│  Idle Detector │
└─────────────────┘           │  (worker)       │
                                       │
┌─────────────────┐           │
│  containerd     │─── Process stats ──┘
└─────────────────┘

Idle Detector (every 1 hour):
  1. Query Prometheus for HTTP traffic (last 48h)
  2. Query containerd for process activity
  3. Query database for user login (last 48h)
  4. If ALL three = zero → Mark for deletion
  5. Send warning notification (24h before delete)
  6. Delete VM after grace period
```

## Build Order (Dependency Graph)

```
Phase 1: Foundation
  └─→ Authentication (GitHub OAuth, RBAC)
      └─→ User Dashboard (basic)

Phase 2: Core
  └─→ VM Management (create, start, stop)
      └─→ Resource Quotas
          └─→ k3s integration

Phase 3: Networking
  └─→ Cloudflare DNS API
      └─→ Cloudflare Tunnel automation
          └─→ Domain management UI

Phase 4: Observability
  └─→ Prometheus + Grafana
      └─→ Loki logging
          └─→ Alerting

Phase 5: Advanced
  └─→ Idle detection
      └─→ Auto-delete
          └─→ Admin panel
```

## Security Boundaries

```
┌─────────────────────────────────────────────────────────────┐
│  Trust Boundary: User → Platform                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  User Namespace 1    User Namespace 2    Platform Namespace │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │ NetworkPolicy│    │ NetworkPolicy│    │ NetworkPolicy│  │
│  │ Deny All     │    │ Deny All     │    │ Allow Mgmt   │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│         │                   │                   │           │
│         └───────────────────┼───────────────────┘           │
│                             │                               │
│                    ┌────────▼────────┐                      │
│                    │  API Gateway    │                      │
│                    │  (Auth + Rate)  │                      │
│                    └─────────────────┘                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘

Security Measures:
- NetworkPolicy: Deny inter-namespace traffic by default
- RBAC: Namespace-scoped service accounts
- Secrets: Encrypted at rest (sops + age)
- Images: Private registry, vulnerability scanning
- Audit: All API calls logged
```

## Component Boundaries

| Component | Responsibility | Interface |
|-----------|---------------|-----------|
| **Auth Module** | GitHub OAuth, JWT, RBAC, NIM validation | REST API, middleware |
| **VM Module** | Container lifecycle, resource limits | Kubernetes API |
| **Domain Module** | Cloudflare DNS, Tunnel setup | Cloudflare API |
| **Quota Module** | Quota enforcement, usage tracking | Kubernetes ResourceQuota |
| **Metrics Module** | Prometheus, Grafana, alerts | Prometheus API |
| **Idle Module** | Idle detection, cleanup | Background worker |
| **Database** | User data, VM metadata, audit logs | PostgreSQL |

## Scaling Strategy

### Vertical Scaling (Phase 1)
- Increase server resources (CPU, RAM, storage)
- k3s handles scheduling automatically
- Limit: Single server capacity

### Horizontal Scaling (Phase 2+)
- Add more servers to k3s cluster
- Proxmox manages bare metal provisioning
- Traefik load balances across nodes
- Requires: Shared storage (Longhorn, Ceph)

### Burst Scaling
- Cloud burst to AWS/GCP during peak
- Requires: Multi-cloud networking
- Defer to v2+

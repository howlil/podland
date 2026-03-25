# Features Research: PaaS Platform

## Feature Categories for Student PaaS

### 1. Authentication & Authorization

**Table Stakes:**
- GitHub OAuth login (required)
- Email domain verification (@student.unand.ac.id)
- NIM validation for role assignment (SI vs non-SI)
- Session management with JWT
- Role-based access control (superadmin, internal, external)
- Logout and session invalidation

**Differentiators:**
- Magic link fallback (if GitHub unavailable)
- 2FA for superadmin accounts
- Session activity logging
- Automatic role promotion/demotion based on NIM

**Complexity Notes:**
- GitHub OAuth: Low complexity (mature libraries)
- Email verification: Medium (need student database or pattern matching)
- NIM validation: Low (pattern: YY1152XX = SI, where 1152 is SI code)

### 2. VM (Container) Management

**Table Stakes:**
- Create "VM" (container) with resource limits
- Select OS template (Ubuntu, Debian)
- Configure CPU/RAM allocation (within quota)
- Start/stop/restart VM
- Delete VM
- VM status dashboard (running, stopped, pending)
- Console/terminal access to VM

**Differentiators:**
- VM templates (pre-configured stacks: Node.js, Go, Python)
- Custom OS images (user uploads)
- VM cloning (duplicate existing VM)
- Scheduled start/stop (cron-based)
- Auto-delete after 2 days idle (combined: no HTTP + no process + no login)

**Complexity Notes:**
- Container creation: Medium (Kubernetes Jobs/Deployments)
- OS templates: Medium (pre-built container images)
- Console access: High (need WebSocket terminal, e.g., xterm.js + ttyd)
- Idle detection: Medium (Prometheus metrics + custom logic)

### 3. Domain & Networking

**Table Stakes:**
- Automatic subdomain assignment (vm-name.podland.app)
- Cloudflare DNS API integration
- Cloudflare Tunnel automatic setup
- Custom domain support (CNAME to Podland)
- SSL/TLS automatic (Let's Encrypt via Cloudflare)
- Domain list per user

**Differentiators:**
- Domain marketplace (users can share/sell domains)
- Automatic HTTPS redirection
- Custom SSL certificates (user-provided)
- DNS record management UI (A, CNAME, MX, TXT)

**Complexity Notes:**
- Cloudflare API: Low (excellent Go SDK)
- Tunnel automation: Medium (cloudflared in sidecar)
- Custom domains: Medium (validation, DNS propagation checks)

### 4. Resource Quotas & Limits

**Table Stakes:**
- Per-role quotas:
  - Superadmin: Unlimited
  - Internal (SI): 1 CPU, 2GB RAM, 10GB storage, 5 VMs max
  - External: 0.5 CPU, 1GB RAM, 5GB storage, 2 VMs max
- Real-time quota usage display
- Quota enforcement (hard limits)
- Quota increase request workflow

**Differentiators:**
- Dynamic quota adjustment (based on usage patterns)
- Quota borrowing (borrow from future allocation)
- Quota marketplace (users can trade quotas)

**Complexity Notes:**
- Kubernetes ResourceQuota: Low (native feature)
- Real-time display: Low (Prometheus metrics)
- Quota requests: Medium (approval workflow)

### 5. Monitoring & Observability

**Table Stakes:**
- Grafana dashboard per VM (CPU, RAM, disk, network)
- Prometheus metrics collection
- Loki log aggregation
- Alert thresholds (CPU > 90%, RAM > 85%)
- Historical data (30 days retention)
- Export metrics (CSV, JSON)

**Differentiators:**
- Custom metrics (user-defined)
- Alert webhooks (Discord, Slack, email)
- Anomaly detection (ML-based)
- Cost estimation (resource usage → cost)

**Complexity Notes:**
- Grafana dashboards: Low (pre-built templates)
- Metrics collection: Low (Prometheus native)
- Log aggregation: Medium (Loki + Promtail setup)
- Alerts: Medium (Alertmanager configuration)

### 6. User Dashboard

**Table Stakes:**
- VM list with status
- Create VM wizard
- Resource usage summary
- Recent activity log
- Profile settings
- API key management

**Differentiators:**
- Dark/light mode
- Mobile-responsive design
- Keyboard shortcuts
- Command palette (Cmd+K)
- Usage analytics (charts, trends)

### 7. Admin Panel (Superadmin Only)

**Table Stakes:**
- User management (list, ban, role change)
- System health dashboard
- Resource utilization across all VMs
- Audit logs
- Platform announcements

**Differentiators:**
- Automated moderation (flag suspicious activity)
- Bulk operations (batch VM operations)
- Maintenance mode
- Backup/restore UI

### 8. API & Integrations

**Table Stakes:**
- REST API for all operations
- API documentation (OpenAPI/Swagger)
- Rate limiting per API key
- Webhook support (VM events)

**Differentiators:**
- GraphQL API
- Terraform provider for Podland
- CLI tool for power users
- GitHub App integration (auto-deploy on push)

## Out of Scope (Anti-Features)

| Feature | Why Exclude |
|---------|-------------|
| Real VM (qemu/kvm) | Contradicts shared-resource model, high overhead |
| Windows OS templates | Licensing complexity, resource-heavy |
| GPU acceleration | Single-server limitation, niche use case |
| Multi-region deployment | Single-server constraint |
| Managed databases | Out of scope, users can run their own in VMs |
| Load balancer as service | Users can deploy their own (Traefik, NGINX) |
| CI/CD pipelines | Users can self-host (GitHub Actions, Drone) |

## Feature Dependencies

```
Authentication → VM Management → Monitoring
     ↓                ↓
Domain Setup ←───────┘
     ↓
Observability
```

## Complexity Summary

| Category | Complexity | Risk |
|----------|------------|------|
| Authentication | Low | Low |
| VM Management | Medium-High | Medium |
| Domain/Networking | Medium | Low |
| Resource Quotas | Low | Low |
| Monitoring | Medium | Low |
| Dashboard | Low | Low |
| Admin Panel | Low | Low |
| API | Low | Low |

**Highest Risk:** Console/terminal access (security implications), idle detection (accuracy)

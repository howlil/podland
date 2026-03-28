# Phase 3 Context: Networking

**Phase:** 3 of 5
**Goal:** VMs are accessible via HTTPS with automatic domain and tunnel setup
**Requirements:** 6 (DOMAIN-01 through DOMAIN-06)
**Status:** Context gathered — all decisions locked, ready for research and planning

---

## Prior Context (From Phase 1 + Phase 2)

### Architecture Decisions (Locked)

| Decision | Value | Rationale |
|----------|-------|-----------|
| Orchestration | k3s | Cloud native, production-ready, 500MB RAM footprint |
| Backend | Go 1.25+ | Excellent k3s ecosystem, type-safe, performant |
| Frontend | React + TanStack Router + Tailwind v4 | Modern DX, type-safe routing |
| Database | PostgreSQL 15 | Battle-tested, JSONB flexibility |
| VM Abstraction | Docker containers with resource limits | Shared resource model, fast startup |
| Ingress | Traefik | Already configured for wildcard subdomain routing |

### Existing Infrastructure (Phase 1 + 2)

```
podland/
├── apps/
│   ├── backend/          # Go service (OAuth, JWT, sessions, VM CRUD working)
│   └── frontend/         # React + TanStack Router (dashboard, VM management working)
├── infra/
│   ├── docker-compose/   # Local dev (postgres + backend + frontend)
│   └── k3s/              # Production deployment
│       ├── namespace.yaml
│       ├── postgres.yaml
│       ├── backend.yaml
│       ├── frontend.yaml
│       ├── traefik-config.yaml  # IngressRouteTCP for *.podland.app
│       └── secrets/
│           └── local-secrets.yaml
└── packages/
    └── types/            # VM types include domain?: string field
```

### Reusable Patterns (From Phase 1 + 2)

- **Database:** PostgreSQL with connection pooling, migrations via SQL
- **Auth:** JWT + refresh tokens, HTTP-only cookies
- **API:** RESTful endpoints with middleware auth
- **k3s:** Namespace per user, Deployment per VM, PVC for storage
- **VM Model:** Already has `domain?: string` field ready for Phase 3

### Phase 2 Decisions Applied

- VM = Kubernetes Deployment with resource limits
- PVC per VM for persistent storage
- Traefik IngressRoute configured for `*.podland.app` wildcard routing
- NetworkPolicy default-deny for isolation
- VM status tracking (running, stopped, pending)

---

## Phase 3 Decisions (Implementation Details)

All decisions below are locked. Research and planning agents use these to determine what to investigate and how to implement.

### 1. Cloudflare DNS Automation

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **DNS Record Type** | CNAME | `vm-name.podland.app CNAME tunnel.podland.app` — single point of change, works with Cloudflare Tunnel, no updates when server IP changes |
| **API Integration** | Go SDK (`github.com/cloudflare/cloudflare-go`) | Type-safe, integrated with existing Go backend, real-time feedback, easy testing with mocks |
| **DNS Propagation** | Show "pending" status + poll | Standard cloud pattern (AWS, GCP, Vercel), user sees progress, backend retries on failure |
| **Rate Limiting** | Exponential backoff | Simple, proven pattern, built into Go SDK, sufficient for 500 users (no queue infrastructure needed) |

**Implementation Pattern:**
```go
// Backend creates DNS record on VM creation
func CreateVMDNS(vmName, vmIP string) (string, error) {
    subdomain := fmt.Sprintf("%s.podland.app", vmName)
    
    // Create CNAME record via Go SDK
    record := cloudflare.DNSRecord{
        Type: "CNAME",
        Name: subdomain,
        Content: "tunnel.podland.app",
        Proxied: true, // Enable Cloudflare proxy
    }
    
    err := api.CreateDNSRecord(ctx, zoneID, record)
    // On 429: exponential backoff (2^attempt seconds, max 3 retries)
    
    return subdomain, err
}

// Background worker polls until DNS active
func PollDNSStatus(subdomain string) error {
    for i := 0; i < 30; i++ { // 5 min timeout
        record := api.GetDNSRecord(subdomain)
        if record.Status == "active" {
            return nil
        }
        time.Sleep(10 * time.Second)
    }
    return errors.New("DNS propagation timeout")
}
```

**User Flow:**
```
1. User creates VM with name "my-app"
2. Backend generates subdomain: "my-app.podland.app"
3. Backend creates CNAME record via Cloudflare API
4. VM status shows "pending DNS" with loading indicator
5. Background worker polls Cloudflare API every 10s
6. When DNS active, VM status changes to "running"
7. User can access https://my-app.podland.app
```

---

### 2. Cloudflare Tunnel Setup

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Deployment Model** | Single Deployment | One `cloudflared` for entire cluster — 50MB RAM total, matches Cloudflare docs, only viable option for free tier (5 tunnel limit) |
| **Tunnel Scope** | Per Cluster | Single tunnel for all VMs — free tier compatible, no per-VM tunnel management overhead |
| **Authentication** | Service tokens | Cloudflare's recommended approach — one-time setup, automatic rotation, no backend involvement |
| **Failover** | Kubernetes restart | Let k8s restart failed pods — simple, built-in, ~30s downtime acceptable for v1 |

**Implementation Pattern:**
```yaml
# infra/k3s/cloudflared.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudflared
  namespace: podland
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cloudflared
  template:
    metadata:
      labels:
        app: cloudflared
    spec:
      containers:
      - name: cloudflared
        image: cloudflare/cloudflared:latest
        args:
        - tunnel
        - --config
        - /etc/cloudflared/config.yml
        - run
        livenessProbe:
          httpGet:
            path: /health
            port: 2000
          initialDelaySeconds: 10
          periodSeconds: 30
        volumeMounts:
        - name: config
          mountPath: /etc/cloudflared
      volumes:
      - name: config
        secret:
          secretName: cloudflared-config
---
apiVersion: v1
kind: Secret
metadata:
  name: cloudflared-config
  namespace: podland
type: Opaque
stringData:
  config.yml: |
    tunnel: podland-cluster-tunnel
    credentials-file: /etc/cloudflared/creds.json
    ingress:
    - hostname: "*.podland.app"
      service: http://traefik.podland.svc:80
    - service: http_status:404
```

**Resource Math:**
- Single deployment: 1 × 50MB = 50MB RAM (negligible)
- Sidecar model: 500 VMs × 50MB = 25GB RAM (impossible)
- Free tier: 5 tunnels max (per-VM tunnels require $500/month)

---

### 3. Domain Assignment Flow

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Assignment** | Hybrid (auto-assign default, allow customize later) | Best UX — works out of box like Vercel/Netlify, user can customize in settings (Phase 5) |
| **Subdomain Format** | `{vm-name}.podland.app` | Memorable, user-chosen, short URLs — collision handled with user suffix |
| **Custom Domains** | Not in v1 | Avoid SSL/DNS verification complexity — ship Phase 3 faster, add in Phase 5 |
| **Deletion** | Auto-delete DNS | Clean, secure, standard cloud pattern — prevents orphaned DNS and subdomain takeover |

**Implementation Pattern:**
```go
// VM creation with automatic domain assignment
func CreateVM(input VMCreateInput, userID string) (*VM, error) {
    // 1. Create VM in k8s
    vm := createKubernetesDeployment(input)
    
    // 2. Generate subdomain from VM name
    subdomain := sanitizeSubdomain(input.Name) // "My App" → "my-app"
    
    // 3. Handle collisions
    if domainExists(subdomain) {
        subdomain = fmt.Sprintf("%s-user%s", subdomain, userID)
    }
    
    // 4. Create DNS record (CNAME to tunnel)
    createDNSRecord(subdomain, vm.ID)
    
    // 5. Return VM with domain
    vm.Domain = subdomain + ".podland.app"
    vm.DomainStatus = "pending"
    return vm, nil
}

// VM deletion with auto-cleanup
func DeleteVM(vmID string) error {
    vm := getVM(vmID)
    
    // 1. Delete k8s resources
    deleteKubernetesDeployment(vmID)
    
    // 2. Delete DNS record (auto-cleanup)
    deleteDNSRecord(vm.Domain)
    
    // 3. Delete database record
    deleteVMFromDB(vmID)
    
    return nil
}
```

**Collision Handling:**
- "blog" + user-123 → `blog.podland.app` (if available)
- "blog" + user-456 → `blog-user456.podland.app` (if "blog" taken)

**Subdomain Sanitization:**
- Lowercase: `MyApp` → `my-app`
- Special chars: `my_app!` → `my-app`
- Length limit: 63 chars (DNS spec)

---

### 4. SSL/TLS Management

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **SSL Mode** | Full (strict) | End-to-end encryption (Cloudflare → Origin encrypted), certificate validation prevents MITM, industry standard |
| **Certificates** | Cloudflare Origin CA | 15-year validity (set and forget), free, auto-renewable, wildcard support, instant issuance |
| **Scope** | Wildcard `*.podland.app` | One cert for all VMs — instant subdomain SSL, no per-VM cert management |
| **Renewal** | Automatic (Origin CA) | 15-year validity — optional annual renewal cron, no expiry tracking needed |

**Implementation Pattern:**
```yaml
# Generate Origin CA cert (one-time setup)
# 1. Go to Cloudflare Dashboard → SSL/TLS → Origin Server → Create Certificate
# 2. Generate wildcard cert for *.podland.app
# 3. Download cert and key
# 4. Create k8s secret

apiVersion: v1
kind: Secret
metadata:
  name: origin-ca-cert
  namespace: podland
type: kubernetes.io/tls
stringData:
  tls.crt: |
    -----BEGIN CERTIFICATE-----
    (Origin CA wildcard cert for *.podland.app)
    -----END CERTIFICATE-----
  tls.key: |
    -----BEGIN PRIVATE KEY-----
    (Private key)
    -----END PRIVATE KEY-----
---
# Traefik uses this secret for TLS termination
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: vm-ingress
  namespace: podland
  annotations:
    traefik.ingress.kubernetes.io/router.tls: "true"
spec:
  tls:
  - hosts:
    - "*.podland.app"
    secretName: origin-ca-cert
  rules:
  - host: "*.podland.app"
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: traefik
            port: 80
```

**Security Comparison:**
| Mode | User → Cloudflare | Cloudflare → Origin | Mitigation |
|------|-------------------|---------------------|------------|
| Flexible | HTTPS | **HTTP** (unencrypted) | ❌ Network MITM possible |
| Full | HTTPS | HTTPS (self-signed) | ⚠️ No cert validation |
| **Full (strict)** | **HTTPS** | **HTTPS (validated)** | ✅ End-to-end encrypted + verified |

---

## Code Context (Scouted from Codebase)

### Existing Assets for Phase 3

**Backend (`apps/backend/`):**
- `handler/vm_handler.go` — VM CRUD handlers (already has `Domain` field in response)
- `internal/entity/vm.go` — VM entity with `Domain string` field
- `internal/k8s/vm_manager.go` — Kubernetes VM operations (extend for DNS)
- `internal/config/config.go` — Config loading (add Cloudflare API key)

**Frontend (`apps/frontend/`):**
- `lib/api.ts` — API client (extend with domain endpoints)
- `components/dashboard/VMCountCard.tsx` — VM list (add domain display)
- `routes/dashboard/-vms/$id.tsx` — VM detail (add domain status badge)

**Types (`packages/types/`):**
- `Domain` interface already defined (id, vmId, userId, subdomain, domain, status)
- `VM` interface has optional `domain?: string` field
- `ErrorCodes` has DOMAIN_NOT_FOUND, DOMAIN_ALREADY_EXISTS, DOMAIN_PENDING

**Infrastructure (`infra/k3s/`):**
- `traefik-config.yaml` — IngressRouteTCP already configured for `*.podland.app`
- Existing namespace structure ready for cloudflared deployment

---

## Requirements (From ROADMAP.md)

| ID | Requirement | Success Criteria |
|----|-------------|------------------|
| DOMAIN-01 | System automatically assigns subdomain to VM | New VM gets `vm-name.podland.app` |
| DOMAIN-02 | System creates Cloudflare DNS record via API | DNS A/CNAME record created within 60s |
| DOMAIN-03 | System creates Cloudflare Tunnel configuration | cloudflared deployed, tunnel active |
| DOMAIN-04 | VM is accessible via HTTPS with automatic SSL | HTTPS request returns 200 OK |
| DOMAIN-05 | User can view list of domains with VM mappings | Domain list page shows all user's domains |
| DOMAIN-06 | User can delete domain | Delete removes DNS + tunnel config |

---

## Technical Milestones (From ROADMAP.md)

- [ ] Cloudflare API integration (Go SDK)
- [ ] DNS record creation/deletion
- [ ] Cloudflare Tunnel automation (cloudflared deployment)
- [ ] Traefik IngressRoute configuration
- [ ] Automatic HTTPS redirection
- [ ] Domain management UI

---

## Risk Mitigation (From ROADMAP.md)

| Risk | Mitigation |
|------|------------|
| Cloudflare API rate limiting | Exponential backoff, max 10 req/min |
| DNS propagation delay | Show "pending" status, poll until active |
| Tunnel connection issues | Health check, k8s auto-restart |

---

## Deferred Ideas (For Future Phases)

**Phase 5 candidates (not in v1):**
- Custom domain support (user brings own domain like `myapp.com`)
- Per-VM tunnel isolation (requires paid Cloudflare plan)
- Domain vanity URLs (short links, QR codes)
- SSL certificate transparency logs
- HSTS preloading

**Out of scope for Phase 3:**
- Let's Encrypt integration (Origin CA is sufficient)
- Per-VM certificates (wildcard covers all)
- DNS verification flows (no custom domains in v1)
- Certificate renewal monitoring (15-year validity)

---

## Next Steps

**For Research Agent:**
1. Cloudflare Go SDK usage patterns (DNS record CRUD)
2. cloudflared Kubernetes deployment examples
3. Cloudflare Origin CA certificate installation
4. Traefik + Cloudflare Tunnel integration patterns

**For Planning Agent:**
1. Break down 6 requirements into implementation tasks
2. Estimate effort per task (hours/days)
3. Define acceptance criteria per task
4. Identify dependencies (DNS before tunnel, tunnel before SSL)

---

*Context created: 2026-03-28*
*All 16 decisions locked — ready for research and planning*

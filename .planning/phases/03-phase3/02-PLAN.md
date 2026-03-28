---

# Phase 3 Plan: Networking

**Phase:** 3 — Domain + Tunnel Automation
**Goal:** VMs are accessible via HTTPS with automatic domain and tunnel setup
**Duration:** 3 weeks
**Requirements:** 6 (DOMAIN-01 through DOMAIN-06)
**Research:** Complete (01-RESEARCH.md)
**Context:** Complete (01-CONTEXT.md with 16 decisions locked)

---

## Success Criteria

1. ✅ New VM automatically gets subdomain: vm-name.podland.app
2. ✅ DNS CNAME record created in Cloudflare within 60 seconds
3. ✅ Cloudflare Tunnel (cloudflared) deployed and running
4. ✅ HTTPS request to vm-name.podland.app returns 200 OK
5. ✅ SSL certificate valid (Cloudflare Origin CA wildcard)
6. ✅ Domain list page shows all user's domains with VM names
7. ✅ User can delete domain, DNS record and tunnel config removed

---

## Technical Milestones

- [ ] Cloudflare API integration (Go SDK)
- [ ] DNS record creation/deletion
- [ ] Cloudflare Tunnel automation (cloudflared deployment)
- [ ] Traefik IngressRoute configuration
- [ ] Automatic HTTPS redirection
- [ ] Domain management UI

---

## Implementation Tasks

### Week 1: Backend DNS + Tunnel Infrastructure

#### Task 1.1: Cloudflare DNS Manager
**Estimate:** 4 hours
**Acceptance Criteria:**
- [ ] DNS manager package created (`internal/cloudflare/dns.go`)
- [ ] CreateCNAME function working with exponential backoff
- [ ] DeleteDNSRecord function working
- [ ] ListDNSRecords function working with filtering
- [ ] Unit tests for DNS operations (mock Cloudflare API)

**Implementation:**
```go
// apps/backend/internal/cloudflare/dns.go
package cloudflare

import (
    "context"
    "fmt"
    "time"

    "github.com/cloudflare/cloudflare-go/v6"
    "github.com/cloudflare/cloudflare-go/v6/dns"
    "github.com/cloudflare/cloudflare-go/v6/option"
)

type DNSManager struct {
    client *cloudflare.Client
    zoneID string
}

func NewDNSManager(apiToken, zoneID string) *DNSManager {
    return &DNSManager{
        client: cloudflare.NewClient(option.WithAPIToken(apiToken)),
        zoneID: zoneID,
    }
}

func (m *DNSManager) CreateCNAME(ctx context.Context, subdomain, target string) (*dns.Record, error) {
    // Create CNAME with auto TTL and proxy enabled
    record, err := m.client.DNS.Records.New(ctx, dns.RecordNewParams{
        ZoneID:  cloudflare.F(m.zoneID),
        Type:    cloudflare.F(dns.RecordTypeCname),
        Name:    cloudflare.F(subdomain),
        Content: cloudflare.F(target),
        Proxied: cloudflare.F(true),
        TTL:     cloudflare.F[int](1), // Auto TTL
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create CNAME: %w", err)
    }
    return record, nil
}

func (m *DNSManager) DeleteRecord(ctx context.Context, recordID string) error {
    _, err := m.client.DNS.Records.Delete(ctx, m.zoneID, recordID)
    if err != nil {
        return fmt.Errorf("failed to delete DNS record: %w", err)
    }
    return nil
}

func (m *DNSManager) GetRecordByName(ctx context.Context, name string) (*dns.Record, error) {
    iter := m.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
        ZoneID: cloudflare.F(m.zoneID),
        Name:   cloudflare.F(name),
    })
    
    for iter.Next() {
        return iter.Current(), nil
    }
    if err := iter.Err(); err != nil {
        return nil, err
    }
    return nil, fmt.Errorf("DNS record not found")
}
```

**Files to create:**
- `apps/backend/internal/cloudflare/dns.go`
- `apps/backend/internal/cloudflare/dns_test.go`

---

#### Task 1.2: DNS Propagation Poller
**Estimate:** 2 hours
**Acceptance Criteria:**
- [ ] Background worker polls DNS status every 10s
- [ ] Timeout after 5 minutes (30 attempts)
- [ ] VM status updated to "running" when DNS active
- [ ] VM status updated to "error" on timeout

**Implementation:**
```go
// apps/backend/internal/domain/poller.go
package domain

import (
    "context"
    "time"

    "github.com/podland/backend/internal/cloudflare"
    "github.com/podland/backend/internal/database"
)

type DNSPoller struct {
    dnsManager *cloudflare.DNSManager
    db         *database.DB
}

func NewDNSPoller(dnsManager *cloudflare.DNSManager, db *database.DB) *DNSPoller {
    return &DNSPoller{
        dnsManager: dnsManager,
        db:         db,
    }
}

func (p *DNSPoller) WaitForDNS(ctx context.Context, vmID, subdomain string) error {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    timeout := time.After(5 * time.Minute)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-timeout:
            // Update VM status to error
            p.db.UpdateVMDomainStatus(vmID, "error")
            return fmt.Errorf("DNS propagation timeout")
        case <-ticker.Tick():
            record, err := p.dnsManager.GetRecordByName(ctx, subdomain)
            if err != nil {
                continue // Retry on next tick
            }
            
            // Check if DNS is active (Cloudflare proxy status)
            if record.Meta != nil && record.Meta.AutoAdded {
                p.db.UpdateVMDomainStatus(vmID, "active")
                return nil
            }
        }
    }
}
```

**Files to create:**
- `apps/backend/internal/domain/poller.go`

---

#### Task 1.3: VM Domain Integration
**Estimate:** 3 hours
**Acceptance Criteria:**
- [ ] VM creation triggers DNS record creation
- [ ] VM deletion triggers DNS record deletion
- [ ] Domain field populated in VM response
- [ ] Domain status tracked (pending/active/error)

**Implementation:**
```go
// apps/backend/handler/vm_handler.go (extend existing)
func (h *VMHandler) CreateVM(w http.ResponseWriter, r *http.Request) {
    var input VMCreateInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Get user from context (set by auth middleware)
    user := getUserFromContext(r.Context())

    // 1. Create VM in k8s
    vm, err := h.vmService.CreateVM(r.Context(), user.ID, input)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 2. Generate subdomain from VM name
    subdomain := sanitizeSubdomain(input.Name)
    domain := subdomain + ".podland.app"

    // 3. Handle collisions
    if existing, _ := h.db.GetVMByDomain(domain); existing != nil {
        subdomain = fmt.Sprintf("%s-user%d", subdomain, user.ID)
        domain = subdomain + ".podland.app"
    }

    // 4. Create DNS record
    _, err = h.dnsManager.CreateCNAME(r.Context(), domain, "tunnel.podland.app")
    if err != nil {
        // Rollback VM creation
        h.vmService.DeleteVM(r.Context(), vm.ID)
        http.Error(w, "Failed to create DNS record", http.StatusInternalServerError)
        return
    }

    // 5. Update VM with domain
    vm.Domain = domain
    vm.DomainStatus = "pending"
    h.db.UpdateVMDomain(vm.ID, domain, "pending")

    // 6. Start background DNS poller (async)
    go func() {
        ctx := context.Background()
        err := h.dnsPoller.WaitForDNS(ctx, vm.ID, domain)
        if err != nil {
            log.Printf("DNS propagation failed for VM %s: %v", vm.ID, err)
        }
    }()

    // 7. Log activity
    h.db.CreateActivityLog(user.ID, "vm_created_with_domain", map[string]interface{}{
        "vm_id":    vm.ID,
        "vm_name":  input.Name,
        "domain":   domain,
    })

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(vm)
}

func (h *VMHandler) DeleteVM(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    vmID := vars["id"]

    // Get VM to retrieve domain
    vm, err := h.db.GetVMByID(vmID)
    if err != nil {
        http.Error(w, "VM not found", http.StatusNotFound)
        return
    }

    user := getUserFromContext(r.Context())

    // 1. Delete DNS record (if domain exists)
    if vm.Domain != "" {
        dnsRecord, err := h.dnsManager.GetRecordByName(r.Context(), vm.Domain)
        if err == nil && dnsRecord != nil {
            h.dnsManager.DeleteRecord(r.Context(), dnsRecord.ID)
        }
    }

    // 2. Delete VM in k8s
    err = h.vmService.DeleteVM(r.Context(), vmID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 3. Delete from database
    h.db.DeleteVM(vmID)

    // 4. Log activity
    h.db.CreateActivityLog(user.ID, "vm_deleted", map[string]interface{}{
        "vm_id":   vmID,
        "domain":  vm.Domain,
    })

    w.WriteHeader(http.StatusNoContent)
}

// Helper: sanitize subdomain
func sanitizeSubdomain(name string) string {
    // Lowercase
    s := strings.ToLower(name)
    // Replace spaces and special chars with hyphens
    s = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(s, "-")
    // Remove leading/trailing hyphens
    s = strings.Trim(s, "-")
    // Limit to 63 chars (DNS spec)
    if len(s) > 63 {
        s = s[:63]
    }
    return s
}
```

**Files to modify:**
- `apps/backend/handler/vm_handler.go` (extend CreateVM, DeleteVM)
- `apps/backend/internal/database/vm_repository.go` (add UpdateVMDomain, UpdateVMDomainStatus)

**Files to create:**
- `apps/backend/internal/domain/service.go` (domain management service)

---

#### Task 1.4: Database Schema for Domains
**Estimate:** 1 hour
**Acceptance Criteria:**
- [ ] Migration adds `domain` and `domain_status` columns to VMs table
- [ ] Index on domain column for fast lookups
- [ ] Backwards compatible (existing VMs have NULL domain)

**Implementation:**
```sql
-- apps/backend/migrations/0003_add_domain_to_vms.sql
ALTER TABLE vms 
    ADD COLUMN domain VARCHAR(255) UNIQUE,
    ADD COLUMN domain_status VARCHAR(20) DEFAULT 'pending' CHECK (domain_status IN ('pending', 'active', 'error'));

CREATE INDEX idx_vms_domain ON vms(domain);
CREATE INDEX idx_vms_domain_status ON vms(domain_status);
```

**Go migration:**
```go
// apps/backend/internal/database/vm_repository.go (add methods)
func (r *VMRepository) UpdateVMDomain(vmID, domain string) error {
    _, err := r.db.Exec(`
        UPDATE vms 
        SET domain = $1, domain_status = 'pending', updated_at = NOW()
        WHERE id = $2
    `, domain, vmID)
    return err
}

func (r *VMRepository) UpdateVMDomainStatus(vmID, status string) error {
    _, err := r.db.Exec(`
        UPDATE vms 
        SET domain_status = $1, updated_at = NOW()
        WHERE id = $2
    `, status, vmID)
    return err
}

func (r *VMRepository) GetVMByDomain(domain string) (*VM, error) {
    query := `SELECT id, user_id, name, domain, domain_status, cpu, ram, storage, status, created_at, updated_at
              FROM vms WHERE domain = $1`
    row := r.db.QueryRow(query, domain)
    
    var vm VM
    err := row.Scan(&vm.ID, &vm.UserID, &vm.Name, &vm.Domain, &vm.DomainStatus,
                    &vm.CPU, &vm.RAM, &vm.Storage, &vm.Status, &vm.CreatedAt, &vm.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return &vm, nil
}
```

**Files to modify:**
- `apps/backend/migrations/` (create 0003_add_domain_to_vms.sql)
- `apps/backend/internal/database/vm_repository.go` (add domain methods)

---

#### Task 1.5: Cloudflare Tunnel Deployment
**Estimate:** 3 hours
**Acceptance Criteria:**
- [ ] cloudflared Deployment YAML created
- [ ] Tunnel credentials stored as Kubernetes Secret
- [ ] Health check configured (readiness probe)
- [ ] Tunnel routes `*.podland.app` to Traefik

**Implementation:**
```yaml
# infra/k3s/cloudflared.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: cloudflared
---
apiVersion: v1
kind: Secret
metadata:
  name: cloudflared-credentials
  namespace: cloudflared
type: Opaque
stringData:
  # Generate tunnel token via: cloudflared tunnel login && cloudflared tunnel create podland-cluster
  tunnel.json: |
    {
      "AccountTag": "YOUR_ACCOUNT_TAG",
      "TunnelID": "YOUR_TUNNEL_ID",
      "TunnelSecret": "YOUR_TUNNEL_SECRET"
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudflared
  namespace: cloudflared
  labels:
    app: cloudflared
spec:
  replicas: 2  # 2 replicas for HA (not load balancing)
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
        image: cloudflare/cloudflared:2025.1.0
        args:
        - tunnel
        - --config
        - /etc/cloudflared/config/config.yaml
        - run
        - --no-autoupdate  # Disable auto-update for version control
        env:
        - name: TUNNEL_CREDENTIAL_FILE
          value: /etc/cloudflared/tunnel/tunnel.json
        readinessProbe:
          httpGet:
            path: /ready
            port: 2000
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        livenessProbe:
          httpGet:
            path: /health
            port: 2000
          initialDelaySeconds: 30
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
        volumeMounts:
        - name: credentials
          mountPath: /etc/cloudflared/tunnel
          readOnly: true
        - name: config
          mountPath: /etc/cloudflared/config
          readOnly: true
      volumes:
      - name: credentials
        secret:
          secretName: cloudflared-credentials
      - name: config
        configMap:
          name: cloudflared-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudflared-config
  namespace: cloudflared
data:
  config.yaml: |
    tunnel: podland-cluster
    credentials-file: /etc/cloudflared/tunnel/tunnel.json
    
    # Route all *.podland.app traffic to Traefik ingress
    ingress:
    - hostname: "*.podland.app"
      service: http://traefik.podland.svc.cluster.local:80
      originRequest:
        noTLSVerify: true  # Traefik uses self-signed Origin CA
    
    # Default 404 for non-matching routes
    - service: http_status:404
    
    # Logging
    loglevel: info
    metrics: "0.0.0.0:2000"
---
apiVersion: v1
kind: Service
metadata:
  name: cloudflared-metrics
  namespace: cloudflared
  labels:
    app: cloudflared
spec:
  selector:
    app: cloudflared
  ports:
  - name: metrics
    port: 2000
    targetPort: 2000
  type: ClusterIP
```

**Setup Instructions (one-time):**
```bash
# 1. Install cloudflared CLI
brew install cloudflared  # macOS
# or download from https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/

# 2. Authenticate with Cloudflare
cloudflared tunnel login

# 3. Create tunnel
cloudflared tunnel create podland-cluster

# 4. Extract credentials from ~/.cloudflared/<TUNNEL_ID>.json
# 5. Create Kubernetes secret with credentials
kubectl create secret generic cloudflared-credentials \
  --from-file=tunnel.json=~/.cloudflared/<TUNNEL_ID>.json \
  --namespace=cloudflared

# 6. Deploy
kubectl apply -f infra/k3s/cloudflared.yaml
```

**Files to create:**
- `infra/k3s/cloudflared.yaml`

---

### Week 2: SSL + Traefik Configuration

#### Task 2.1: Cloudflare Origin CA Certificate
**Estimate:** 2 hours
**Acceptance Criteria:**
- [ ] Origin CA wildcard certificate generated for `*.podland.app`
- [ ] Certificate installed as Kubernetes TLS secret
- [ ] Traefik configured to use Origin CA secret
- [ ] HTTPS working end-to-end

**Implementation:**
```bash
# 1. Generate Origin CA Certificate (one-time)
# Go to Cloudflare Dashboard → SSL/TLS → Origin Server → Create Certificate
# Select:
#   - Hostnames: *.podland.app, podland.app
#   - Key type: RSA (2048)
#   - Certificate Authority: Cloudflare Origin
#   - Valid for: 15 years

# 2. Download certificate and key
# Save as cert.pem and key.pem

# 3. Create Kubernetes TLS secret
kubectl create secret tls origin-ca-cert \
  --cert=cert.pem \
  --key=key.pem \
  --namespace=podland

# 4. Verify secret
kubectl get secret origin-ca-cert -n podland -o yaml
```

**Files to create:**
- `infra/k3s/origin-ca-secret.yaml` (optional, if not using kubectl CLI)

---

#### Task 2.2: Traefik HTTPS Configuration
**Estimate:** 3 hours
**Acceptance Criteria:**
- [ ] Traefik IngressRoute configured for wildcard subdomains
- [ ] TLS termination with Origin CA certificate
- [ ] Automatic HTTP → HTTPS redirection
- [ ] Cloudflare trusted IPs configured (for real IP forwarding)

**Implementation:**
```yaml
# infra/k3s/traefik-https.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: vm-ingress
  namespace: podland
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: websecure,web
    traefik.ingress.kubernetes.io/router.tls: "true"
    traefik.ingress.kubernetes.io/router.tls.certresolver: origin-ca
spec:
  tls:
  - hosts:
    - "*.podland.app"
    - podland.app
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
  - host: podland.app
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: traefik
            port: 80
---
# Traefik Middleware for HTTPS redirection
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: redirect-https
  namespace: podland
spec:
  redirectScheme:
    scheme: https
    permanent: true
---
# Traefik ServerTransport for Cloudflare trusted IPs
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: cloudflare-transport
  namespace: podland
spec:
  serverName: traefik.podland.svc
  insecureSkipVerify: true  # Origin CA is self-signed
---
# Traefik configuration to trust Cloudflare IPs
# Add to traefik-config.yaml or ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: traefik-config
  namespace: podland
data:
  traefik.yaml: |
    entryPoints:
      web:
        address: ":80"
        http:
          redirections:
            entryPoint:
              to: websecure
              scheme: https
              permanent: true
      websecure:
        address: ":443"
        http:
          tls:
            certResolver: origin-ca
    
    providers:
      kubernetesCRD:
        allowCrossNamespace: true
      kubernetesIngress:
        allowExternalNameServices: true
    
    # Trust Cloudflare IPs for real client IP
    entryPoints:
      websecure:
        forwardedHeaders:
          trustedIPs:
          - 173.245.48.0/20
          - 103.21.244.0/22
          - 103.22.200.0/22
          - 103.31.4.0/22
          - 141.101.64.0/18
          - 108.162.192.0/18
          - 190.93.240.0/20
          - 188.114.96.0/20
          - 197.234.240.0/22
          - 198.41.128.0/17
          - 162.158.0.0/15
          - 104.16.0.0/13
          - 104.24.0.0/14
          - 172.64.0.0/13
          - 131.0.72.0/22
```

**Files to modify:**
- `infra/k3s/traefik-config.yaml` (extend with HTTPS config)

**Files to create:**
- `infra/k3s/traefik-https.yaml`

---

#### Task 2.3: Domain Service Layer
**Estimate:** 3 hours
**Acceptance Criteria:**
- [ ] DomainService created for domain operations
- [ ] GetDomainsByUserID returns all user's domains
- [ ] DeleteDomain cascades to DNS and database
- [ ] Error handling for domain operations

**Implementation:**
```go
// apps/backend/internal/domain/service.go
package domain

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/podland/backend/internal/cloudflare"
    "github.com/podland/backend/internal/database"
)

type DomainService struct {
    dnsManager *cloudflare.DNSManager
    db         *database.DB
}

func NewDomainService(dnsManager *cloudflare.DNSManager, db *database.DB) *DomainService {
    return &DomainService{
        dnsManager: dnsManager,
        db:         db,
    }
}

// Domain represents a domain mapping
type Domain struct {
    ID         string `json:"id"`
    VMID       string `json:"vm_id"`
    UserID     string `json:"user_id"`
    Subdomain  string `json:"subdomain"`
    Domain     string `json:"domain"`
    Status     string `json:"status"` // pending, active, error
    CreatedAt  string `json:"created_at"`
}

// GetDomainsByUserID returns all domains for a user
func (s *DomainService) GetDomainsByUserID(ctx context.Context, userID string) ([]*Domain, error) {
    query := `
        SELECT id, vm_id, domain, domain_status as status, created_at
        FROM vms
        WHERE user_id = $1 AND domain IS NOT NULL
        ORDER BY created_at DESC
    `
    
    rows, err := s.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var domains []*Domain
    for rows.Next() {
        var d Domain
        var subdomain string
        // Extract subdomain from domain (vm-name.podland.app → vm-name)
        if err := rows.Scan(&d.ID, &d.VMID, &d.Domain, &d.Status, &d.CreatedAt); err != nil {
            return nil, err
        }
        // Parse subdomain
        d.Subdomain = parseSubdomain(d.Domain)
        d.UserID = userID
        domains = append(domains, &d)
    }
    
    return domains, nil
}

// DeleteDomain deletes a domain and its DNS record
func (s *DomainService) DeleteDomain(ctx context.Context, domainID, userID string) error {
    // Get domain
    vm, err := s.db.GetVMByID(domainID)
    if err != nil {
        return fmt.Errorf("domain not found")
    }
    
    // Check ownership
    if vm.UserID != userID {
        return fmt.Errorf("unauthorized")
    }
    
    if vm.Domain == "" {
        return fmt.Errorf("domain not assigned")
    }
    
    // Delete DNS record
    dnsRecord, err := s.dnsManager.GetRecordByName(ctx, vm.Domain)
    if err == nil && dnsRecord != nil {
        if err := s.dnsManager.DeleteRecord(ctx, dnsRecord.ID); err != nil {
            return fmt.Errorf("failed to delete DNS record: %w", err)
        }
    }
    
    // Update database
    _, err = s.db.ExecContext(ctx, `
        UPDATE vms 
        SET domain = NULL, domain_status = NULL, updated_at = NOW()
        WHERE id = $1
    `, domainID)
    
    if err != nil {
        return fmt.Errorf("failed to update database: %w", err)
    }
    
    return nil
}

// Helper: parse subdomain from full domain
func parseSubdomain(domain string) string {
    // vm-name.podland.app → vm-name
    const suffix = ".podland.app"
    if len(domain) > len(suffix) && domain[len(domain)-len(suffix):] == suffix {
        return domain[:len(domain)-len(suffix)]
    }
    return domain
}
```

**Files to create:**
- `apps/backend/internal/domain/service.go`
- `apps/backend/handler/domain_handler.go` (next task)

---

#### Task 2.4: Domain API Handler
**Estimate:** 2 hours
**Acceptance Criteria:**
- [ ] GET /api/domains returns user's domains
- [ ] DELETE /api/domains/:id deletes domain
- [ ] Auth middleware validates user session
- [ ] Error responses follow API conventions

**Implementation:**
```go
// apps/backend/handler/domain_handler.go
package handler

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/podland/backend/internal/domain"
)

type DomainHandler struct {
    domainService *domain.DomainService
}

func NewDomainHandler(domainService *domain.DomainService) *DomainHandler {
    return &DomainHandler{
        domainService: domainService,
    }
}

// GetDomains returns all domains for the authenticated user
func (h *DomainHandler) GetDomains(w http.ResponseWriter, r *http.Request) {
    user := getUserFromContext(r.Context())
    
    domains, err := h.domainService.GetDomainsByUserID(r.Context(), user.ID)
    if err != nil {
        http.Error(w, "Failed to fetch domains", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "domains": domains,
    })
}

// DeleteDomain deletes a domain by ID
func (h *DomainHandler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    domainID := vars["id"]
    
    user := getUserFromContext(r.Context())
    
    err := h.domainService.DeleteDomain(r.Context(), domainID, user.ID)
    if err != nil {
        if err.Error() == "domain not found" || err.Error() == "unauthorized" {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}
```

**Router setup:**
```go
// apps/backend/cmd/main.go (extend router)
r.HandleFunc("/api/domains", domainHandler.GetDomains).Methods("GET")
r.HandleFunc("/api/domains/{id}", domainHandler.DeleteDomain).Methods("DELETE")
```

**Files to create:**
- `apps/backend/handler/domain_handler.go`
- `apps/backend/cmd/main.go` (modify to add routes)

---

### Week 3: Frontend + Testing

#### Task 3.1: Domain List UI Component
**Estimate:** 3 hours
**Acceptance Criteria:**
- [ ] DomainList component displays user's domains
- [ ] Status badges (pending/active/error)
- [ ] Delete button with confirmation
- [ ] Empty state when no domains
- [ ] Responsive design (mobile/desktop)

**Implementation:**
```tsx
// apps/frontend/src/components/domains/DomainList.tsx
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../lib/api'
import { Trash2, ExternalLink, Clock, AlertCircle, CheckCircle } from 'lucide-react'

interface Domain {
  id: string
  vm_id: string
  subdomain: string
  domain: string
  status: 'pending' | 'active' | 'error'
  created_at: string
}

export function DomainList() {
  const queryClient = useQueryClient()
  
  const { data: domains, isLoading } = useQuery<{ domains: Domain[] }>({
    queryKey: ['domains'],
    queryFn: () => api.get('/api/domains'),
  })
  
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/api/domains/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['domains'] })
    },
  })
  
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="w-4 h-4 text-yellow-500" />
      case 'active':
        return <CheckCircle className="w-4 h-4 text-green-500" />
      case 'error':
        return <AlertCircle className="w-4 h-4 text-red-500" />
    }
  }
  
  if (isLoading) {
    return <div className="animate-pulse space-y-4">Loading domains...</div>
  }
  
  if (!domains || domains.domains.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 dark:text-gray-400">
          No domains yet. Create a VM to get started.
        </p>
      </div>
    )
  }
  
  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {domains.domains.map((domain) => (
          <div
            key={domain.id}
            className="bg-white dark:bg-gray-800 rounded-lg shadow p-4"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  {getStatusIcon(domain.status)}
                  <span className="text-sm font-medium capitalize">
                    {domain.status}
                  </span>
                </div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                  {domain.domain}
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  VM: {domain.vm_id}
                </p>
              </div>
              
              <div className="flex items-center gap-2">
                <a
                  href={`https://${domain.domain}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                  title="Open domain"
                >
                  <ExternalLink className="w-4 h-4" />
                </a>
                <button
                  onClick={() => {
                    if (confirm(`Delete domain ${domain.domain}?`)) {
                      deleteMutation.mutate(domain.id)
                    }
                  }}
                  disabled={deleteMutation.isPending}
                  className="p-2 text-red-400 hover:text-red-600 disabled:opacity-50"
                  title="Delete domain"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
            
            <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
              <p className="text-xs text-gray-500 dark:text-gray-400">
                Created {new Date(domain.created_at).toLocaleDateString()}
              </p>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
```

**Files to create:**
- `apps/frontend/src/components/domains/DomainList.tsx`
- `apps/frontend/src/routes/domains.tsx` (new route)

---

#### Task 3.2: Domain Route Page
**Estimate:** 2 hours
**Acceptance Criteria:**
- [ ] /domains route accessible from dashboard nav
- [ ] Page title and header
- [ ] DomainList component integrated
- [ ] Back button to dashboard

**Implementation:**
```tsx
// apps/frontend/src/routes/domains.tsx
import { createFileRoute, Link } from '@tanstack/react-router'
import { DomainList } from '../components/domains/DomainList'
import { ArrowLeft } from 'lucide-react'

export const Route = createFileRoute('/domains')({
  component: DomainsPage,
})

function DomainsPage() {
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          <Link
            to="/dashboard"
            className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              My Domains
            </h1>
            <p className="text-gray-500 dark:text-gray-400 mt-1">
              Manage your VM domains and subdomains
            </p>
          </div>
        </div>
        
        {/* Domain List */}
        <DomainList />
      </div>
    </div>
  )
}
```

**Files to create:**
- `apps/frontend/src/routes/domains.tsx`

---

#### Task 3.3: VM Detail Domain Display
**Estimate:** 2 hours
**Acceptance Criteria:**
- [ ] VM detail page shows domain (if assigned)
- [ ] Domain status badge visible
- [ ] Click to open domain in new tab
- [ ] Loading state during DNS propagation

**Implementation:**
```tsx
// apps/frontend/src/routes/dashboard/-vms/$id.tsx (extend existing)
import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { api } from '../../../lib/api'
import { ExternalLink, Clock, CheckCircle, AlertCircle } from 'lucide-react'

// ... existing code ...

function VMDetailPage({ vmId }: { vmId: string }) {
  const { data: vm, isLoading } = useQuery({
    queryKey: ['vm', vmId],
    queryFn: () => api.get(`/api/vms/${vmId}`),
  })
  
  if (isLoading) {
    return <div>Loading VM...</div>
  }
  
  if (!vm) {
    return <div>VM not found</div>
  }
  
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="w-4 h-4 text-yellow-500" />
      case 'active':
        return <CheckCircle className="w-4 h-4 text-green-500" />
      case 'error':
        return <AlertCircle className="w-4 h-4 text-red-500" />
    }
  }
  
  return (
    <div className="space-y-6">
      {/* ... existing VM info ... */}
      
      {/* Domain Section */}
      {vm.domain && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Domain
          </h2>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              {getStatusIcon(vm.domain_status || 'pending')}
              <span className="text-sm capitalize">
                {vm.domain_status || 'pending'}
              </span>
            </div>
            <a
              href={`https://${vm.domain}`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 text-blue-600 hover:text-blue-700 dark:text-blue-400"
            >
              <span className="font-mono">{vm.domain}</span>
              <ExternalLink className="w-4 h-4" />
            </a>
          </div>
          {vm.domain_status === 'pending' && (
            <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
              DNS propagation in progress. This may take up to 5 minutes.
            </p>
          )}
        </div>
      )}
    </div>
  )
}
```

**Files to modify:**
- `apps/frontend/src/routes/dashboard/-vms/$id.tsx` (extend with domain display)

---

#### Task 3.4: Integration Testing
**Estimate:** 4 hours
**Acceptance Criteria:**
- [ ] Test VM creation with domain assignment
- [ ] Test DNS propagation polling
- [ ] Test domain deletion
- [ ] Test domain list API
- [ ] Test HTTPS access to VM

**Implementation:**
```bash
# apps/backend/tests/domain_integration_test.go
package tests

import (
    "context"
    "net/http"
    "testing"
    "time"
    
    "github.com/podland/backend/internal/cloudflare"
    "github.com/podland/backend/internal/database"
    "github.com/podland/backend/internal/domain"
)

func TestDomainCreationAndDeletion(t *testing.T) {
    // Setup
    db := database.NewTestDB()
    dnsManager := cloudflare.NewDNSManager(testAPIToken, testZoneID)
    domainService := domain.NewDomainService(dnsManager, db)
    
    ctx := context.Background()
    
    // Create test user and VM
    user := db.CreateTestUser()
    vm := db.CreateTestVM(user.ID)
    
    // Test domain assignment
    subdomain := "test-vm"
    domain := subdomain + ".podland.app"
    
    _, err := dnsManager.CreateCNAME(ctx, domain, "tunnel.podland.app")
    if err != nil {
        t.Fatalf("Failed to create DNS record: %v", err)
    }
    
    // Verify DNS record exists
    record, err := dnsManager.GetRecordByName(ctx, domain)
    if err != nil || record == nil {
        t.Fatalf("DNS record not found")
    }
    
    // Test domain deletion
    err = dnsManager.DeleteRecord(ctx, record.ID)
    if err != nil {
        t.Fatalf("Failed to delete DNS record: %v", err)
    }
    
    // Verify DNS record deleted
    _, err = dnsManager.GetRecordByName(ctx, domain)
    if err == nil {
        t.Fatalf("Expected error, got nil")
    }
}

func TestDNSPropagationPolling(t *testing.T) {
    // Test polling logic with mock DNS manager
    // ...
}
```

**Files to create:**
- `apps/backend/tests/domain_integration_test.go`

---

## Testing Checklist

### Backend Tests
- [ ] DNS manager CRUD operations
- [ ] DNS poller timeout handling
- [ ] Domain service GetDomainsByUserID
- [ ] Domain service DeleteDomain (with ownership check)
- [ ] Domain handler API endpoints
- [ ] Integration test: VM creation → DNS → deletion

### Frontend Tests
- [ ] DomainList component renders
- [ ] Status badges display correctly
- [ ] Delete confirmation works
- [ ] Domain route accessible
- [ ] VM detail shows domain

### Manual Testing
- [ ] Create VM → verify DNS record in Cloudflare dashboard
- [ ] Access https://vm-name.podland.app → verify 200 OK
- [ ] Check SSL certificate validity
- [ ] Delete VM → verify DNS record removed
- [ ] Test domain list page UI
- [ ] Test on mobile and desktop

---

## Deployment Checklist

### Pre-deployment
- [ ] Cloudflare API token created (DNS:Edit, Zone:Read)
- [ ] Cloudflare tunnel created (`cloudflared tunnel create`)
- [ ] Origin CA certificate generated
- [ ] Kubernetes secrets created (tunnel credentials, Origin CA)

### Deployment
- [ ] Run database migrations
- [ ] Deploy cloudflared (`kubectl apply -f infra/k3s/cloudflared.yaml`)
- [ ] Deploy Traefik HTTPS config
- [ ] Deploy backend with new env vars
- [ ] Deploy frontend

### Post-deployment
- [ ] Test VM creation end-to-end
- [ ] Verify DNS propagation
- [ ] Test HTTPS access
- [ ] Test domain deletion
- [ ] Monitor cloudflared logs

---

## Environment Variables

**Backend:**
```bash
# Cloudflare
CLOUDFLARE_API_TOKEN=your_api_token
CLOUDFLARE_ZONE_ID=your_zone_id
CLOUDFLARE_TUNNEL_ID=your_tunnel_id

# Domain
BASE_DOMAIN=podland.app
```

**Frontend:**
```bash
# No new env vars needed
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Cloudflare API rate limiting | Exponential backoff (built into SDK), max 10 req/min |
| DNS propagation delay | Show "pending" status, poll every 10s, 5 min timeout |
| Tunnel connection issues | Health check, k8s auto-restart, 2 replicas for HA |
| SSL certificate expiry | 15-year validity, calendar reminder for renewal |
| Domain collisions | Append user suffix (vm-name-user123) |

---

## Files Summary

**New Files:**
- `apps/backend/internal/cloudflare/dns.go`
- `apps/backend/internal/cloudflare/dns_test.go`
- `apps/backend/internal/domain/poller.go`
- `apps/backend/internal/domain/service.go`
- `apps/backend/handler/domain_handler.go`
- `apps/backend/tests/domain_integration_test.go`
- `apps/backend/migrations/0003_add_domain_to_vms.sql`
- `infra/k3s/cloudflared.yaml`
- `infra/k3s/traefik-https.yaml`
- `apps/frontend/src/components/domains/DomainList.tsx`
- `apps/frontend/src/routes/domains.tsx`

**Modified Files:**
- `apps/backend/handler/vm_handler.go`
- `apps/backend/internal/database/vm_repository.go`
- `apps/backend/cmd/main.go`
- `apps/backend/internal/database/database.go` (add domain methods)
- `infra/k3s/traefik-config.yaml`
- `apps/frontend/src/routes/dashboard/-vms/$id.tsx`

---

## Estimated Effort

| Week | Tasks | Hours |
|------|-------|-------|
| Week 1 | DNS Manager, Poller, VM Integration, DB Schema, Tunnel | 13 hours |
| Week 2 | Origin CA, Traefik HTTPS, Domain Service, API Handler | 10 hours |
| Week 3 | Frontend UI, Domain Route, VM Detail, Testing | 11 hours |
| **Total** | | **34 hours** |

**Duration:** 3 weeks (part-time)

---

## Next Steps

1. Review and approve this plan
2. Set up Cloudflare credentials (API token, tunnel, Origin CA)
3. Execute tasks in order (Week 1 → Week 2 → Week 3)
4. Run integration tests
5. Deploy to k3s cluster
6. Manual testing and validation

---

*Plan created: 2026-03-29*
*Ready for execution*

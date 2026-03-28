# Phase 3 Research: Networking

**Phase:** 3 of 5
**Goal:** VMs are accessible via HTTPS with automatic domain and tunnel setup
**Research Date:** 2026-03-29
**Status:** Complete — ready for planning

---

## Research Area 1: Cloudflare Go SDK for DNS Record CRUD

### Key Findings

**SDK Version:** Use `github.com/cloudflare/cloudflare-go/v6` (latest stable as of 2025)

**Authentication Methods:**
```go
// Option 1: API Token (recommended)
client := cloudflare.NewClient(
    option.WithAPIToken("YOUR_API_TOKEN"),
)

// Option 2: API Key + Email (legacy)
client := cloudflare.NewClient(
    option.WithAPIKey("YOUR_API_KEY"),
    option.WithEmail("your@email.com"),
)

// Option 3: From environment (CLOUDFLARE_API_TOKEN)
client := cloudflare.NewClient()
```

### DNS Record CRUD Operations

**Create DNS Record (CNAME):**
```go
package dns

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

func (m *DNSManager) CreateCNAME(ctx context.Context, subdomain, target string, proxied bool) (*dns.Record, error) {
    record, err := m.client.DNS.Records.New(ctx, dns.RecordNewParams{
        ZoneID: cloudflare.F(m.zoneID),
        Type:   cloudflare.F(dns.RecordTypeCname),
        Name:   cloudflare.F(subdomain),
        Content: cloudflare.F(target),
        Proxied: cloudflare.F(proxied),
        TTL:     cloudflare.F[int](1), // Auto TTL
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create CNAME record: %w", err)
    }
    return record, nil
}
```

**List DNS Records with Filters:**
```go
func (m *DNSManager) ListRecords(ctx context.Context, name string) ([]dns.Record, error) {
    var allRecords []dns.Record

    // Auto-paging through all results
    iter := m.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
        ZoneID: cloudflare.F(m.zoneID),
        Name:   cloudflare.F(name),
        Type:   cloudflare.F(dns.RecordTypeCname),
    })

    for iter.Next() {
        allRecords = append(allRecords, iter.Current())
    }

    if err := iter.Err(); err != nil {
        return nil, fmt.Errorf("failed to list records: %w", err)
    }

    return allRecords, nil
}
```

**Update DNS Record:**
```go
func (m *DNSManager) UpdateRecord(ctx context.Context, recordID, content string) (*dns.Record, error) {
    record, err := m.client.DNS.Records.Edit(ctx, dns.RecordEditParams{
        ZoneID:   cloudflare.F(m.zoneID),
        RecordID: cloudflare.F(recordID),
        Content:  cloudflare.F(content),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to update record: %w", err)
    }
    return record, nil
}
```

**Delete DNS Record:**
```go
func (m *DNSManager) DeleteRecord(ctx context.Context, recordID string) error {
    _, err := m.client.DNS.Records.Delete(ctx, dns.RecordDeleteParams{
        ZoneID:   cloudflare.F(m.zoneID),
        RecordID: cloudflare.F(recordID),
    })
    if err != nil {
        return fmt.Errorf("failed to delete record: %w", err)
    }
    return nil
}
```

### Error Handling Patterns

```go
import "errors"

func handleCloudflareError(err error) error {
    var apierr *cloudflare.Error
    if errors.As(err, &apierr) {
        // Log request/response for debugging
        log.Printf("Request: %s", apierr.DumpRequest(true))
        log.Printf("Response: %s", apierr.DumpResponse(true))

        // Check error type
        switch {
        case apierr.Type() == cloudflare.ErrorTypeRatelimit:
            return fmt.Errorf("rate limited: %w", err)
        case apierr.Type() == cloudflare.ErrorTypeAuthentication:
            return fmt.Errorf("authentication failed: %w", err)
        case apierr.Type() == cloudflare.ErrorTypeAuthorization:
            return fmt.Errorf("authorization failed: %w", err)
        case apierr.Type() == cloudflare.ErrorTypeNotFound:
            return fmt.Errorf("resource not found: %w", err)
        }
    }
    return err
}
```

### Rate Limiting Configuration

```go
// Configure client with retry policy
client := cloudflare.NewClient(
    option.WithAPIToken("YOUR_TOKEN"),
    option.WithMaxRetries(3),              // Default is 2
    option.WithRequestTimeout(30*time.Second),
)

// Per-request retry override
record, err := client.DNS.Records.New(
    ctx,
    dns.RecordNewParams{...},
    option.WithMaxRetries(5), // Override for this request
)
```

**Default Retry Behavior:**
- Retries on: 408, 409, 429, and >=500 errors
- Default retries: 2
- Uses exponential backoff internally

---

## Research Area 2: cloudflared Kubernetes Deployment

### Key Findings

**Deployment Model:** Single deployment for entire cluster (as per Phase 3 decisions)

**Resource Requirements:**
- RAM: ~50MB per instance
- CPU: Minimal (~0.1 cores)
- Free tier limit: 5 tunnels per account

### Complete Deployment YAML

```yaml
# infra/k3s/cloudflared-deployment.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: podland
---
apiVersion: v1
kind: Secret
metadata:
  name: tunnel-token
  namespace: podland
type: Opaque
stringData:
  token: <YOUR_TUNNEL_TOKEN_FROM_CLOUDFLARE_DASHBOARD>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudflared
  namespace: podland
  labels:
    app: cloudflared
spec:
  replicas: 2  # HA - multiple replicas for failover only
  selector:
    matchLabels:
      app: cloudflared
  template:
    metadata:
      labels:
        app: cloudflared
    spec:
      # Required for ICMP traffic (ping, traceroute)
      securityContext:
        sysctls:
          - name: net.ipv4.ping_group_range
            value: "65532 65532"
      containers:
      - name: cloudflared
        image: cloudflare/cloudflared:latest
        args:
        - tunnel
        - --no-autoupdate
        - --loglevel
        - info
        - --metrics
        - 0.0.0.0:2000
        - run
        env:
        - name: TUNNEL_TOKEN
          valueFrom:
            secretKeyRef:
              name: tunnel-token
              key: token
        ports:
        - containerPort: 2000
          name: metrics
        livenessProbe:
          httpGet:
            path: /ready
            port: 2000
          failureThreshold: 1
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 2000
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "50Mi"
            cpu: "100m"
          limits:
            memory: "100Mi"
            cpu: "200m"
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
              - ALL
---
apiVersion: v1
kind: Service
metadata:
  name: cloudflared-metrics
  namespace: podland
spec:
  selector:
    app: cloudflared
  ports:
  - port: 2000
    targetPort: 2000
    name: metrics
```

### Tunnel Configuration (Dashboard Setup)

**Step 1: Create Tunnel**
```bash
# Install cloudflared locally for setup
brew install cloudflared  # macOS
# or
curl -L --output cloudflared https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
chmod +x cloudflared

# Create tunnel
./cloudflared tunnel create --name podland-cluster-tunnel
```

**Step 2: Get Tunnel Token**
```bash
# Extract token from credentials file
cat ~/.cloudflared/podland-cluster-tunnel.json
```

**Step 3: Configure Public Hostname (Dashboard)**
1. Go to Zero Trust Dashboard → Networks → Tunnels
2. Select `podland-cluster-tunnel`
3. Click "Add Public Hostname"
4. Configure:
   - **Subdomain:** `*` (wildcard)
   - **Domain:** `podland.app`
   - **Service:** `http://traefik.podland.svc.cluster.local:80`
   - **TLS:** Enable "No TLS Verify" (if using self-signed Origin CA)

### Best Practices

**DO:**
- Use 2 replicas for HA (failover only, not load balancing)
- Set `--no-autoupdate` for controlled versioning
- Use liveness probe on `/ready` endpoint
- Store tunnel token in Kubernetes Secret
- Set resource limits

**DON'T:**
- Don't use autoscaling (downscaling breaks connections)
- Don't run more than 5 tunnels (free tier limit)
- Don't mount credentials as files (use env vars)

---

## Research Area 3: Cloudflare Origin CA Certificate

### Key Findings

**Certificate Properties:**
- Validity: 1-15 years (recommend 15 years for "set and forget")
- Wildcard support: Yes (`*.podland.app`)
- SANs: Up to 200 per certificate
- Cost: Free
- Renewal: Manual (but 15-year validity)

### Generation Steps (Dashboard)

1. Go to Cloudflare Dashboard → SSL/TLS → Origin Server
2. Click "Create Certificate"
3. Choose:
   - **Method:** "Create a certificate"
   - **Hostnames:** `*.podland.app`, `podland.app`
   - **Validity:** 15 years (maximum)
4. Click "Create"
5. Download certificate and private key

### Kubernetes TLS Secret

```yaml
# infra/k3s/origin-ca-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: origin-ca-cert
  namespace: podland
type: kubernetes.io/tls
stringData:
  tls.crt: |
    -----BEGIN CERTIFICATE-----
    (Paste Origin CA wildcard certificate here)
    -----END CERTIFICATE-----
  tls.key: |
    -----BEGIN PRIVATE KEY-----
    (Paste private key here)
    -----END PRIVATE KEY-----
```

### Automation with cert-manager + origin-ca-issuer

**Install Origin CA Issuer:**
```yaml
# 1. Install CRDs
kubectl apply -f https://github.com/cloudflare/origin-ca-issuer/releases/latest/download/crds.yaml

# 2. Install RBAC
kubectl apply -f https://github.com/cloudflare/origin-ca-issuer/releases/latest/download/rbac.yaml

# 3. Install Controller
kubectl apply -f https://github.com/cloudflare/origin-ca-issuer/releases/latest/download/manifests.yaml
```

**Configure API Token Secret:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudflare-api-token
  namespace: podland
type: Opaque
stringData:
  api-token: YOUR_CLOUDFLARE_API_TOKEN
```

**Create OriginIssuer:**
```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: cloudflare-origin-issuer
  namespace: podland
spec:
  ca:
    secretName: origin-ca-issuer
---
apiVersion: origin.cert-manager.io/v1alpha1
kind: OriginIssuer
metadata:
  name: cloudflare-origin-issuer
  namespace: podland
spec:
  auth:
    serviceKeyRef:
      name: cloudflare-api-token
      key: api-token
```

**Request Wildcard Certificate:**
```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wildcard-cert
  namespace: podland
spec:
  secretName: origin-ca-cert
  issuerRef:
    name: cloudflare-origin-issuer
    kind: OriginIssuer
  dnsNames:
  - "*.podland.app"
  - "podland.app"
  duration: 36000h  # ~15 years
  renewBefore: 720h # Renew 30 days before expiry
```

### Manual Installation (Without cert-manager)

```bash
# Create secret from certificate files
kubectl create secret tls origin-ca-cert \
  --cert=path/to/cert.pem \
  --key=path/to/key.pem \
  -n podland
```

---

## Research Area 4: Traefik + Cloudflare Tunnel Integration

### Key Findings

**Architecture:**
```
Internet → Cloudflare Edge → Cloudflare Tunnel → Traefik (k8s) → VM Services
```

**Traffic Flow:**
1. User requests `vm-name.podland.app`
2. Cloudflare DNS resolves to tunnel
3. Tunnel forwards to Traefik ingress
4. Traefik routes to VM service based on hostname

### Traefik Static Configuration

```yaml
# infra/k3s/traefik-config.yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: traefik
  namespace: podland
spec:
  values:
    ports:
      web:
        port: 8000
        expose: true
        exposedPort: 80
        protocol: http
      websecure:
        port: 8443
        expose: true
        exposedPort: 443
        protocol: https
        tls:
          enabled: true
          certResolver: ""  # Use Origin CA, not Let's Encrypt
    providers:
      kubernetesCRD:
        enabled: true
        allowCrossNamespace: true
      kubernetesIngress:
        enabled: true
    additionalArguments:
    - "--providers.kubernetesingress.ingressendpoint.publishedservice=podland/traefik"
    - "--entrypoints.web.http.redirections.entryPoint.to=websecure"
    - "--entrypoints.web.http.redirections.entryPoint.scheme=https"
```

### IngressRoute for Wildcard Domains

```yaml
# infra/k3s/vm-ingressroute.yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: vm-ingress
  namespace: podland
spec:
  entryPoints:
    - websecure
  routes:
  - match: HostRegexp(`{subdomain:[a-z0-9-]+}.podland.app`)
    kind: Rule
    services:
    - name: traefik
      port: 8000
      # Traefik will route based on hostname to appropriate backend
    middlewares:
    - name: strip-prefix
  tls:
    secretName: origin-ca-cert  # Reference the Origin CA secret
```

### VM-Specific IngressRoute (Dynamic)

```yaml
# Generated per VM by backend controller
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: vm-my-app
  namespace: podland
  labels:
    vm-id: "12345"
spec:
  entryPoints:
    - websecure
  routes:
  - match: Host(`my-app.podland.app`)
    kind: Rule
    services:
    - name: vm-my-app-service
      port: 8080
  tls:
    secretName: origin-ca-cert
```

### Middleware for Security Headers

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: security-headers
  namespace: podland
spec:
  headers:
    customFrameOptionsValue: "DENY"
    customRequestHeaders:
      X-Forwarded-Proto: "https"
    stsIncludeSubdomains: true
    stsPreload: true
    stsSeconds: 31536000
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: strip-prefix
  namespace: podland
spec:
  stripPrefix:
    prefixes:
      - /api
```

### Handling Real Client IPs

```yaml
# Traefik configuration for trusted IPs (Cloudflare)
additionalArguments:
- "--entrypoints.websecure.forwardedHeaders.trustedIPs=173.245.48.0/20,103.21.244.0/22,103.22.200.0/22,103.31.4.0/22,141.101.64.0/18,108.162.192.0/18,190.93.240.0/20,188.114.96.0/20,197.234.240.0/22,198.41.128.0/17,162.158.0.0/15,104.16.0.0/13,104.24.0.0/14,172.64.0.0/13,131.0.72.0/22"
- "--entrypoints.websecure.proxyProtocol.trustedIPs=173.245.48.0/20,103.21.244.0/22"
```

---

## Research Area 5: DNS Propagation Handling

### Key Findings

**Cloudflare DNS Propagation:**
- **Internal propagation:** 30 seconds to a few minutes
- **Global propagation:** Depends on TTL (typically 5 minutes with auto TTL)
- **Cloudflare network:** Near-instant (< 60 seconds)

### DNS Record Status Fields

Cloudflare DNS API response includes:
```json
{
  "result": {
    "id": "record_id",
    "type": "CNAME",
    "name": "vm-name.podland.app",
    "content": "tunnel.podland.app",
    "proxied": true,
    "ttl": 1,
    "created_on": "2026-03-29T10:00:00Z",
    "modified_on": "2026-03-29T10:00:00Z"
  }
}
```

**Note:** Cloudflare doesn't expose a "status" field. Records are active immediately in their network.

### Polling Strategy

```go
package dns

import (
    "context"
    "fmt"
    "time"

    "github.com/cloudflare/cloudflare-go/v6/dns"
)

type PollConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay time.Duration
    Multiplier float64
}

func DefaultPollConfig() PollConfig {
    return PollConfig{
        MaxAttempts:  30,        // 5 minutes total
        InitialDelay: 5 * time.Second,
        MaxDelay:     30 * time.Second,
        Multiplier:   1.5,
    }
}

func (m *DNSManager) WaitForDNSActive(ctx context.Context, subdomain string, cfg PollConfig) error {
    delay := cfg.InitialDelay

    for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            // Check if record exists and is proxied
            records, err := m.ListRecords(ctx, subdomain)
            if err != nil {
                return fmt.Errorf("failed to check DNS status: %w", err)
            }

            if len(records) > 0 {
                // Record exists - Cloudflare DNS is active
                // For proxied records, check Cloudflare status
                if records[0].Proxied != nil && *records[0].Proxied {
                    // Proxied record - typically active within 60s
                    return nil
                }
                return nil // Non-proxied records are immediate
            }

            // Exponential backoff with max delay
            delay = time.Duration(float64(delay) * cfg.Multiplier)
            if delay > cfg.MaxDelay {
                delay = cfg.MaxDelay
            }
        }
    }

    return fmt.Errorf("DNS propagation timeout after %d attempts", cfg.MaxAttempts)
}
```

### TTL Considerations

**Cloudflare Auto TTL Values:**
- Proxied records: Automatic (Cloudflare controls)
- Non-proxied records:
  - Free plan: Minimum 300s (5 minutes)
  - Paid plans: Minimum 60s (1 minute)

**Recommendation:** Use `ttl: 1` (Auto) for all Phase 3 DNS records

### Application-Level Handling

```go
// Backend service pattern for VM creation with DNS
func (s *VMService) CreateVM(ctx context.Context, input VMCreateInput) (*VM, error) {
    // 1. Create VM in k8s
    vm, err := s.k8sManager.CreateDeployment(ctx, input)
    if err != nil {
        return nil, err
    }

    // 2. Generate subdomain
    subdomain := sanitizeSubdomain(input.Name)
    fqdn := fmt.Sprintf("%s.podland.app", subdomain)

    // 3. Create DNS record
    _, err = s.dnsManager.CreateCNAME(ctx, fqdn, "tunnel.podland.app", true)
    if err != nil {
        // Rollback VM
        s.k8sManager.DeleteDeployment(ctx, vm.ID)
        return nil, err
    }

    // 4. Set initial status
    vm.Domain = fqdn
    vm.DomainStatus = DomainStatusPending

    // 5. Start background polling
    go func() {
        bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
        defer cancel()

        err := s.dnsManager.WaitForDNSActive(bgCtx, fqdn, DefaultPollConfig())
        if err != nil {
            log.Printf("DNS propagation failed for %s: %v", fqdn, err)
            s.updateDomainStatus(vm.ID, DomainStatusFailed)
        } else {
            s.updateDomainStatus(vm.ID, DomainStatusActive)
        }
    }()

    return vm, nil
}
```

### User Experience Pattern

**Frontend Status Display:**
```tsx
// VM Domain Status Badge
function DomainStatusBadge({ status }: { status: DomainStatus }) {
  switch (status) {
    case 'pending':
      return (
        <Badge variant="warning">
          <Spinner size="sm" />
          DNS Propagating...
        </Badge>
      );
    case 'active':
      return (
        <Badge variant="success">
          <CheckIcon />
          Active
        </Badge>
      );
    case 'failed':
      return (
        <Badge variant="danger">
          DNS Setup Failed
        </Badge>
      );
  }
}
```

---

## Gotchas and Pitfalls

### Cloudflare API

1. **Rate Limits:**
   - Free tier: 100 requests/minute per API token
   - Use exponential backoff (SDK handles automatically)
   - Batch DNS operations when possible

2. **Zone ID Confusion:**
   - Zone ID ≠ Domain name
   - Get Zone ID from dashboard or API: `GET /zones?name=podland.app`
   - Store Zone ID in config (don't lookup every time)

3. **Proxied vs DNS-Only:**
   - `proxied: true` = Cloudflare proxy + SSL (orange cloud)
   - `proxied: false` = DNS only (gray cloud)
   - For tunnels: Always use `proxied: true`

### Cloudflare Tunnel

1. **Tunnel Token Security:**
   - Token = tunnel credentials
   - Store in Kubernetes Secret, never in code
   - Rotate tokens periodically

2. **Connection Limits:**
   - Free tier: 5 tunnels max
   - Each tunnel: Unlimited hostnames
   - Solution: Single tunnel for entire cluster

3. **Health Checks:**
   - Use `/ready` endpoint (not `/health`)
   - `/ready` checks Cloudflare connection status
   - Failure threshold = 1 for fast detection

### Origin CA Certificates

1. **Browser Trust:**
   - Origin CA certs are NOT trusted by browsers directly
   - MUST use Cloudflare proxy (orange cloud) for SSL
   - Direct access shows certificate warning

2. **Wildcard Limitations:**
   - Only covers one subdomain level: `*.podland.app`
   - Does NOT cover: `*.sub.podland.app`
   - Solution: Use multiple SANs if needed

3. **Renewal:**
   - No automatic renewal
   - 15-year validity = set and forget
   - Set calendar reminder for year 14

### Traefik Configuration

1. **Wildcard Host Matching:**
   - Use `HostRegexp()` for wildcard matching
   - Pattern: `HostRegexp(\`{subdomain:[a-z0-9-]+}.podland.app\`)`
   - Standard `Host()` doesn't support wildcards

2. **TLS Secret Namespace:**
   - Secret must be in same namespace as IngressRoute
   - Or use Traefik middleware for cross-namespace

3. **HTTPS Redirection:**
   - Configure at entrypoint level
   - Don't use middleware for global redirect (performance)

### DNS Propagation

1. **False Negatives:**
   - DNS may appear inactive from your location
   - Cloudflare network has it active immediately
   - Poll from multiple locations if needed

2. **TTL Caching:**
   - Old DNS records cached by ISPs
   - Use low TTL during development
   - Production: Auto TTL (Cloudflare managed)

3. **Propagation ≠ Accessibility:**
   - DNS active ≠ Tunnel ready
   - Check both DNS and tunnel health
   - Tunnel health: `/ready` endpoint

---

## Implementation Checklist

### Prerequisites
- [ ] Cloudflare account with domain (`podland.app`)
- [ ] API Token with `Zone:DNS:Edit` permission
- [ ] Zone ID for domain
- [ ] Cloudflare Tunnel created (dashboard)
- [ ] Origin CA certificate generated

### Infrastructure
- [ ] Origin CA TLS secret created in k8s
- [ ] cloudflared deployment configured
- [ ] Tunnel token stored in secret
- [ ] Traefik configured with Origin CA cert
- [ ] IngressRoute for wildcard domains

### Backend
- [ ] DNS manager service implemented
- [ ] VM creation triggers DNS record creation
- [ ] Background worker for DNS polling
- [ ] Domain status tracking (pending/active/failed)
- [ ] VM deletion triggers DNS cleanup

### Frontend
- [ ] Domain status badge component
- [ ] Domain list view (DOMAIN-05)
- [ ] Domain delete action (DOMAIN-06)
- [ ] Loading states for DNS propagation

---

## References

### Official Documentation

1. **Cloudflare Go SDK:**
   - GitHub: https://github.com/cloudflare/cloudflare-go
   - GoDoc: https://pkg.go.dev/github.com/cloudflare/cloudflare-go/v6
   - API Docs: https://developers.cloudflare.com/api/

2. **Cloudflare Tunnel:**
   - Kubernetes Guide: https://developers.cloudflare.com/tunnel/deployment-guides/kubernetes/
   - Configuration: https://developers.cloudflare.com/tunnel/configuration/
   - Troubleshooting: https://developers.cloudflare.com/tunnel/troubleshooting/

3. **Cloudflare DNS:**
   - DNS Records: https://developers.cloudflare.com/dns/manage-dns-records/
   - TTL Behavior: https://developers.cloudflare.com/dns/manage-dns-records/reference/ttl-behavior/
   - API Reference: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-create-dns-record

4. **Origin CA:**
   - Documentation: https://developers.cloudflare.com/ssl/origin-configuration/origin-ca/
   - Origin CA Issuer: https://github.com/cloudflare/origin-ca-issuer

5. **Traefik:**
   - IngressRoute: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/
   - TLS Configuration: https://doc.traefik.io/traefik/https/tls/
   - Middlewares: https://doc.traefik.io/traefik/middlewares/overview/

### Community Resources

1. **Traefik + Cloudflare Tunnel:**
   - Community Discussion: https://community.traefik.io/t/cloudflare-origin-certificate-co-existing-with-lets-encrypt/29445
   - Reddit: https://www.reddit.com/r/homelab/comments/1lxmax2/possible_to_use_cloudflare_tunnel_traefik/

2. **DNS Propagation:**
   - Cloudflare DNS Speed: https://apipark.com/technews/QGslvOeL.html
   - Server Fault Discussion: https://serverfault.com/questions/125371/how-long-does-it-take-for-dns-records-to-propagate

---

*Research completed: 2026-03-29*
*All 5 research areas covered with implementation patterns*
*Ready for planning phase*

# Pitfalls Research: PaaS Platform Mistakes

## Common PaaS Pitfalls & Prevention Strategies

### 1. Resource Exhaustion (Critical)

**Pitfall:** One user consumes all server resources, starving others.

**Real Example:**
> "A student deployed a crypto miner that maxed out CPU. 50 other VMs became unresponsive."

**Warning Signs:**
- No ResourceQuota defined
- Limits > Requests (overcommit too aggressive)
- No monitoring alerts

**Prevention:**
```yaml
# Always set both requests AND limits
resources:
  requests:
    cpu: "0.5"
    memory: 1Gi
  limits:
    cpu: "1"
    memory: 2Gi
```

**Phase:** Phase 2 (VM Management + Quotas)

---

### 2. Noisy Neighbor Problem

**Pitfall:** Shared storage/network causes performance interference.

**Real Example:**
> "VM A doing heavy disk I/O. VM B (same node) experiences 10x latency."

**Warning Signs:**
- Single storage class for all users
- No network policies
- No I/O limits

**Prevention:**
- Use local-lvm for persistent data (better isolation)
- NetworkPolicy per namespace
- Consider storage IOPS limits (if supported)

**Phase:** Phase 2 (Resource Quotas)

---

### 3. Security: Container Escape

**Pitfall:** User breaks out of container, gains host access.

**Real Example:**
> "Student used privileged container to access host filesystem. Compromised all VMs."

**Warning Signs:**
- `privileged: true` in pod spec
- Host path mounts
- Running as root

**Prevention:**
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
```

**Phase:** Phase 2 (VM Management) - NON-NEGOTIABLE

---

### 4. Idle Detection False Positives

**Pitfall:** Active VMs incorrectly marked as idle and deleted.

**Real Example:**
> "Background worker VM had no HTTP traffic but was processing jobs. Got auto-deleted."

**Warning Signs:**
- Single signal for idle detection (HTTP only)
- No grace period
- No user notification

**Prevention:**
- Combined signals: HTTP + Process + Login
- 48-hour minimum idle window
- 24-hour warning before deletion
- User can "pin" VM to prevent auto-delete

**Phase:** Phase 5 (Idle Detection)

---

### 5. Cloudflare Rate Limiting

**Pitfall:** API rate limits hit during bulk operations.

**Real Example:**
> "Bulk VM creation (50 users) hit Cloudflare API limit. Half the DNS records failed."

**Warning Signs:**
- No rate limiting in client
- No retry logic
- Bulk operations without batching

**Prevention:**
```go
// Use exponential backoff
backoff := time.Second
for i := 0; i < maxRetries; i++ {
    err := cloudflareAPI.CreateDNSRecord(...)
    if err == nil { break }
    if isRateLimitError(err) {
        time.Sleep(backoff)
        backoff *= 2
        continue
    }
}
```

**Phase:** Phase 3 (Domain Management)

---

### 6. Database Connection Exhaustion

**Pitfall:** Too many concurrent connections crash PostgreSQL.

**Real Example:**
> "Each VM opened 5 DB connections. 100 VMs = 500 connections. PostgreSQL crashed."

**Warning Signs:**
- Direct DB access from VMs
- No connection pooling
- No max connection limit

**Prevention:**
- PgBouncer for connection pooling
- Max 100 connections from platform
- VMs access DB via API only (not direct)

**Phase:** Phase 2 (Backend Architecture)

---

### 7. Log Storage Explosion

**Pitfall:** Logs consume all disk space.

**Real Example:**
> "Debug logging left enabled. 500 VMs × 1GB/day = disk full in 3 days."

**Warning Signs:**
- No log rotation
- No retention policy
- Debug level in production

**Prevention:**
```yaml
# Loki retention config
retention:
  period: 30d
  max_size: 50GB
  max_chunks: 100000
```

**Phase:** Phase 4 (Monitoring)

---

### 8. GitHub OAuth Misconfiguration

**Pitfall:** Wrong OAuth settings allow unauthorized access.

**Real Example:**
> "Callback URL not validated. Attacker redirected OAuth to their server."

**Warning Signs:**
- Hardcoded callback URLs
- No state parameter validation
- No PKCE

**Prevention:**
```go
// Always validate state parameter
state := generateSecureRandomString()
session.Set("oauth_state", state)
// On callback
if callbackState != session.Get("oauth_state") {
    return Error("Invalid state")
}
```

**Phase:** Phase 1 (Authentication) - CRITICAL

---

### 9. NIM Validation Edge Cases

**Pitfall:** Valid students rejected due to NIM format changes.

**Real Example:**
> "New batch used different NIM format. All legitimate students blocked."

**Warning Signs:**
- Hardcoded validation regex
- No manual override
- No logging of rejections

**Prevention:**
```go
// Flexible validation + manual review queue
func validateNIM(nim string) (valid bool, needsReview bool) {
    if matchesCurrentFormat(nim) {
        return true, false
    }
    if matchesOldFormat(nim) {
        return true, false  // Legacy support
    }
    return false, true  // Queue for manual review
}
```

**Phase:** Phase 1 (Authentication)

---

### 10. Single Point of Failure

**Pitfall:** One component failure takes down entire platform.

**Real Example:**
> "Prometheus disk full. Alerting stopped. Nobody noticed outage for 6 hours."

**Warning Signs:**
- No redundancy
- No health checks
- No alerting on alerting system

**Prevention:**
- Health checks on all critical components
- Alert on disk usage > 80%
- Separate monitoring storage

**Phase:** Phase 4 (Monitoring)

---

## Pitfall Prevention Summary by Phase

| Phase | Pitfalls to Prevent |
|-------|---------------------|
| **Phase 1 (Auth)** | OAuth misconfiguration, NIM validation bugs |
| **Phase 2 (VM + Quotas)** | Resource exhaustion, container escape, noisy neighbor, DB connection exhaustion |
| **Phase 3 (Domain)** | Cloudflare rate limiting, DNS propagation issues |
| **Phase 4 (Monitoring)** | Log explosion, SPOF, missing alerts |
| **Phase 5 (Idle)** | False positive idle detection, data loss |

## Red Flags During Development

🚩 **Stop and fix immediately if you see:**
- Containers running as root
- No resource limits defined
- Direct database access from user VMs
- OAuth without state parameter
- No audit logging for admin actions
- Secrets in environment variables (use Kubernetes Secrets)
- No backup strategy for PostgreSQL

## Testing Checklist

Before each phase launch:

- [ ] Load test: 100 concurrent VMs
- [ ] Chaos test: Kill random pods
- [ ] Security scan: Trivy/Grype on images
- [ ] Penetration test: Container escape attempts
- [ ] Backup/restore test: PostgreSQL recovery

## Recommended Tools

| Purpose | Tool | Why |
|---------|------|-----|
| Security Scanning | Trivy | Fast, comprehensive, CI integration |
| Policy Enforcement | OPA Gatekeeper | Kubernetes-native policy engine |
| Backup | Velero | Kubernetes backup + restore |
| Monitoring | Prometheus + Grafana | Industry standard |
| Log Management | Loki | Lightweight, cost-effective |
| Chaos Testing | Chaos Mesh | Kubernetes-native chaos engineering |

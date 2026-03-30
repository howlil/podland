# Phase 5 Research: Admin + Polish

**Phase:** 5 of 5
**Status:** Research complete — ready for planning

---

## Research Questions Answered

### 1. Admin Panel RBAC Best Practices

**Question:** What's the best pattern for admin-only route protection in Go?

**Answer:** **Router-level middleware groups** — cleanest separation, applied once.

**Findings:**
- Kubernetes RBAC best practices emphasize "least privilege" and "clear separation"
- Go/chi router patterns support route groups with middleware
- Middleware applied at group level affects all child routes
- Single point of control for admin authorization

**Implementation Pattern:**
```go
r.Route("/api/admin", func(r chi.Router) {
    r.Use(middleware.AdminOnly)
    // All routes here require superadmin
})
```

**Source:** [Kubernetes RBAC best practices](https://kubernetes.io/docs/concepts/security/rbac-good-practices/)

---

### 2. Idle Container Detection with Prometheus

**Question:** How to accurately detect idle containers using Prometheus metrics?

**Answer:** **Combined signals** — No HTTP + no CPU + no SSH for 48h.

**Findings:**
- cAdvisor provides `container_network_receive_bytes_total` for network I/O
- cAdvisor provides `container_cpu_usage_seconds_total` for CPU activity
- Promtail can capture SSH login events from auth logs
- Combined criteria eliminates false positives

**Prometheus Queries:**
```promql
# No network activity for 48h
sum_over_time(container_network_receive_bytes_total{vm_id="abc"}[48h]) == 0

# No CPU activity for 48h
avg_over_time(container_cpu_usage_seconds_total{vm_id="abc"}[48h]) < 0.01
```

**Source:** [Monitoring Kubernetes Pods with Prometheus](https://medium.com/cloud-native-daily/monitoring-kubernetes-pods-resource-usage-with-prometheus-and-grafana-c17848febadc)

---

### 3. Email Provider Comparison (SendGrid vs AWS SES vs Mailgun)

**Question:** Which email provider is best for Podland (500 users, <1000 emails/month)?

**Answer:** **SendGrid** — Free tier covers volume, zero engineering overhead.

**Comparison:**

| Provider | Free Tier | Paid Tier | Deliverability | Setup Time | Best For |
|----------|-----------|-----------|----------------|------------|----------|
| **SendGrid** | 100/day | $15/50k | ✅ Excellent | <1 hour | **Small projects** |
| AWS SES | None | $0.10/1000 | Good | 2-4 days | High volume (>100k/month) |
| Mailgun | 5/hour | $35/50k | Good | <1 hour | Testing |
| Self-hosted | Free | Free | ❌ Poor | High | Never |

**Key Research Insight:**
- SendGrid free tier: 100 emails/day × 30 days = 3,000 emails/month
- Podland needs: ~500 users × 2 emails/month (idle warnings) = 1,000 emails/month
- **SendGrid covers 3x required volume for free**
- AWS SES requires ~$12,000 engineering time to build equivalent infrastructure

**Source:** [SendGrid vs AWS SES comparison](https://xmit.sh/versus/amazon-ses-vs-sendgrid)

---

### 4. Load Testing Tools for Kubernetes (k6 vs Locust)

**Question:** Which load testing tool is best for Kubernetes API testing?

**Answer:** **k6** — Resource efficient, Kubernetes-native, Prometheus integration.

**Comparison:**

| Feature | k6 | Locust |
|---------|-----|--------|
| Language | JavaScript | Python |
| Concurrency Model | Event-loop (efficient) | Per-user threads (heavy) |
| Resource Usage | 100k VUs @ ~500MB RAM | 10k VUs @ ~1GB RAM |
| Kubernetes Support | ✅ Operator available | Helm chart |
| Prometheus Integration | ✅ Native | Via exporters |
| CI/CD Integration | ✅ GitHub Actions native | Good |
| Learning Curve | Low (JS familiar) | Medium (Python needed) |

**Research Insight:**
- k6 uses 10x less RAM for same load (event-loop vs threads)
- k6 has Kubernetes Operator for native k8s integration
- k6 metrics automatically exported to Prometheus
- **k6 is "default choice for K8s teams"**

**Source:** [Load Testing in Kubernetes: Top Tools & Best Practices](https://testkube.io/blog/load-testing-in-kubernetes-tools-and-best-practices)

---

## Component Versions (2026)

| Component | Version | Notes |
|-----------|---------|-------|
| SendGrid | Latest API v3 | Free tier available |
| k6 | v0.50+ | JavaScript test scripts |
| Go | 1.25+ | Backend language |
| React | 18+ | Frontend framework |
| PostgreSQL | 15 | Database |

---

## Resource Requirements

### Email System (SendGrid)

| Component | Cost | Volume |
|-----------|------|--------|
| SendGrid Free | $0/month | 100 emails/day |
| Podland Usage | ~1,000/month | Well within free tier |

**Setup Requirements:**
- SendGrid account (free)
- API key (5 minutes to generate)
- Domain verification (optional for free tier)

---

### Idle Detection System

| Component | CPU | RAM | Storage |
|-----------|-----|-----|---------|
| Idle Detector (goroutine) | 10m | 32Mi | None |
| Prometheus queries | Included in Phase 4 | | |

**Resource Overhead:** Negligible (single goroutine, hourly execution)

---

### Load Testing (k6)

| Test Scenario | VUs | Duration | Resource Usage |
|---------------|-----|----------|----------------|
| Critical paths | 100 | 5 minutes | ~200MB RAM, 100m CPU |
| Stress test | 200 | 5 minutes | ~400MB RAM, 200m CPU |

**Execution Frequency:** On-demand (before launch), not continuous

---

### Backup System (pg_dump)

| Component | CPU | RAM | Storage |
|-----------|-----|-----|---------|
| Backup cron job | 100m (during backup) | 128Mi | 10Gi (7-day retention) |
| S3 storage | $0.023/GB/month | ~1GB backups | ~$0.02/month |

**Backup Schedule:** Daily at 3 AM (low-traffic time)

---

## Implementation Patterns

### Admin Authorization

```go
// middleware/admin.go
func AdminOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value("user_id")
        if userID == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        user, err := userRepo.GetByID(r.Context(), userID)
        if err != nil || user.Role != "superadmin" {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

### Idle Detection Logic

```go
// internal/idle/detector.go
func (d *Detector) findIdleVMs() ([]VM, error) {
    query := `
        SELECT v.id, v.user_id, v.name
        FROM vms v
        WHERE v.status = 'running'
        AND v.is_pinned = false
        AND v.id NOT IN (
            SELECT vm_id FROM vm_metrics_48h
            WHERE network_bytes > 0
            OR cpu_seconds > 0.01
            OR ssh_logins > 0
        )
    `
    
    return d.vmRepo.Query(query)
}
```

---

### Email Retry Logic

```go
// internal/email/service.go
func (s *EmailService) SendWithRetry(email Email, maxRetries int) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        if err := s.Send(email); err == nil {
            return nil
        } else {
            lastErr = err
        }
        
        // Exponential backoff: 1m, 4m, 9m...
        backoff := time.Duration(i*i) * time.Minute
        time.Sleep(backoff)
    }
    
    return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

---

### Load Test k6 Script

```javascript
// tests/load/critical-paths.js
export const options = {
  vus: 100,
  duration: '5m',
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate==0'],
  },
};

export default function () {
  // Auth flow
  const token = login();
  
  // VM lifecycle
  const vmID = createVM(token);
  startVM(token, vmID);
  stopVM(token, vmID);
  deleteVM(token, vmID);
}
```

---

## API Design

### Admin Endpoints

```
GET    /api/admin/users              # List all users
PATCH  /api/admin/users/{id}/role    # Change user role
POST   /api/admin/users/{id}/ban     # Ban user
GET    /api/admin/health             # System health dashboard
GET    /api/admin/audit-log          # Audit log
```

### Idle VM Endpoints

```
POST   /api/vms/{id}/pin             # Pin VM
DELETE /api/vms/{id}/pin             # Unpin VM
GET    /api/vms/{id}/idle-status     # Check idle status
```

### Email Endpoints

```
POST   /api/notifications/resend     # Resend notification as email
```

---

## Database Schema

```sql
-- Audit logs (new table)
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  action VARCHAR(255) NOT NULL,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Add pinned field to vms
ALTER TABLE vms ADD COLUMN is_pinned BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE vms ADD COLUMN idle_warned_at TIMESTAMP;

-- Indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_vms_is_pinned ON vms(is_pinned);
```

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| False positive idle detection | Low | High | Combined criteria (all 3 signals) |
| Email delivery failures | Medium | Medium | Retry with backoff, in-app fallback |
| Load test failures | Medium | High | Strict criteria, optimize before launch |
| Backup corruption | Low | Critical | Test restore procedure |
| Admin panel XSS | Low | High | Input sanitization, CSP headers |

---

## Next Steps

**Ready for planning:** All research questions answered.

**Planning should produce:**
1. Week-by-week task breakdown
2. Database migration scripts
3. Backend handler implementations
4. Frontend admin panel components
5. Load testing scripts
6. Backup automation scripts

---

*Research completed: 2026-03-29*
*All technical questions answered — ready for planning*

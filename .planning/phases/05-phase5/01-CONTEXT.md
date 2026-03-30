# Phase 5 Context: Admin + Polish

**Phase:** 5 of 5
**Goal:** Superadmin can manage platform, API is complete, ready for launch
**Requirements:** 10 (ADMIN-01 through ADMIN-05, IDLE-01 through IDLE-04, VM-08)
**Status:** Context gathered — all decisions locked, ready for research and planning

---

## Prior Context (From Phase 1-4)

### Architecture Decisions (Locked)

| Decision | Value | Rationale |
|----------|-------|-----------|
| Orchestration | k3s | Cloud native, production-ready, 500MB RAM footprint |
| Backend | Go 1.25+ | Excellent k3s ecosystem, type-safe, performant |
| Frontend | React + TanStack Router + Tailwind v4 | Modern DX, type-safe routing |
| Database | PostgreSQL 15 | Battle-tested, JSONB flexibility |
| VM Abstraction | Docker containers with resource limits | Shared resource model, fast startup |
| Ingress | Traefik | Already configured for wildcard subdomain routing |
| Monitoring | Prometheus + Loki + Grafana | Phase 4 implementation |
| Notifications | In-app + PostgreSQL table | Phase 4 implementation |

### Existing Infrastructure (Phase 1-4)

```
podland/
├── apps/
│   ├── backend/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/          # Auth, VM, metrics, logs, notifications
│   │   │   ├── middleware/       # CORS, CSRF, Auth
│   │   │   ├── usecase/          # Business logic
│   │   │   ├── repository/       # Data access
│   │   │   ├── entity/           # Domain models (User, VM, Notification)
│   │   │   └── auth/             # JWT, OAuth, sessions
│   │   └── migrations/           # 001-004 (users, VMs, quotas, notifications)
│   └── frontend/
│       ├── src/
│       │   ├── routes/
│       │   │   ├── dashboard/    # VMs, profile, observability
│       │   │   └── auth/
│       │   └── components/
│       └── public/
├── infra/
│   └── k3s/
│       ├── namespace.yaml
│       ├── postgres.yaml
│       ├── backend.yaml
│       ├── frontend.yaml
│       └── monitoring/           # Phase 4: Prometheus, Loki, Grafana
└── packages/
    └── types/
```

### Reusable Patterns

- **Backend:** Clean architecture (handler → usecase → repository)
- **Frontend:** TanStack Router with file-based routing
- **k3s:** Namespace per user, Deployment per VM, PVC for storage
- **Notifications:** PostgreSQL table with in-app polling (Phase 4)
- **Role-based access:** User roles (internal/external/superadmin) in database

### Database Schema (Existing)

```sql
-- users table (Phase 1)
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL CHECK (role IN ('internal', 'external', 'superadmin')),
  nim VARCHAR(20),
  -- ...
);

-- vms table (Phase 2)
CREATE TABLE vms (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  status VARCHAR(20) NOT NULL,
  -- ...
);

-- notifications table (Phase 4)
CREATE TABLE notifications (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  is_read BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW(),
  -- ...
);
```

---

## Phase 5 Decisions (Implementation Details)

All decisions below are locked. Planning agent uses these to create actionable tasks.

### 1. Admin Panel Access Control

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Route Protection** | Router-level group with middleware | Admin routes in `/api/admin/*` group, `AdminOnly` middleware applied once — clean separation, no code duplication |
| **Superadmin Assignment** | Database flag (manual SQL) | `UPDATE users SET role = 'superadmin' WHERE email = '...'` — most secure, one-time setup, no code dependencies |
| **UI Visibility** | Conditional rendering | `{user.role === 'superadmin' && <AdminNav />}` — clear UX, non-admins never see admin features |
| **Audit Logging** | Middleware auto-logging | Every admin request logged automatically (user ID, action, IP, timestamp) — consistent, no developer error |

**Implementation Pattern:**
```go
// cmd/main.go
r.Route("/api/admin", func(r chi.Router) {
    r.Use(middleware.AdminOnly)    // Applied to all admin routes
    r.Use(middleware.AuditLogger)  // Auto-log all actions
    
    r.Get("/users", adminHandler.ListUsers)
    r.Patch("/users/{id}/role", adminHandler.ChangeRole)
    r.Post("/users/{id}/ban", adminHandler.BanUser)
    r.Get("/health", adminHandler.SystemHealth)
    r.Get("/audit-log", adminHandler.AuditLog)
})

// middleware/audit.go
func AuditLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value("user_id")
        action := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
        ip := r.RemoteAddr
        
        auditRepo.Create(ctx, &AuditLog{
            UserID: userID,
            Action: action,
            IP:     ip,
        })
        
        next.ServeHTTP(w, r)
    })
}
```

**Database Schema Additions:**
```sql
-- Audit logs table
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  action VARCHAR(255) NOT NULL,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

---

### 2. Idle VM Detection Strategy

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Idle Criteria** | Combined (all signals must be true) | No HTTP traffic + no CPU activity + no SSH login for 48h — accurate, no false positives |
| **Detection Frequency** | Hourly cron job | Run idle detection every hour — predictable schedule, good balance, simple implementation |
| **Warning Period** | Fixed 24h notice | Standard grace period, sufficient for students to respond, simple to implement |
| **Pin Exemption** | Role-based limits | External: 1 pin, Internal: 3 pins — matches existing quota model, fair allocation |

**Implementation Pattern:**
```go
// internal/idle/detector.go
type Detector struct {
    prometheusURL string
    vmRepo        repository.VMRepository
    notificationRepo repository.NotificationRepository
}

func (d *Detector) Run() {
    // Query Prometheus for idle VMs
    // No HTTP traffic for 48h
    idleHTTP := `sum_over_time(container_network_receive_bytes_total[48h]) == 0`
    // No CPU activity for 48h
    idleCPU := `avg_over_time(container_cpu_usage_seconds_total[48h]) < 0.01`
    // No SSH login for 48h
    idleSSH := `count_over_time({job="promtail"} |= "Accepted" [48h]) == 0`
    
    // Find VMs matching all criteria
    idleVMs := d.findIdleVMs(idleHTTP, idleCPU, idleSSH)
    
    for _, vm := range idleVMs {
        // Check if pinned
        if vm.IsPinned {
            continue
        }
        
        // Check if already warned
        if d.isWarned(vm.ID) {
            // 24h passed, delete VM
            d.deleteVM(vm)
        } else {
            // Send warning notification
            d.sendWarning(vm)
        }
    }
}

// Check pin limit
func (d *Detector) CanPin(userID string, role string) bool {
    limit := 1
    if role == "internal" {
        limit = 3
    }
    
    pinnedCount := d.vmRepo.GetPinnedCount(userID)
    return pinnedCount < limit
}
```

**Prometheus Queries:**
```promql
# No HTTP traffic for 48h
sum_over_time(container_network_receive_bytes_total{vm_id="abc"}[48h]) == 0

# No CPU activity for 48h
avg_over_time(container_cpu_usage_seconds_total{vm_id="abc"}[48h]) < 0.01

# No SSH login for 48h (via Promtail logs)
count_over_time({vm_id="abc"} |= "Accepted" [48h]) == 0
```

**Database Schema Additions:**
```sql
-- Add pinned field to vms table
ALTER TABLE vms ADD COLUMN is_pinned BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE vms ADD COLUMN idle_warned_at TIMESTAMP;

CREATE INDEX idx_vms_is_pinned ON vms(is_pinned);
```

---

### 3. Email Notification System

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Email Provider** | SendGrid | Free tier (100 emails/day = 3,000/month), excellent deliverability, <1 hour setup, no engineering overhead |
| **Email Templates** | Multipart (HTML + text) | Email client chooses best format — Gmail shows HTML, Apple Mail may show text |
| **User Preferences** | System-determined only | Critical emails always sent (idle warnings, security alerts), no preferences UI needed |
| **Fallback Strategy** | Retry with backoff | 3 retries with 1m, 5m, 15m delays — handles transient failures, no message queue needed |

**Implementation Pattern:**
```go
// internal/email/service.go
type EmailService struct {
    apiKey    string
    fromEmail string
    client    *sendgrid.Client
}

func (s *EmailService) SendIdleWarning(user User, vm VM) error {
    data := EmailData{
        "UserName": user.Name,
        "VMName":   vm.Name,
        "VMID":     vm.ID,
        "DeleteAt": time.Now().Add(24 * time.Hour),
    }
    
    htmlBody := renderTemplate("idle_warning.html", data)
    textBody := renderTemplate("idle_warning.txt", data)
    
    msg := mail.NewV3Mail()
    msg.SetFrom(mail.NewEmail("Podland", s.fromEmail))
    msg.AddMailPersonalization(mail.NewPersonalization())
    msg.Personalizations[0].AddTos(mail.NewEmail(user.Name, user.Email))
    msg.SetSubject("Your VM will be deleted in 24 hours")
    msg.SetContent(mail.NewContent("text/html", htmlBody))
    msg.AddContent(mail.NewContent("text/plain", textBody))
    
    // Retry with backoff
    for i := 0; i < 3; i++ {
        if _, err := s.client.Send(msg); err == nil {
            return nil
        }
        time.Sleep(time.Duration(i*i) * time.Minute) // 1m, 4m, 9m
    }
    
    return errors.New("failed to send email after 3 retries")
}
```

**Email Templates:**
```html
<!-- templates/idle_warning.html -->
<html>
<body>
  <h1>VM Deletion Warning</h1>
  <p>Hi {{.UserName}},</p>
  <p>Your VM <strong>{{.VMName}}</strong> has been idle for 48 hours.</p>
  <p>It will be deleted on {{.DeleteAt}} unless you log in and use it.</p>
  <a href="https://podland.app/dashboard/vms/{{.VMID}}">View VM</a>
</body>
</html>
```

```txt
<!-- templates/idle_warning.txt -->
VM Deletion Warning

Hi {{.UserName}},

Your VM {{.VMName}} has been idle for 48 hours.
It will be deleted on {{.DeleteAt}} unless you log in and use it.

View VM: https://podland.app/dashboard/vms/{{.VMID}}
```

**Environment Variables:**
```bash
# Email configuration
SENDGRID_API_KEY=SG.xxxxx
SENDGRID_FROM_EMAIL=noreply@podland.app
```

---

### 4. Load Testing & Launch Readiness

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Load Testing Tool** | k6 | JavaScript-based, resource efficient (event-loop), Kubernetes operator, Prometheus integration |
| **Test Scenarios** | Critical paths only | Auth flow, VM lifecycle (create/start/stop/delete), metrics API, quota enforcement — covers 80% usage |
| **Success Criteria** | Strict (p95 < 500ms, 0% errors) | Launch-ready standard, motivates optimization, clear pass/fail |
| **Backup Strategy** | Automated daily pg_dump | Cron job dumps PostgreSQL to S3 — simple, reliable, fast restore |

**Implementation Pattern:**
```javascript
// tests/load/critical-paths.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 100,           // 100 concurrent users
  duration: '5m',     // 5 minutes
  thresholds: {
    http_req_duration: ['p(95)<500'], // p95 < 500ms
    http_req_failed: ['rate==0'],     // 0% errors
  },
};

export default function () {
  const baseURL = 'http://localhost:8080';
  
  // 1. Auth flow
  const loginRes = http.post(`${baseURL}/api/auth/login`);
  check(loginRes, {
    'login status is 200': (r) => r.status === 200,
  });
  
  const token = loginRes.json('access_token');
  const headers = { 'Authorization': `Bearer ${token}` };
  
  // 2. VM create
  const createRes = http.post(
    `${baseURL}/api/vms`,
    JSON.stringify({ name: 'load-test-vm', os: 'ubuntu-2204', tier: 'micro' }),
    { headers }
  );
  check(createRes, {
    'create VM status is 201': (r) => r.status === 201,
  });
  
  const vmID = createRes.json('id');
  
  // 3. VM start
  const startRes = http.post(
    `${baseURL}/api/vms/${vmID}/start`,
    null,
    { headers }
  );
  check(startRes, {
    'start VM status is 200': (r) => r.status === 200,
  });
  
  // 4. Metrics fetch
  const metricsRes = http.get(
    `${baseURL}/api/vms/${vmID}/metrics?range=24h`,
    { headers }
  );
  check(metricsRes, {
    'metrics status is 200': (r) => r.status === 200,
  });
  
  // 5. VM stop
  const stopRes = http.post(
    `${baseURL}/api/vms/${vmID}/stop`,
    null,
    { headers }
  );
  check(stopRes, {
    'stop VM status is 200': (r) => r.status === 200,
  });
  
  // 6. VM delete
  const deleteRes = http.del(
    `${baseURL}/api/vms/${vmID}`,
    null,
    { headers }
  );
  check(deleteRes, {
    'delete VM status is 200': (r) => r.status === 200,
  });
  
  sleep(1);
}
```

**Backup Script:**
```bash
#!/bin/bash
# /etc/cron.daily/pg-backup

set -e

# Configuration
DB_NAME="podland"
DB_USER="podland"
BACKUP_DIR="/backups"
S3_BUCKET="s3://podland-backups"
DATE=$(date +%Y%m%d-%H%M%S)

# Create backup
pg_dump "postgresql://${DB_USER}@localhost/${DB_NAME}" | gzip > "${BACKUP_DIR}/podland-${DATE}.sql.gz"

# Upload to S3
aws s3 cp "${BACKUP_DIR}/podland-${DATE}.sql.gz" "${S3_BUCKET}/"

# Clean up old backups (keep 7 days)
find "${BACKUP_DIR}" -name "podland-*.sql.gz" -mtime +7 -delete

echo "Backup completed: podland-${DATE}.sql.gz"
```

**Success Criteria Checklist:**
- [ ] p95 response time < 500ms (all endpoints)
- [ ] 0% error rate (no 5xx errors)
- [ ] All resources healthy (CPU < 80%, RAM < 80%)
- [ ] 100 concurrent users sustained for 5 minutes
- [ ] Database backup completed successfully
- [ ] Backup restore tested (dry run)

---

## Requirements (From ROADMAP.md)

| ID | Requirement | Success Criteria |
|----|-------------|------------------|
| ADMIN-01 | Superadmin can view list of all users | Admin panel shows user list with filters |
| ADMIN-02 | Superadmin can change user role | Role change persists, user sees new permissions |
| ADMIN-03 | Superadmin can ban/unban users | Banned user cannot sign in |
| ADMIN-04 | Superadmin can view system health dashboard | Dashboard shows cluster CPU/RAM/storage |
| ADMIN-05 | System logs all admin actions to audit log | Audit log shows who did what, when |
| IDLE-01 | System detects idle VMs (48h no activity) | Idle VMs identified hourly |
| IDLE-02 | System sends warning 24h before delete | User receives email + in-app notification |
| IDLE-03 | System auto-deletes idle VM after grace period | VM deleted if still idle after 24h |
| IDLE-04 | User can pin VM to prevent auto-delete | Pinned VMs excluded from idle detection |
| VM-08 | Quota enforcement verified at scale | Load test: 100 concurrent VMs, quotas enforced |

---

## Technical Milestones

- [ ] Admin panel UI (users, system health, audit log)
- [ ] Idle detection worker (Prometheus queries + logic)
- [ ] Email notification system (SendGrid integration)
- [ ] Auto-delete cron job (hourly idle detection)
- [ ] Pin VM feature (database + UI)
- [ ] Load testing (k6 scripts + execution)
- [ ] Backup automation (pg_dump + S3 upload)
- [ ] Security audit (Trivy scan, dependency check)

---

## Code Context (What Exists)

### Backend (`apps/backend/`)

**Extend these:**
- `internal/handler/` — Add admin handlers
- `internal/middleware/` — Add AdminOnly, AuditLogger
- `internal/entity/user.go` — Add ban/is_pinned fields
- `internal/repository/` — Add audit log, pin operations
- `cmd/main.go` — Wire up admin routes, idle detector

**New files to create:**
- `internal/handler/admin_handler.go` — Admin panel handlers
- `internal/middleware/audit.go` — Audit logging middleware
- `internal/idle/detector.go` — Idle VM detection logic
- `internal/email/service.go` — SendGrid email service
- `migrations/005_phase5_admin.sql` — Audit logs, pin fields

### Frontend (`apps/frontend/`)

**Extend these:**
- `components/layout/DashboardLayout.tsx` — Add admin nav (conditional)
- `routes/dashboard/-vms/$id.tsx` — Add pin button

**New files to create:**
- `routes/admin/index.tsx` — Admin dashboard
- `routes/admin/users.tsx` — User management
- `routes/admin/health.tsx` — System health
- `routes/admin/audit-log.tsx` — Audit log viewer
- `components/admin/UserTable.tsx` — User list with actions

### Infrastructure (`infra/k3s/`)

**New files to create:**
- `backups/namespace.yaml`
- `backups/backup-cronjob.yaml` — Daily pg_dump job
- `backups/backup-pvc.yaml` — Temporary storage

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| False positive idle detection | Low | High | Combined criteria (all signals must be true) |
| Email delivery failures | Medium | Medium | Retry with backoff, in-app fallback |
| Load test failures | Medium | High | Strict success criteria, optimize before launch |
| Backup corruption | Low | Critical | Test restore procedure, keep 7-day backups |
| Admin panel XSS/CSRF | Low | High | Same CSRF protection as main app, input sanitization |

---

## Deferred Ideas (For Future Phases)

| Idea | Deferred To | Reason |
|------|-------------|--------|
| User-configurable email preferences | Phase 5 (if time permits) or v2 | UI + database overhead, not critical for launch |
| Velero for full cluster backup | v2 | pg_dump sufficient for database-only backup |
| Soak testing (24h load) | v2 | Critical paths sufficient for student project |
| Tiered idle warnings (72h → 24h → 1h) | v2 | Single 24h warning sufficient |
| Pin justification + admin approval | v2 | Role-based limits automatic, no approval needed |
| Separate admin subdomain | v2 | Conditional rendering sufficient for single app |

---

## Next Steps

**For Planning Agent:**
1. Break down 10 requirements into implementation tasks
2. Estimate effort per task (hours)
3. Define acceptance criteria per task
4. Identify dependencies (admin panel before audit log, idle detector before email)

**Planning Output:** `02-PLAN.md` with week-by-week breakdown

---

*Context created: 2026-03-29*
*All 16 decisions locked — ready for planning*

# Phase 5 Planning Summary

**Date:** 2026-03-29
**Status:** ✅ Planning Complete — Ready for Execution

---

## What Was Done

### 1. Context Discussion ✅
**Duration:** ~1 hour
**Gray Areas Discussed:** 4 major areas, 16 decisions total

All decisions documented in `01-CONTEXT.md`:
- Admin Panel Access Control (4 decisions)
- Idle VM Detection Strategy (4 decisions)
- Email Notification System (4 decisions)
- Load Testing & Launch Readiness (4 decisions)

**Output:** `.planning/phases/05-phase5/01-CONTEXT.md`

---

### 2. Research Phase ✅
**File:** `.planning/phases/05-phase5/02-RESEARCH.md`

**Research Questions Answered:**
1. ✅ Admin panel RBAC best practices — Router-level middleware groups
2. ✅ Idle container detection with Prometheus — Combined signals (HTTP + CPU + SSH)
3. ✅ Email provider comparison — SendGrid wins (free tier, zero overhead)
4. ✅ Load testing tools for Kubernetes — k6 (resource efficient, K8s-native)

**Key Research Insights:**
- SendGrid free tier: 100 emails/day = 3,000/month (covers 500 users easily)
- k6 uses 10x less RAM than Locust for same load (event-loop vs threads)
- Combined idle criteria eliminates false positives
- Router-level middleware is cleanest separation for admin routes

---

### 3. Implementation Plan ✅
**File:** `.planning/phases/05-phase5/03-PLAN.md`

**Plan Summary:**
- **Duration:** 4 weeks (70 hours total)
- **Requirements:** 10 (ADMIN-01 through ADMIN-05, IDLE-01 through IDLE-04, VM-08)
- **Success Criteria:** 10 measurable outcomes
- **Technical Milestones:** 8 key deliverables

**Week Breakdown:**

| Week | Focus | Tasks | Hours |
|------|-------|-------|-------|
| 1 | Admin Panel Backend | 4 tasks | 18h |
| 2 | Idle Detection + Email | 4 tasks | 20h |
| 3 | Frontend Admin + Pin UI | 4 tasks | 16h |
| 4 | Load Testing + Backup | 4 tasks | 16h |

**Key Implementation Files:**

**Backend:**
```
apps/backend/internal/
├── middleware/admin.go           # Admin authorization
├── middleware/audit.go           # Audit logging
├── handler/admin_handler.go      # Admin panel handlers
├── idle/detector.go              # Idle VM detection
└── email/service.go              # SendGrid integration
```

**Frontend:**
```
apps/frontend/src/
├── routes/admin/
│   ├── index.tsx                 # Admin dashboard
│   ├── users.tsx                 # User management
│   ├── health.tsx                # System health
│   └── audit-log.tsx             # Audit log viewer
└── components/admin/
    └── UserTable.tsx             # User list with actions
```

**Infrastructure:**
```
infra/k3s/backups/
├── backup-cronjob.yaml           # Daily pg_dump
└── backup-pvc.yaml               # Backup storage
```

**Tests:**
```
tests/load/
└── critical-paths.js             # k6 load test script
```

---

## Key Decisions Summary

### Admin Panel
- **Route Protection:** Router-level group with `AdminOnly` middleware
- **Superadmin Assignment:** Database flag (manual SQL, one-time setup)
- **UI Visibility:** Conditional rendering (superadmin only)
- **Audit Logging:** Middleware auto-logging (asynchronous)

### Idle Detection
- **Criteria:** Combined (no HTTP + no CPU + no SSH for 48h)
- **Frequency:** Hourly cron job (single goroutine)
- **Warning:** Fixed 24h notice (single notification)
- **Pin Limits:** External: 1, Internal: 3 (role-based)

### Email System
- **Provider:** SendGrid (free tier, 100/day)
- **Templates:** Multipart HTML + text
- **Preferences:** System-determined (critical emails only)
- **Fallback:** Retry with backoff (3 retries, 1m/4m/9m)

### Load Testing + Backup
- **Tool:** k6 (JavaScript, efficient, K8s-native)
- **Scenarios:** Critical paths only (auth + VM lifecycle)
- **Success:** Strict (p95 < 500ms, 0% errors, 100 VUs)
- **Backup:** Automated daily pg_dump to S3

---

## Database Schema Changes

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

-- VM pinning (modify existing table)
ALTER TABLE vms ADD COLUMN is_pinned BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE vms ADD COLUMN idle_warned_at TIMESTAMP;

-- Indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_vms_is_pinned ON vms(is_pinned);
```

---

## Environment Variables Required

```bash
# Admin panel
# (no new env vars - uses existing JWT auth)

# Email system
SENDGRID_API_KEY=SG.xxxxx
SENDGRID_FROM_EMAIL=noreply@podland.app

# Idle detection
PROMETHEUS_URL=http://prometheus.monitoring.svc:9090

# Backup
S3_BUCKET=s3://podland-backups
AWS_ACCESS_KEY_ID=xxxxx
AWS_SECRET_ACCESS_KEY=xxxxx
```

---

## Next Steps

### Immediate (Execution)
1. **Start Week 1 tasks** — Create admin middleware
2. **Set up SendGrid account** — Generate API key (5 minutes)
3. **Create S3 bucket** — For backup storage

### Commands to Start
```bash
# 1. Create SendGrid account
# Visit: https://sendgrid.com/free
# Generate API key

# 2. Create S3 bucket
aws s3 mb s3://podland-backups

# 3. Start implementation
# Follow 03-PLAN.md week-by-week
```

### Verification
After implementation, verify with:
```bash
# Run load test
k6 run tests/load/critical-paths.js

# Test backup restore
pg_dump podland | gzip > test-restore.sql.gz
gunzip test-restore.sql.gz
psql postgres < test-restore.sql

# Verify admin panel
# Login as superadmin → access /admin → verify all features
```

---

## Files Created/Updated

| File | Status | Purpose |
|------|--------|---------|
| `.planning/phases/05-phase5/01-CONTEXT.md` | ✅ Created | 16 implementation decisions |
| `.planning/phases/05-phase5/02-RESEARCH.md` | ✅ Created | Technical research findings |
| `.planning/phases/05-phase5/03-PLAN.md` | ✅ Created | Week-by-week plan |
| `.planning/phases/05-phase5/04-SUMMARY.md` | ✅ Created | This summary |

---

## Ready for Execution ✅

**All planning artifacts complete:**
- ✅ CONTEXT.md — 16 decisions locked
- ✅ RESEARCH.md — Technical questions answered
- ✅ PLAN.md — 4-week implementation plan
- ✅ SUMMARY.md — Progress tracking

**Next command:** `/ez:execute-phase 5`

**Or start manually:**
```bash
# Week 1, Task 1.1: Admin Authorization Middleware
# Create apps/backend/internal/middleware/admin.go
```

---

*Planning completed: 2026-03-29*
*Ready for implementation — 70 hours estimated*

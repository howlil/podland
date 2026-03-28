# Phase 2 Planning Summary

**Date:** 2026-03-27
**Status:** ✅ Planning Complete — Ready for Execution

---

## What Was Done

### 1. Interactive Discussion Session ✅
**Duration:** ~2 hours
**Gray Areas Discussed:** 4 major areas, 12 decisions total

#### VM Image Strategy (4/4 decisions)
- ✅ Base Image: Build sendiri dari official Ubuntu/Debian
- ✅ Preinstall: Essentials only (git, curl, wget, vim, htop)
- ✅ Build & Hosting: GitHub Actions + Docker Hub
- ✅ Versioning: Semantic + monthly builds

#### VM Storage Strategy (4/4 decisions)
- ✅ Storage Type: Single root filesystem (`/` = 100% quota)
- ✅ Storage Class: local-lvm
- ✅ Backup: Snapshot on VM delete (7-day retention)
- ✅ OS Disk: Total includes OS (standard VPS model)

#### VM Network Access (4/4 decisions)
- ✅ SSH Access: SSH via Ingress
- ✅ SSH Key: System generates keypair (Ed25519)
- ✅ Port Exposure: All ports via Ingress
- ✅ Firewall: Default allow 22/80/443, user can add more

**Output:** `.planning/phases/2-discussion-results.md`

---

### 2. Context Documentation ✅
**File:** `.planning/phases/2-CONTEXT.md`

**Contents:**
- Prior context from Phase 1
- All 12 decisions with rationale
- Database schema additions
- Kubernetes manifest examples
- API endpoint specifications
- Deferred ideas (for Phase 5)
- Code context (what exists)
- Research questions

---

### 3. Research Phase ✅
**File:** `.planning/phases/2-RESEARCH.md`

**Research Questions Answered:**
1. ✅ k3s client-go patterns — Implementation pattern with dynamic client
2. ✅ Pre-built OS images — GitHub Actions + Docker Hub workflow
3. ✅ Ingress controller — Traefik configuration for wildcard + TCP
4. ✅ ResourceQuota + PVC — Storage limits with DB tracking
5. ✅ PodSecurityStandard — `restricted` level compliance
6. ✅ SSH key generation — Ed25519 via `golang.org/x/crypto/ssh`
7. ✅ VM configuration — cloud-init for SSH key injection

**Confidence:** High — All questions answered with implementation patterns

---

### 4. Implementation Plan ✅
**File:** `.planning/phases/2-PLAN.md`

**Plan Summary:**
- **Duration:** 4 weeks
- **Requirements:** 16 (VM-01 through VM-11, QUOTA-01 through QUOTA-05, API-01 through API-04)
- **Success Criteria:** 16 measurable outcomes
- **Technical Milestones:** 12 key deliverables

**Week Breakdown:**
- **Week 1:** Database + Backend Core
  - Task 1.1: Database Schema Migration (4h)
  - Task 1.2: Backend k8s Module (VMManager) (8h)
  - Task 1.3: Backend SSH Key Generation (2h)
  - Task 1.4: Backend Quota Enforcement (6h)

- **Week 2:** Backend API + Frontend Core
  - Task 2.1: Backend VM Handlers (8h)
  - Task 2.2: Frontend VM List View (4h)
  - Task 2.3: Frontend VM Detail Page (6h)
  - Task 2.4: Frontend Create VM Wizard (6h)

- **Week 3:** Infrastructure + Testing
  - Task 3.1: k3s Traefik Configuration (4h)
  - Task 3.2: GitHub Actions Workflow (2h)
  - Task 3.3: Integration Testing (4h)

- **Week 4:** Polish + Documentation
  - Task 4.1: Load Testing (2h)
  - Task 4.2: Documentation (2h)
  - Buffer time for unexpected issues

**Key Implementation Files:**
```
backend/
├── internal/k8s/vm_manager.go       # k3s client for VM lifecycle
├── internal/ssh/keygen.go           # SSH keypair generation
├── internal/database/quota.go       # Quota enforcement
├── internal/database/vms.go         # VM database operations
└── handlers/vms.go                  # VM API endpoints

frontend/
├── src/routes/dashboard/vms.tsx            # VM list page
├── src/routes/dashboard/vms/$id.tsx        # VM detail page
└── src/components/vm/CreateVMWizard.tsx    # Create VM form

infra/
├── k3s/traefik-config.yaml        # Traefik configuration
└── ../.github/workflows/build-vm-images.yml  # VM image builds

database/
└── migrations/002_phase2_vm_quota.sql  # Schema migration
```

---

## Key Decisions Summary

### Architecture
- **Orchestration:** k3s (confirmed from Phase 1)
- **Backend:** Go 1.25+ with client-go
- **Frontend:** React + TanStack Router
- **Database:** PostgreSQL 15

### VM Tiers (6 tiers, VPS-style)
| Tier | CPU | RAM | Storage | Available To |
|------|-----|-----|---------|--------------|
| nano | 0.25 | 512 MB | 5 GB | External |
| micro | 0.5 | 1 GB | 10 GB | Both |
| small | 1.0 | 2 GB | 20 GB | Internal |
| medium | 2.0 | 4 GB | 40 GB | Internal |
| large | 4.0 | 8 GB | 80 GB | Internal |
| xlarge | 4.0 | 8 GB | 100 GB | Internal (max) |

### Quota Limits
| Role | CPU | RAM | Storage | VMs |
|------|-----|-----|---------|-----|
| External | 0.5 | 1 GB | 10 GB | 2 |
| Internal | 4.0 | 8 GB | 100 GB | 5 |

### Security
- **PodSecurityStandard:** `restricted` level
- **Container Security:** Non-root, no privilege escalation, drop ALL capabilities
- **Network Policy:** Default deny (allow 22/80/443)
- **SSH Key:** Ed25519, shown once, not stored by backend

---

## Next Steps

### Immediate (Execution)
1. **Start Week 1 tasks** — Database migration + k8s module
2. **Test k3s connectivity** — Ensure backend can connect to k3s cluster
3. **Setup GitHub OAuth** — For testing (if not already done)

### Commands to Start
```bash
# 1. Run database migration
cd apps/backend
go run migrations/002_phase2_vm_quota.sql

# 2. Install k8s dependencies
go get k8s.io/api@v0.29.0
go get k8s.io/apimachinery@v0.29.0
go get k8s.io/client-go@v0.29.0

# 3. Start implementation
# Follow 2-PLAN.md week-by-week
```

### Verification
After implementation, verify with:
```bash
# Run load test
./scripts/load-test.sh

# Check all success criteria
# See 2-PLAN.md "Success Criteria" section
```

---

## Files Created/Updated

| File | Status | Purpose |
|------|--------|---------|
| `.planning/phases/2-CONTEXT.md` | ✅ Created | Implementation decisions |
| `.planning/phases/2-RESEARCH.md` | ✅ Created | Research findings |
| `.planning/phases/2-PLAN.md` | ✅ Created | Implementation plan |
| `.planning/phases/2-discussion-results.md` | ✅ Created | Discussion log |
| `.planning/STATE.md` | ✅ Updated | Phase progress tracking |

---

## Ready for Execution ✅

**All planning artifacts complete:**
- ✅ CONTEXT.md — Decisions locked
- ✅ RESEARCH.md — Questions answered
- ✅ PLAN.md — Tasks defined
- ✅ STATE.md — Progress tracked

**Next command:** `/ez:execute-phase 2`

---

*Planning completed: 2026-03-27*
*Ready for implementation*

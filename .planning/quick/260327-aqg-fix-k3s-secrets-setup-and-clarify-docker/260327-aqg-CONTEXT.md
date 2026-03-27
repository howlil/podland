# Quick Task 260327-aqg: Context

**Task:** Fix k3s secrets setup and clarify Docker Compose structure

**Date:** 2026-03-27

---

## Issues Identified

### 1. k3s Secrets Not Working ❌

**Problem:** SealedSecrets are placeholder values, not real encrypted secrets.

**Current State:**
- `secrets/postgres-sealedsecret.yaml` - Contains `AgBy8sF3......` (placeholder)
- `secrets/backend-sealedsecret.yaml` - Contains `AgBy8sF3...REPLACE_WITH_GENERATED_VALUE...`

**Why it fails:**
- SealedSecrets must be generated using `kubeseal` CLI with actual secret values
- Placeholder values won't decrypt in k3s cluster
- Backend deployment references these secrets → pods will fail to start

**Fix needed:**
1. Create working plain Kubernetes Secrets for local development (not committed to git)
2. Update documentation to clarify SealedSecrets are for production only
3. Add `.env` example for local dev secrets

---

### 2. Docker Compose Structure Ambiguity ❌

**Problem:** `infra/database/docker-compose.yml` contains BOTH postgres AND backend services.

**Current structure:**
```
infra/
├── database/
│   └── docker-compose.yml  ← Contains postgres + backend services
└── k3s/
    ├── postgres.yaml       ← k3s deployment
    ├── backend.yaml        ← k3s deployment
    └── secrets/            ← SealedSecrets (not working)
```

**Why it's confusing:**
1. Folder name `database/` suggests only database, but also has backend service
2. Frontend is missing from Docker Compose (only backend + postgres)
3. In k3s folder, naming is clear (`postgres.yaml`, `backend.yaml`) but secrets don't work
4. User has to maintain two deployment configs (Docker Compose + k3s) with different structures

**User's questions:**
> "kenapa dkcer compose database ada backend juga? knpa fe nya gada dan knpa di foler db? jadinya ambigu tpi di di k3s ada backend yaml dan postges yaml penaamnya jug ambgiu yg fe ga join?"

**Valid concerns:**
- ✅ Backend in `database/` folder is confusing
- ✅ Frontend missing from Docker Compose
- ✅ Folder naming doesn't match contents
- ✅ k3s structure is clearer (separate files per service)

---

## Recommended Fixes

### Fix 1: k3s Secrets Setup

**Option A: Plain Secrets for Local Dev (Recommended)**
```yaml
# infra/k3s/secrets/local-secrets.yaml (gitignored)
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
  namespace: podland
type: Opaque
stringData:
  password: "dev-password-123"
---
apiVersion: v1
kind: Secret
metadata:
  name: podland-backend-secret
  namespace: podland
type: Opaque
stringData:
  jwt-secret: "dev-jwt-secret-min-32-chars"
  refresh-token-secret: "dev-refresh-secret-min-32-chars"
  github-client-id: "dev-client-id"
  github-client-secret: "dev-client-secret"
```

**Option B: Keep SealedSecrets but add generation script**
- Update `generate-sealed-secrets.sh` to work properly
- Add README explaining how to generate

**Decision:** Use Option A for local dev, keep SealedSecrets as template for production.

---

### Fix 2: Docker Compose Structure

**Option A: Rename and restructure (Recommended)**
```
infra/
├── docker-compose/
│   ├── docker-compose.yml      ← All services (postgres + backend + frontend)
│   └── .env                    ← Environment variables
├── k3s/
│   ├── namespace.yaml
│   ├── postgres.yaml
│   ├── backend.yaml
│   ├── frontend.yaml           ← Add missing frontend deployment
│   └── secrets/
│       ├── local-secrets.yaml  ← For local dev (gitignored)
│       └── *.sealedsecret.yaml ← For production (committed)
```

**Option B: Separate services**
```
infra/
├── database/
│   └── docker-compose.yml      ← Only postgres
├── backend/
│   └── docker-compose.yml      ← Only backend
├── frontend/
│   └── docker-compose.yml      ← Only frontend
└── all/
    └── docker-compose.yml      ← Combined (references other folders)
```

**Decision:** Option A is simpler and matches k3s structure.

---

## Files to Create/Modify

### Create:
1. `infra/k3s/secrets/local-secrets.yaml.example` - Template for local secrets
2. `infra/k3s/secrets/local-secrets.yaml` - Actual secrets (gitignored)
3. `infra/k3s/frontend.yaml` - Frontend deployment (missing!)
4. `infra/docker-compose/.env` - Environment variables
5. `infra/docker-compose/docker-compose.yml` - Combined compose file

### Modify:
1. `infra/k3s/backend.yaml` - Fix secret references if needed
2. `infra/database/docker-compose.yml` - Move to `infra/docker-compose/`
3. `infra/README.md` - Update documentation

### Add to .gitignore:
- `infra/k3s/secrets/local-secrets.yaml`
- `infra/docker-compose/.env`

---

## Acceptance Criteria

- [ ] k3s secrets work for local development
- [ ] Frontend deployment exists in k3s
- [ ] Docker Compose structure is clear (all services in one place)
- [ ] Documentation explains the difference between Docker Compose (local dev) and k3s (production)
- [ ] No ambiguous folder names

---

*Context gathered: 2026-03-27*

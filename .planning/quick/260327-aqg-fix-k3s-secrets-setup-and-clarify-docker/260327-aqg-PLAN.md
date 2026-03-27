# Quick Task 260327-aqg: Plan

**Task:** Fix k3s secrets setup and clarify Docker Compose structure
**Date:** 2026-03-27
**Estimate:** 2 hours

---

## Objectives

1. ✅ Fix k3s secrets to work for local development
2. ✅ Add missing frontend deployment for k3s
3. ✅ Restructure Docker Compose to be clear and complete
4. ✅ Update documentation

---

## Tasks

### Task 1: Create Local Secrets for k3s

**Files:**
- `infra/k3s/secrets/local-secrets.yaml.example`
- `infra/k3s/secrets/.gitignore`

**Implementation:**
```yaml
# local-secrets.yaml.example
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
  namespace: podland
type: Opaque
stringData:
  password: "change-me-dev-password"
---
apiVersion: v1
kind: Secret
metadata:
  name: podland-backend-secret
  namespace: podland
type: Opaque
stringData:
  jwt-secret: "change-me-jwt-secret-min-32-chars"
  refresh-token-secret: "change-me-refresh-secret-min-32-chars"
  github-client-id: "change-me-client-id"
  github-client-secret: "change-me-client-secret"
```

**Acceptance:**
- [ ] Example file created with placeholder values
- [ ] `.gitignore` added to prevent committing real secrets

---

### Task 2: Create Frontend Deployment for k3s

**File:** `infra/k3s/frontend.yaml`

**Implementation:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: podland-frontend
  namespace: podland
spec:
  replicas: 1
  selector:
    matchLabels:
      app: podland-frontend
  template:
    metadata:
      labels:
        app: podland-frontend
    spec:
      containers:
        - name: frontend
          image: podland/frontend:latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
              name: http
          env:
            - name: VITE_API_URL
              value: "http://localhost:8080"
          livenessProbe:
            httpGet:
              path: /
              port: 80
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /
              port: 80
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 200m
              memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: podland-frontend
  namespace: podland
spec:
  selector:
    app: podland-frontend
  ports:
    - port: 80
      targetPort: 80
      protocol: TCP
  type: LoadBalancer
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: frontend-build-pvc
  namespace: podland
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 500Mi
  storageClassName: local-path
```

**Acceptance:**
- [ ] Frontend deployment created
- [ ] Service exposes port 80
- [ ] Resource limits defined

---

### Task 3: Restructure Docker Compose

**Files:**
- `infra/docker-compose/docker-compose.yml` (new)
- `infra/docker-compose/.env` (new)
- `infra/database/docker-compose.yml` (delete or deprecate)

**Implementation:**
```yaml
# infra/docker-compose/docker-compose.yml
version: "3.8"

services:
  postgres:
    image: postgres:15-alpine
    container_name: podland-postgres
    restart: unless-stopped
    env_file:
      - .env
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U podland"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: ../../apps/backend
      dockerfile: Dockerfile
    container_name: podland-backend
    restart: unless-stopped
    env_file:
      - .env
    ports:
      - "8080:8080"
    volumes:
      - avatar_storage:/app/uploads/avatars
    depends_on:
      postgres:
        condition: service_healthy

  frontend:
    build:
      context: ../../apps/frontend
      dockerfile: Dockerfile
    container_name: podland-frontend
    restart: unless-stopped
    env_file:
      - .env
    ports:
      - "3000:80"
    depends_on:
      - backend

volumes:
  postgres_data:
  avatar_storage:
```

**Acceptance:**
- [ ] All 3 services in one compose file
- [ ] Clear folder name (`docker-compose/` not `database/`)
- [ ] `.env` file for configuration

---

### Task 4: Update Documentation

**Files:**
- `infra/README.md` (create)
- `infra/k3s/README.md` (create)

**Content:**
```markdown
# Infrastructure

## Local Development (Docker Compose)

```bash
cd infra/docker-compose
docker-compose up -d
```

## Production (k3s)

### 1. Setup secrets

```bash
cd infra/k3s/secrets
cp local-secrets.yaml.example local-secrets.yaml
# Edit local-secrets.yaml with your values
kubectl apply -f local-secrets.yaml
```

### 2. Apply manifests

```bash
kubectl apply -f namespace.yaml
kubectl apply -f postgres.yaml
kubectl apply -f backend.yaml
kubectl apply -f frontend.yaml
```
```

**Acceptance:**
- [ ] Clear documentation for both deployment methods
- [ ] Explains when to use Docker Compose vs k3s

---

## Files Summary

| Action | File | Purpose |
|--------|------|---------|
| Create | `infra/k3s/secrets/local-secrets.yaml.example` | Template for local dev secrets |
| Create | `infra/k3s/secrets/.gitignore` | Prevent committing secrets |
| Create | `infra/k3s/frontend.yaml` | Frontend deployment (was missing) |
| Create | `infra/docker-compose/docker-compose.yml` | Combined compose file |
| Create | `infra/docker-compose/.env` | Environment variables |
| Create | `infra/README.md` | Infrastructure documentation |
| Modify | `infra/k3s/backend.yaml` | Minor cleanup if needed |
| Deprecate | `infra/database/docker-compose.yml` | Replace with new structure |

---

## Acceptance Criteria

- [ ] k3s secrets work (using local-secrets.yaml for dev)
- [ ] Frontend deployment exists in k3s
- [ ] Docker Compose has all 3 services (postgres, backend, frontend)
- [ ] Folder structure is clear and not ambiguous
- [ ] Documentation explains both deployment methods

---

*Plan created: 2026-03-27*

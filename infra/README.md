# Infrastructure

Podland deployment infrastructure for both local development and production.

## Quick Start

### Local Development (Docker Compose)

For local development and testing:

```bash
cd infra/docker-compose

# Copy environment template
cp .env.example .env

# Edit .env with your values (especially GitHub OAuth credentials)

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

**Services:**
- PostgreSQL: `localhost:5432`
- Backend API: `localhost:8080`
- Frontend: `localhost:3000`

---

### Production (k3s)

For production deployment on k3s cluster:

#### 1. Setup Secrets

```bash
cd infra/k3s/secrets

# Copy the example file
cp local-secrets.yaml.example local-secrets.yaml

# Edit with your actual values
# IMPORTANT: Never commit local-secrets.yaml to git!
```

**Required secrets:**
- PostgreSQL password
- JWT secret (min 32 chars)
- Refresh token secret (min 32 chars)
- GitHub OAuth Client ID
- GitHub OAuth Client Secret

#### 2. Apply to Cluster

```bash
cd infra/k3s

# Apply secrets first
kubectl apply -f secrets/local-secrets.yaml

# Apply namespace
kubectl apply -f namespace.yaml

# Apply services
kubectl apply -f postgres.yaml
kubectl apply -f backend.yaml
kubectl apply -f frontend.yaml

# Check status
kubectl get pods -n podland
kubectl get services -n podland
```

**Services:**
- PostgreSQL: `postgres.podland.svc.cluster.local:5432` (ClusterIP)
- Backend: `podland-backend.podland.svc.cluster.local:8080` (ClusterIP)
- Frontend: `podland-frontend.podland.svc.cluster.local:80` (LoadBalancer)

---

## Structure

```
infra/
├── docker-compose/           # Local development
│   ├── docker-compose.yml    # All services (postgres, backend, frontend)
│   ├── .env.example          # Environment template
│   └── .gitignore
├── k3s/                      # Production (k3s cluster)
│   ├── namespace.yaml        # Kubernetes namespace
│   ├── postgres.yaml         # PostgreSQL StatefulSet
│   ├── backend.yaml          # Backend Deployment
│   ├── frontend.yaml         # Frontend Deployment
│   └── secrets/
│       ├── local-secrets.yaml.example  # Template for local dev
│       ├── local-secrets.yaml          # Actual secrets (gitignored)
│       └── *.sealedsecret.yaml         # SealedSecrets for production
└── README.md                 # This file
```

---

## When to Use What

| Environment | Use | Why |
|-------------|-----|-----|
| Local development | Docker Compose | Simple, fast, no cluster needed |
| Testing/Staging | k3s | Matches production, tests K8s manifests |
| Production | k3s | Cloud native, scalable, resilient |

---

## Troubleshooting

### Docker Compose

**Backend can't connect to database:**
```bash
# Check database is healthy
docker-compose ps postgres

# Check connection string in .env
grep DATABASE_URL .env
```

**Frontend can't connect to backend:**
```bash
# Check VITE_API_URL in .env
grep VITE_API_URL .env

# Should be http://localhost:8080 for local dev
```

### k3s

**Pods not starting:**
```bash
# Check events
kubectl get events -n podland --sort-by='.lastTimestamp'

# Check pod logs
kubectl logs <pod-name> -n podland
```

**Secrets not found:**
```bash
# List secrets
kubectl get secrets -n podland

# Check secret content
kubectl get secret postgres-secret -n podland -o jsonpath='{.data}'
```

---

## Next Steps

1. **Setup GitHub OAuth** - Create OAuth app in GitHub Developer Settings
2. **Configure database** - Update connection strings if using external database
3. **Deploy to k3s** - Follow production deployment guide above

---

*Last updated: 2026-03-27*

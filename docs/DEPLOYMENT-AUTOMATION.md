# Deployment Automation Guide

**Created:** 2026-03-30  
**Status:** Ready to use

---

## Quick Start

### Option 1: Manual Script (Now)

```bash
# Set environment variables
export PODLAND_SERVER_HOST="podland.app"
export PODLAND_SERVER_USER="ubuntu"
export PODLAND_SSH_KEY="~/.ssh/id_ed25519"

# Run deployment
chmod +x scripts/deploy-prod.sh
./scripts/deploy-prod.sh
```

**What it does:**
1. Builds backend & frontend Docker images
2. Pushes to Docker Hub
3. SSH to server
4. Applies Kubernetes manifests
5. Restarts deployments
6. Health check

**Time:** ~5-10 minutes

---

### Option 2: GitHub Actions (Automated)

**Setup:**

1. **Add repository secrets:**

```bash
# Go to: GitHub Repo → Settings → Secrets and variables → Actions

# Add these secrets:
DOCKERHUB_USERNAME       # Your Docker Hub username
DOCKERHUB_TOKEN          # Docker Hub access token
PROD_SERVER_HOST         # Server hostname (e.g., podland.app)
PROD_SERVER_USER         # SSH user (e.g., ubuntu)
PROD_SSH_PRIVATE_KEY     # SSH private key for deployment
```

2. **Enable production environment:**

```bash
# Go to: GitHub Repo → Settings → Environments
# Create environment: "production"
# Add deployment branches: main
# Optionally add required reviewers
```

3. **Push to main:**

```bash
git push origin main
# → Auto-deploys to production!
```

**Time:** ~10-15 minutes (fully automated)

---

## Deployment Script Details

### Pre-requisites

**On your laptop:**
- Docker installed
- SSH access to server
- kubectl configured (optional, for debugging)

**On server:**
- k3s cluster running
- kubectl configured
- SSH key-based auth enabled
- Docker Hub credentials (for pulling images)

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PODLAND_SERVER_HOST` | Production server hostname | - |
| `PODLAND_SERVER_USER` | SSH username | `ubuntu` |
| `PODLAND_SSH_KEY` | Path to SSH private key | `~/.ssh/id_ed25519` |

### What Gets Deployed

```
┌──────────────────────────────────────────────────────┐
│  Kubernetes Resources                                │
│                                                       │
│  ┌─────────────────┐  ┌─────────────────┐           │
│  │  Podland        │  │  Monitoring     │           │
│  │  Backend        │  │  Stack          │           │
│  │  - 3 replicas   │  │  - Prometheus   │           │
│  │  - Go 1.25      │  │  - Grafana      │           │
│  │  - Port 8080    │  │  - Loki         │           │
│  └─────────────────┘  │  - Alertmanager │           │
│  ┌─────────────────┐  └─────────────────┘           │
│  │  Podland        │                                 │
│  │  Frontend       │  ┌─────────────────┐           │
│  │  - 3 replicas   │  │  PostgreSQL     │           │
│  │  - React 18     │  │  - 1 PVC        │           │
│  │  - Port 3000    │  │  - 10GB storage │           │
│  └─────────────────┘  └─────────────────┘           │
└──────────────────────────────────────────────────────┘
```

---

## Rollback Strategy

### Quick Rollback

```bash
# Rollback to previous version
kubectl rollout undo deployment/podland-backend
kubectl rollout undo deployment/podland-frontend

# Rollback to specific version
kubectl rollout undo deployment/podland-backend --to-revision=2
```

### Manual Rollback (via Script)

```bash
# Deploy specific version
export BACKEND_IMAGE="podland/backend:abc123"  # Specific commit
export FRONTEND_IMAGE="podland/frontend:abc123"

./scripts/deploy-prod.sh
```

---

## Monitoring Deployment

### Check Deployment Status

```bash
# Watch rollout status
kubectl rollout status deployment/podland-backend
kubectl rollout status deployment/podland-frontend

# View deployment history
kubectl rollout history deployment/podland-backend
kubectl rollout history deployment/podland-frontend
```

### Check Pod Health

```bash
# List all pods
kubectl get pods -n podland

# Watch pod status in real-time
kubectl get pods -n podland -w

# View pod logs
kubectl logs -l app=podland-backend -f
kubectl logs -l app=podland-frontend -f
```

### Health Endpoints

| Endpoint | Description |
|----------|-------------|
| `https://podland.app/api/health` | Backend health check |
| `https://podland.app/api/health/live` | Liveness probe |
| `https://podland.app/api/health/ready` | Readiness probe |
| `https://podland.app/admin/health` | System health dashboard |

---

## Troubleshooting

### Deployment Fails

**Check logs:**

```bash
# Last 100 lines
kubectl logs deployment/podland-backend --tail=100

# With timestamps
kubectl logs deployment/podland-backend -f --timestamps

# Previous instance (if crashed)
kubectl logs deployment/podland-backend --previous
```

**Common issues:**

1. **ImagePullBackOff**
   ```bash
   # Check Docker Hub credentials
   kubectl get secrets -n podland
   kubectl describe secret regcred -n podland
   ```

2. **CrashLoopBackOff**
   ```bash
   # Check environment variables
   kubectl describe deployment/podland-backend
   
   # Check database connection
   kubectl exec deployment/podland-backend -- env | grep DATABASE
   ```

3. **Service Unavailable**
   ```bash
   # Check service endpoints
   kubectl get endpoints -n podland
   
   # Check ingress
   kubectl describe ingress podland-ingress -n podland
   ```

### Health Check Fails

```bash
# SSH to server
ssh ubuntu@podland.app

# Check pod status
kubectl get pods -n podland

# Test health endpoint from inside cluster
kubectl run curl --image=curlimages/curl --rm -it --restart=Never -- \
  http://podland-backend.podland.svc:8080/api/health
```

---

## Security Best Practices

### 1. Use Sealed Secrets

```bash
# Install kubeseal
brew install kubeseal

# Create secret
kubectl create secret generic backend-secrets \
  --from-literal=jwt-secret=supersecret \
  --from-literal=database-url=postgresql://... \
  --dry-run=client -o yaml > backend-secrets.yaml

# Seal it
kubeseal --format yaml < backend-secrets.yaml > backend-sealed-secret.yaml

# Commit sealed secret to Git
git add infra/k3s/secrets/backend-sealed-secret.yaml
git commit -m "Add sealed backend secrets"
```

### 2. Network Policies

```yaml
# infra/k3s/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-isolation
  namespace: podland
spec:
  podSelector:
    matchLabels:
      app: podland-backend
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: podland-frontend
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
```

### 3. Resource Limits

```yaml
# infra/k3s/backend.yaml
spec:
  template:
    spec:
      containers:
      - name: backend
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

---

## Cost Optimization

### 1. Horizontal Pod Autoscaler

```yaml
# infra/k3s/backend-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
  namespace: podland
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: podland-backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### 2. Cluster Autoscaler

```bash
# Enable cluster autoscaler in k3s
# Automatically scales nodes based on demand
```

---

## Next Steps

### Immediate (This Week)

- [ ] Test deployment script on staging
- [ ] Add GitHub Actions secrets
- [ ] Run first automated deployment

### Short-term (v1.1)

- [ ] Add staging environment
- [ ] Implement blue-green deployment
- [ ] Add Slack notifications
- [ ] Set up monitoring alerts

### Long-term (v2.0)

- [ ] GitOps with ArgoCD
- [ ] Infrastructure as Code (Terraform)
- [ ] Multi-region deployment
- [ ] Auto-scaling based on metrics

---

*Guide created: 2026-03-30*  
*Version: v1.0*

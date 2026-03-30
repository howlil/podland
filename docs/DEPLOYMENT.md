# Deployment Guide — Podland v1.0

**Last Updated:** 2026-03-30  
**Version:** v1.0  
**Status:** Production Ready

---

## Overview

This guide covers deploying Podland to production. Podland is a multi-tenant PaaS built on k3s, designed for students to deploy applications with zero DevOps knowledge.

**Infrastructure:**
- Single bare-metal server (Proxmox-managed)
- k3s cluster for container orchestration
- Cloudflare for DNS and Tunnel
- PostgreSQL for database
- Monitoring stack (Prometheus, Grafana, Loki)

---

## Prerequisites

### Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 8 cores | 16+ cores |
| RAM | 16 GB | 32+ GB |
| Storage | 100 GB SSD | 500 GB+ NVMe |
| Network | 1 Gbps | 1 Gbps+ |

### Software Requirements

- Proxmox VE 7.x or later
- Ubuntu 22.04 LTS (for k3s node)
- Docker 20.10+
- k3s 1.28+
- Git
- kubectl

### External Services

- **Cloudflare Pro** account (for API access)
- **GitHub OAuth** application
- **SendGrid** account (for email notifications)

---

## Environment Variables

### Required Variables

Create a `.env` file in `apps/backend/`:

```bash
# ===== DATABASE =====
DATABASE_URL=postgresql://user:password@localhost:5432/podland?sslmode=disable

# ===== AUTHENTICATION =====
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
GITHUB_CLIENT_ID=your-github-oauth-client-id
GITHUB_CLIENT_SECRET=your-github-oauth-client-secret

# ===== CLOUDFLARE =====
CLOUDFLARE_API_TOKEN=your-cloudflare-api-token
CLOUDFLARE_ZONE_ID=your-cloudflare-zone-id

# ===== ALERTMANAGER =====
ALERTMANAGER_WEBHOOK_SECRET=your-webhook-secret-token-min-32-chars

# ===== EMAIL (SENDGRID) =====
SENDGRID_API_KEY=your-sendgrid-api-key
SENDGRID_FROM_EMAIL=noreply@podland.app

# ===== SERVER =====
PORT=8080
```

### Variable Descriptions

| Variable | Description | How to Obtain |
|----------|-------------|---------------|
| `DATABASE_URL` | PostgreSQL connection string | Create DB, format: `postgresql://user:pass@host:port/dbname?sslmode=disable` |
| `JWT_SECRET` | Secret key for JWT signing | Generate: `openssl rand -base64 32` |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | GitHub Settings → Developer Settings → OAuth Apps |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | Same as above |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token | Cloudflare Dashboard → Profile → API Tokens |
| `CLOUDFLARE_ZONE_ID` | Cloudflare zone ID | Cloudflare Dashboard → Overview |
| `ALERTMANAGER_WEBHOOK_SECRET` | Secret for Alertmanager webhook | Generate: `openssl rand -base64 32` |
| `SENDGRID_API_KEY` | SendGrid API key | SendGrid Dashboard → API Keys |
| `SENDGRID_FROM_EMAIL` | From address for emails | Your domain email |
| `PORT` | Backend server port | Default: 8080 |

### Security Notes

⚠️ **IMPORTANT:**
- Never commit `.env` file to Git
- Use strong, randomly generated secrets (min 32 characters)
- Rotate secrets every 90 days
- Store secrets in a secure vault (e.g., 1Password, Bitwarden)

---

## Step 1: Server Setup

### 1.1 Install Ubuntu 22.04

```bash
# On Proxmox, create new VM with:
# - Ubuntu 22.04 LTS ISO
# - 8+ CPU cores
# - 16+ GB RAM
# - 100+ GB storage
```

### 1.2 Update System

```bash
sudo apt update && sudo apt upgrade -y
sudo reboot
```

### 1.3 Install Dependencies

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install k3s
curl -sfL https://get.k3s.io | sh -

# Verify k3s
sudo k3s kubectl get nodes
```

### 1.4 Configure kubectl

```bash
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $USER:$USER ~/.kube/config
```

---

## Step 2: Database Setup

### 2.1 Install PostgreSQL

```bash
sudo apt install postgresql postgresql-contrib -y
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

### 2.2 Create Database and User

```bash
sudo -u postgres psql

CREATE USER podland WITH PASSWORD 'your-secure-password';
CREATE DATABASE podland OWNER podland;
GRANT ALL PRIVILEGES ON DATABASE podland TO podland;
\q
```

### 2.3 Test Connection

```bash
psql postgresql://podland:your-secure-password@localhost:5432/podland
```

---

## Step 3: Backend Deployment

### 3.1 Clone Repository

```bash
git clone https://github.com/your-org/podland.git
cd podland
```

### 3.2 Create .env File

```bash
cd apps/backend
cp .env.example .env
# Edit .env with your values
nano .env
```

### 3.3 Build Backend

```bash
cd apps/backend
go build -o bin/podland ./cmd
```

### 3.4 Create Systemd Service

```bash
sudo nano /etc/systemd/system/podland.service
```

**Content:**

```ini
[Unit]
Description=Podland Backend Service
After=network.target postgresql.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/podland/apps/backend
EnvironmentFile=/home/ubuntu/podland/apps/backend/.env
ExecStart=/home/ubuntu/podland/apps/backend/bin/podland
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 3.5 Enable and Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable podland
sudo systemctl start podland
sudo systemctl status podland
```

### 3.6 Verify Backend

```bash
curl http://localhost:8080/api/health
```

**Expected Response:**

```json
{"status":"ok","timestamp":"2026-03-30T..."}
```

---

## Step 4: Frontend Deployment

### 4.1 Install Node.js

```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
```

### 4.2 Install Dependencies

```bash
cd apps/frontend
npm install
```

### 4.3 Build Frontend

```bash
npm run build
```

### 4.4 Serve with Caddy (Recommended)

```bash
# Install Caddy
sudo add-apt-repository ppa:caddy/stable
sudo apt update
sudo apt install caddy

# Configure Caddy
sudo nano /etc/caddy/Caddyfile
```

**Content:**

```
podland.app {
    reverse_proxy localhost:3000
}

api.podland.app {
    reverse_proxy localhost:8080
}
```

### 4.5 Start Caddy

```bash
sudo systemctl enable caddy
sudo systemctl start caddy
```

---

## Step 5: Kubernetes Resources

### 5.1 Apply Monitoring Stack

```bash
cd infra/k3s/monitoring
kubectl apply -f namespace.yaml
kubectl apply -f prometheus-operator.yaml
kubectl apply -f rbac.yaml
# ... apply all monitoring files
```

### 5.2 Apply Backup CronJob

```bash
cd infra/k3s/backups
kubectl apply -f backup-cronjob.yaml
kubectl apply -f backup-pvc.yaml
```

### 5.3 Verify Pods

```bash
kubectl get pods -n monitoring
kubectl get pods -n podland
```

---

## Step 6: Cloudflare Configuration

### 6.1 Create OAuth App on GitHub

1. Go to GitHub Settings → Developer Settings → OAuth Apps
2. Click "New OAuth App"
3. **Application name:** Podland
4. **Homepage URL:** `https://podland.app`
5. **Authorization callback URL:** `https://podland.app/api/auth/github/callback`
6. Copy Client ID and generate Client Secret

### 6.2 Configure Cloudflare DNS

```bash
# In Cloudflare Dashboard:
# 1. Add A record: @ → your-server-ip
# 2. Add A record: api → your-server-ip
# 3. Add CNAME: *.vm → your-server-ip (for VM subdomains)
```

### 6.3 Generate Cloudflare API Token

1. Go to Cloudflare Dashboard → Profile → API Tokens
2. Create Token → Custom Token
3. **Permissions:**
   - Zone → DNS → Edit
   - Zone → Zone → Read
4. Copy token to `.env`

---

## Step 7: SendGrid Configuration

### 7.1 Create API Key

1. Go to SendGrid Dashboard → API Keys
2. Create API Key → Full Access
3. Copy API key to `.env`

### 7.2 Configure Sender

1. Go to SendGrid Dashboard → Settings → Sender Authentication
2. Verify your domain (podland.app)
3. Set verified email as `SENDGRID_FROM_EMAIL`

---

## Step 8: Create Superadmin User

### 8.1 Sign Up First User

1. Go to `https://podland.app`
2. Sign in with GitHub (use your admin account)
3. Note your user email

### 8.2 Promote to Superadmin

```bash
sudo -u postgres psql -d podland

UPDATE users SET role = 'superadmin' WHERE email = 'your-email@example.com';
\q
```

### 8.3 Verify

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" http://localhost:8080/api/admin/health
```

---

## Step 9: Final Verification

### 9.1 Health Check

```bash
curl https://api.podland.app/api/health
```

### 9.2 Test OAuth Flow

1. Visit `https://podland.app`
2. Click "Sign in with GitHub"
3. Complete OAuth flow
4. Verify dashboard loads

### 9.3 Test Admin Panel

1. Go to `/admin`
2. Verify you can see user list
3. Verify system health dashboard shows metrics

### 9.4 Test VM Creation

1. Create a test VM
2. Verify VM appears in list
3. Verify domain is assigned
4. Verify metrics appear in observability tab

---

## Troubleshooting

### Backend Won't Start

```bash
# Check logs
sudo journalctl -u podland -f

# Common issues:
# - Missing .env variables → Check all required vars
# - Database connection failed → Verify DATABASE_URL
# - Port already in use → Change PORT in .env
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Test connection
psql postgresql://podland:password@localhost:5432/podland

# Check pg_hba.conf allows local connections
sudo nano /etc/postgresql/14/main/pg_hba.conf
```

### Cloudflare DNS Not Working

```bash
# Verify API token has correct permissions
curl -X GET "https://api.cloudflare.com/client/v4/user/tokens/verify" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Check zone ID is correct
curl -X GET "https://api.cloudflare.com/client/v4/zones" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Monitoring Stack Not Working

```bash
# Check pods are running
kubectl get pods -n monitoring

# Check Prometheus is scraping
kubectl port-forward svc/prometheus-operated -n monitoring 9090:9090
# Visit http://localhost:9090/targets
```

---

## Backup and Restore

### Database Backup

```bash
# Manual backup
pg_dump postgresql://podland:password@localhost:5432/podland > backup-$(date +%Y%m%d).sql

# Restore from backup
psql postgresql://podland:password@localhost:5432/podland < backup-20260330.sql
```

### Kubernetes Backup

```bash
# Velero backup (if configured)
velero backup create podland-backup-$(date +%Y%m%d)

# Restore
velero restore create --from-backup podland-backup-20260330
```

---

## Security Checklist

- [ ] All `.env` files excluded from Git (`.gitignore`)
- [ ] Strong, randomly generated secrets (32+ chars)
- [ ] HTTPS enabled for all endpoints
- [ ] Firewall configured (only 80, 443, 22 open)
- [ ] Fail2ban installed and configured
- [ ] Regular security updates enabled
- [ ] Database backups tested
- [ ] Monitoring alerts configured

---

## Post-Deployment Tasks

### Within 24 Hours

- [ ] Verify all health checks pass
- [ ] Test all user flows (sign in, create VM, delete VM)
- [ ] Check monitoring dashboards
- [ ] Verify backup jobs running

### Within 7 Days

- [ ] Review audit logs for admin actions
- [ ] Check for any error spikes in logs
- [ ] Gather user feedback
- [ ] Update npm dependencies (fix esbuild vulnerability)

### Within 30 Days

- [ ] Run Lighthouse audit for performance
- [ ] Fix accessibility issues (WCAG AA)
- [ ] Add OpenAPI documentation
- [ ] Plan v1.1 features

---

## Support

**Documentation:** `/docs` directory  
**API Documentation:** `https://api.podland.app/api/docs` (after Swagger setup)  
**Status Page:** `https://status.podland.app` (configure separately)

---

*Deployment guide created: 2026-03-30*  
*Version: v1.0*

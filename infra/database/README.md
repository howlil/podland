# Local Development Setup

## Quick Start

### Option 1: Docker Compose (Recommended)

1. **Run setup script:**
   ```bash
   # Windows
   setup-dev.bat
   
   # Linux/Mac
   chmod +x setup-dev.sh
   ./setup-dev.sh
   ```

2. **Update GitHub OAuth credentials:**
   - Edit `.env` file
   - Set `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET`
   - Create OAuth app at: https://github.com/settings/developers

3. **Start services:**
   ```bash
   docker-compose up -d
   ```

4. **Access:**
   - Backend: http://localhost:8080
   - PostgreSQL: localhost:5432

### Option 2: Manual Setup

1. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Generate secure secrets:**
   ```bash
   # PostgreSQL password
   openssl rand -base64 32
   
   # JWT secret
   openssl rand -base64 32
   
   # Or use Node.js
   node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"
   ```

3. **Edit `.env`** with your generated secrets

4. **Start:**
   ```bash
   docker-compose up -d
   ```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `POSTGRES_PASSWORD` | Database password | ✅ |
| `JWT_SECRET` | JWT signing secret (min 32 chars) | ✅ |
| `REFRESH_TOKEN_SECRET` | Refresh token secret (min 32 chars) | ✅ |
| `GITHUB_CLIENT_ID` | GitHub OAuth Client ID | ✅ |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth Client Secret | ✅ |

## Troubleshooting

**Port already in use:**
```bash
# Check what's using port 5432 or 8080
netstat -ano | findstr :5432
netstat -ano | findstr :8080
```

**Reset database:**
```bash
docker-compose down -v
docker-compose up -d
```

## k3s Local Development

For testing k3s manifests locally:

1. **Install k3d:**
   ```bash
   # Windows (with scoop)
   scoop install k3d
   
   # Or download from https://k3d.io
   ```

2. **Create local cluster:**
   ```bash
   k3d cluster create podland-dev
   ```

3. **Install Sealed Secrets:**
   ```bash
   kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/latest/download/controller.yaml
   ```

4. **Generate sealed secrets:**
   ```bash
   # Create plain secret first
   cat > secret.yaml <<EOF
   apiVersion: v1
   kind: Secret
   metadata:
     name: postgres-secret
     namespace: podland
   type: Opaque
   stringData:
     password: "your-password"
   EOF
   
   # Seal it
   kubeseal --format yaml < secret.yaml > sealed-secret.yaml
   
   # Apply
   kubectl apply -f sealed-secret.yaml
   
   # Cleanup plain secret
   rm secret.yaml
   ```

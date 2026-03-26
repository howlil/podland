# Podland

**Student PaaS Platform** — Deploy and run applications with zero DevOps knowledge.

## Overview

Podland is a multi-tenant Platform as a Service (PaaS) for students at Universitas Andalas. Built on Kubernetes (k3s), it provides automatic resource allocation, domain setup via Cloudflare, and built-in observability.

**Authentication:** Requires GitHub OAuth with `@student.unand.ac.id` email verification.

## Quick Start

### Prerequisites

- **Node.js** 20+
- **Go** 1.22+
- **PostgreSQL** 15+ (or Docker)
- **GitHub OAuth App** credentials

### 1. Clone and Setup

```bash
cd podland
```

### 2. Create GitHub OAuth App

1. Go to GitHub Settings → Developer Settings → OAuth Apps
2. Click "New OAuth App"
3. Set:
   - **Application name:** Podland
   - **Homepage URL:** `http://localhost:3000`
   - **Authorization callback URL:** `http://localhost:8080/api/auth/github/callback`
4. Copy Client ID and Client Secret

### 3. Configure Environment

**Backend (.env in `apps/backend/`):**
```bash
cp .env.example .env
```

Edit `.env`:
```bash
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
JWT_SECRET=generate-a-secure-random-string-min-32-chars
REFRESH_TOKEN_SECRET=generate-another-secure-random-string-min-32-chars
DATABASE_URL=postgresql://podland:password@localhost:5432/podland?sslmode=disable
```

### 4. Start Database (Docker)

```bash
cd infra/database
docker-compose up -d
```

### 5. Install Dependencies

```bash
# Root
npm install

# Backend
cd apps/backend
go mod download

# Frontend
cd apps/frontend
npm install
```

### 6. Run Development Servers

**Terminal 1 - Backend:**
```bash
cd apps/backend
go run ./cmd/main.go
```

Backend runs on `http://localhost:8080`

**Terminal 2 - Frontend:**
```bash
cd apps/frontend
npm run dev
```

Frontend runs on `http://localhost:3000`

### 7. Test Authentication

1. Open `http://localhost:3000`
2. Click "Sign in with GitHub"
3. Authorize the application
4. If your GitHub primary email ends with `@student.unand.ac.id`, you'll be logged in

## Project Structure

```
podland/
├── apps/
│   ├── backend/           # Go API server
│   │   ├── cmd/           # Entry point
│   │   ├── handlers/      # HTTP handlers
│   │   ├── middleware/    # CORS, CSRF, Auth
│   │   └── internal/      # Private packages
│   └── frontend/          # TanStack Start React app
│       ├── src/
│       │   ├── components/
│       │   ├── routes/
│       │   └── lib/
│       └── public/
├── packages/
│   └── types/             # Shared TypeScript types
├── infra/
│   ├── k3s/               # Kubernetes manifests
│   └── database/          # Docker Compose for local dev
└── uploads/
    └── avatars/           # User avatar storage
```

## Architecture

### Tech Stack

| Layer | Technology |
|-------|------------|
| **Frontend** | React, TanStack Start, Tailwind CSS v4, Zustand |
| **Backend** | Go 1.22, net/http, JWT |
| **Database** | PostgreSQL 15 |
| **Orchestration** | k3s (Kubernetes) |
| **Auth** | GitHub OAuth |

### Authentication Flow

```
1. User clicks "Sign in with GitHub"
2. Backend redirects to GitHub OAuth
3. GitHub redirects back with authorization code
4. Backend exchanges code for access token
5. Backend fetches user info + emails from GitHub
6. Backend validates primary email ends with @student.unand.ac.id
   - Invalid → Rejection page with instructions
   - Valid → Extract NIM, assign role
7. New user → Welcome screen (confirm NIM, accept terms)
8. Existing user → Create session, redirect to dashboard
```

### Session Management

- **Access Token:** JWT, 15-minute expiry, stored in memory
- **Refresh Token:** Opaque, 7-day expiry, HTTP-only cookie
- **CSRF Protection:** Double-submit cookie pattern
- **Session Limit:** Max 3 concurrent sessions per user

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/auth/login` | Initiate GitHub OAuth |
| GET | `/api/auth/github/callback` | OAuth callback |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Logout user |

### Users

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/users/me` | Current user | ✓ |
| GET | `/api/users/{id}` | User by ID | ✓ |
| POST | `/api/users/confirm-nim` | Confirm NIM | ✓ |

### Activity

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/activity` | Activity log | ✓ |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |

## Phase 1 Success Criteria

- [x] User with @student.unand.ac.id GitHub email can sign in
- [x] User with non-student email is rejected with clear error
- [x] NIM containing "1152" → Internal role, others → External
- [x] Session persists after browser refresh (7-day JWT)
- [x] Profile page shows display name, role, NIM
- [x] Sign out invalidates session
- [x] Dashboard displays quota usage (0/0.5 CPU for External)
- [x] Activity log shows "Account created"
- [x] Dashboard responsive (mobile 320px, desktop 1920px)

## Development

### Backend Commands

```bash
# Run tests
go test ./...

# Build
go build -o bin/backend ./cmd/main.go

# Format
go fmt ./...

# Lint
go vet ./...
```

### Frontend Commands

```bash
# Development
npm run dev

# Build
npm run build

# Preview build
npm run preview

# Lint
npm run lint
```

### Monorepo Commands

```bash
# Run all dev servers
npm run dev

# Build all
npm run build

# Clean
npm run clean
```

## Deployment (k3s)

```bash
# Create namespace and secrets
kubectl apply -f infra/k3s/namespace.yaml

# Deploy PostgreSQL
kubectl apply -f infra/k3s/postgres.yaml

# Deploy backend
kubectl apply -f infra/k3s/backend.yaml

# Check status
kubectl get pods -n podland
kubectl get services -n podland
```

## Security

- JWT access tokens (15 min)
- HTTP-only refresh cookies (7 days)
- CSRF protection via double-submit pattern
- CORS with allowed origins
- Session limit (max 3 per user)
- Token rotation on refresh
- Non-root container execution (Phase 2)

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request

---

**Built for students, by students.** 🚀

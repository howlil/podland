# Podland Backend

Go backend for Podland PaaS platform.

## Setup

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- GitHub OAuth App credentials

### Installation

```bash
# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Edit .env with your credentials
# - GITHUB_CLIENT_ID
# - GITHUB_CLIENT_SECRET
# - DATABASE_URL

# Run the server
go run ./cmd/main.go
```

Server will start on `http://localhost:8080`

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/auth/login` | Initiate GitHub OAuth login |
| GET | `/api/auth/github/callback` | OAuth callback handler |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Logout user |

### Users

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/users/me` | Get current user | Yes |
| GET | `/api/users/{id}` | Get user by ID | Yes |
| POST | `/api/users/confirm-nim` | Confirm/update NIM | Yes |

### Activity

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/activity` | Get user activity log | Yes |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |

## Project Structure

```
apps/backend/
├── cmd/
│   └── main.go           # Application entry point
├── handlers/
│   ├── auth.go           # OAuth and session handlers
│   ├── users.go          # User CRUD handlers
│   ├── activity.go       # Activity log handlers
│   └── health.go         # Health check handler
├── middleware/
│   └── middleware.go     # CORS, CSRF, Auth middleware
├── internal/
│   ├── auth/
│   │   ├── jwt.go        # JWT token generation/validation
│   │   ├── oauth.go      # GitHub OAuth integration
│   │   └── session.go    # Session management
│   ├── config/
│   │   └── config.go     # Environment configuration
│   └── database/
│       ├── database.go   # Database connection
│       ├── types.go      # Database types/interfaces
│       └── queries.go    # SQL queries
└── uploads/
    └── avatars/          # User avatar storage
```

## Development

```bash
# Run with auto-reload (requires air)
air

# Run tests
go test ./...

# Build
go build -o bin/backend ./cmd/main.go
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ENV` | No | `production` or `development` |
| `PORT` | No | Server port (default: 8080) |
| `FRONTEND_URL` | Yes | Frontend base URL |
| `GITHUB_CLIENT_ID` | Yes | GitHub OAuth Client ID |
| `GITHUB_CLIENT_SECRET` | Yes | GitHub OAuth Client Secret |
| `GITHUB_CALLBACK_URL` | Yes | OAuth callback URL |
| `JWT_SECRET` | Yes | JWT signing secret (min 32 chars) |
| `REFRESH_TOKEN_SECRET` | Yes | Refresh token secret (min 32 chars) |
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `ALLOWED_ORIGINS` | No | CORS allowed origins (comma-separated) |

## Security

- JWT access tokens (15 min expiry)
- Opaque refresh tokens (7 days, HTTP-only cookies)
- CSRF protection via double-submit cookie pattern
- CORS with allowed origins
- Session limit (max 3 per user)
- Token rotation on refresh

## License

MIT

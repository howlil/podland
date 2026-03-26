# Podland Quick Start Guide

## Prerequisites

- **Go** 1.22+
- **Node.js** 20+
- **Docker** (for PostgreSQL)

## 1. Start Database

```bash
cd infra/database
docker-compose up -d
```

PostgreSQL will run on `localhost:5432`

## 2. Setup GitHub OAuth

1. Go to https://github.com/settings/developers
2. Click "New OAuth App"
3. Fill in:
   - **Application name:** Podland Dev
   - **Homepage URL:** http://localhost:3000
   - **Authorization callback URL:** http://localhost:8080/api/auth/github/callback
4. Copy **Client ID** and generate a new **Client Secret**

## 3. Configure Backend

```bash
cd apps/backend
cp .env.example .env
```

Edit `.env`:
```env
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_client_secret_here
JWT_SECRET=dev-secret-key-min-32-characters-long
REFRESH_TOKEN_SECRET=another-dev-secret-key-min-32-chars
DATABASE_URL=postgresql://podland:podland_password_change_me@localhost:5432/podland?sslmode=disable
```

## 4. Start Backend

```bash
cd apps/backend
go run ./cmd/main.go
```

Backend runs on http://localhost:8080

Check health: http://localhost:8080/api/health

## 5. Start Frontend

Open a new terminal:

```bash
cd apps/frontend
npm install
npm run dev
```

Frontend runs on http://localhost:3000

## 6. Test Authentication

1. Open http://localhost:3000
2. Click "Sign in with GitHub"
3. Authorize the application
4. If your GitHub primary email ends with `@student.unand.ac.id`:
   - You'll see the welcome screen
   - Confirm your NIM
   - Accept terms
   - Redirected to dashboard
5. If not:
   - You'll see the rejection page
   - Follow instructions to add student email

## Testing Checklist

### Backend
- [ ] Health check: http://localhost:8080/api/health
- [ ] OAuth login: http://localhost:8080/api/auth/login

### Frontend
- [ ] Landing page loads
- [ ] Sign in button visible
- [ ] OAuth redirect works
- [ ] Welcome screen shows NIM
- [ ] Dashboard displays after sign-in
- [ ] Profile page shows user data
- [ ] Sign out works

### Responsive
- [ ] Desktop (1920px): Sidebar visible
- [ ] Mobile (320px): Bottom tab bar visible

### Dark Mode
- [ ] System dark mode detected
- [ ] All components styled correctly

## Troubleshooting

### Backend won't start

```bash
# Check if PostgreSQL is running
docker ps

# Check database connection
docker-compose logs postgres
```

### Frontend won't start

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

### OAuth callback fails

- Verify callback URL in GitHub OAuth app settings
- Check GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET in .env
- Ensure backend is running on port 8080

### Database connection error

```bash
# Restart PostgreSQL
docker-compose restart postgres

# Check connection
docker-compose exec postgres pg_isready -U podland
```

## API Testing

```bash
# Health check
curl http://localhost:8080/api/health

# Get current user (requires auth)
curl http://localhost:8080/api/users/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Get activity log (requires auth)
curl http://localhost:8080/api/activity \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Stop Everything

```bash
# Stop frontend (Ctrl+C in terminal)

# Stop backend (Ctrl+C in terminal)

# Stop database
cd infra/database
docker-compose down
```

---

**Happy Testing!** 🚀

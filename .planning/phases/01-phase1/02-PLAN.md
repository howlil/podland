# Phase 1 Plan: Foundation

**Phase:** 1 — Authentication + Basic Dashboard  
**Goal:** Users can authenticate via GitHub OAuth and access basic dashboard  
**Duration:** 3 weeks  
**Requirements:** 10 (AUTH-01 through AUTH-06, DASH-01, DASH-03, DASH-04)

---

## Success Criteria

1. ✅ User with @student.unand.ac.id GitHub email can sign in successfully
2. ✅ User with non-student email is rejected with clear error message
3. ✅ NIM containing "1152" is assigned Internal role, others get External
4. ✅ Session persists after browser refresh (7-day JWT expiry)
5. ✅ Profile page shows correct display name, role, and NIM
6. ✅ Sign out invalidates session and redirects to login
7. ✅ Dashboard displays "0 VMs" and quota usage (0/1 CPU for External)
8. ✅ Activity log shows "Account created" entry
9. ✅ Dashboard is usable on mobile (320px width) and desktop (1920px)

---

## Technical Milestones

- [ ] k3s cluster installed and running
- [ ] PostgreSQL database deployed
- [ ] Go backend with OAuth flow
- [ ] TanStack Start frontend deployed
- [ ] JWT authentication working
- [ ] NIM validation logic implemented

---

## Implementation Tasks

### Week 1: Infrastructure + Backend OAuth

#### Task 1.1: Project Setup
**Estimate:** 4 hours  
**Acceptance Criteria:**
- [ ] Monorepo structure created with Turborepo
- [ ] Go backend directory initialized (`/apps/backend`)
- [ ] TanStack Start frontend directory initialized (`/apps/frontend`)
- [ ] Shared types package created (`/packages/types`)
- [ ] Tailwind v4 configured in frontend
- [ ] Dark mode auto-detection working (system preference)

**Implementation:**
```bash
# Directory structure
podland/
├── apps/
│   ├── backend/          # Go service
│   └── frontend/         # TanStack Start
├── packages/
│   └── types/            # Shared TypeScript types
├── infra/
│   ├── k3s/              # Kubernetes manifests
│   └── database/         # PostgreSQL migration
└── turbo.json
```

---

#### Task 1.2: k3s Cluster Setup
**Estimate:** 3 hours  
**Acceptance Criteria:**
- [ ] k3s installed on server
- [ ] kubectl configured and working
- [ ] k3s service running (`systemctl status k3s`)
- [ ] Default namespace accessible

**Implementation:**
```bash
# Install k3s
curl -sfL https://get.k3s.io | sh -
sudo systemctl enable k3s
sudo systemctl start k3s

# Verify
kubectl cluster-info
kubectl get nodes
```

---

#### Task 1.3: PostgreSQL Database
**Estimate:** 3 hours  
**Acceptance Criteria:**
- [ ] PostgreSQL deployed in k3s via Helm
- [ ] Database `podland` created
- [ ] User `podland` with password created
- [ ] Connection string working from backend

**Implementation:**
```bash
# Deploy PostgreSQL via Helm
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install postgres bitnami/postgresql \
  --set auth.username=podland \
  --set auth.password=<secure-password> \
  --set auth.database=podland \
  --set persistence.size=10Gi

# Get connection string
kubectl get secret postgres-postgresql -o jsonpath='{.data.password}' | base64 -d
```

**Schema Migration:**
```sql
-- users table
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  github_id VARCHAR(255) UNIQUE NOT NULL,
  email VARCHAR(255) NOT NULL,
  display_name VARCHAR(255),
  avatar_url VARCHAR(512),
  nim VARCHAR(20) NOT NULL,
  role VARCHAR(20) NOT NULL CHECK (role IN ('internal', 'external', 'superadmin')),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_github_id ON users(github_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_nim ON users(nim);

-- sessions table
CREATE TABLE sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_token_hash VARCHAR(255) NOT NULL,
  device_info JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMP NOT NULL,
  revoked_at TIMESTAMP,
  replaced_by UUID REFERENCES sessions(id)
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- activity_logs table
CREATE TABLE activity_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  action VARCHAR(100) NOT NULL,
  metadata JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at);
```

---

#### Task 1.4: Go Backend OAuth Setup
**Estimate:** 6 hours  
**Acceptance Criteria:**
- [ ] `golang.org/x/oauth2` package installed
- [ ] GitHub OAuth app created (Client ID + Secret)
- [ ] OAuth config loaded from environment
- [ ] `/api/auth/login` endpoint generates authorization URL
- [ ] `/api/auth/github/callback` endpoint handles callback

**Implementation:**

**Environment Variables:**
```bash
# .env
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
GITHUB_CALLBACK_URL=http://localhost:8080/api/auth/github/callback
JWT_SECRET=your-jwt-secret-min-32-chars
REFRESH_TOKEN_SECRET=your-refresh-secret-min-32-chars
DATABASE_URL=postgresql://podland:password@localhost:5432/podland?sslmode=disable
FRONTEND_URL=http://localhost:3000
```

**OAuth Configuration (Go):**
```go
// backend/internal/auth/oauth.go
package auth

import (
    "context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/github"
)

var OAuthConfig = &oauth2.Config{
    ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
    ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
    Scopes:       []string{"user:email", "read:user"},
    Endpoint:     github.Endpoint,
    RedirectURL:  os.Getenv("GITHUB_CALLBACK_URL"),
}

func GetLoginURL(state string) string {
    return OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
    return OAuthConfig.Exchange(ctx, code)
}

func GetHTTPClient(ctx context.Context, token *oauth2.Token) *http.Client {
    return OAuthConfig.Client(ctx, token)
}
```

**Login Handler:**
```go
// backend/handlers/auth.go
func HandleLogin(w http.ResponseWriter, r *http.Request) {
    state := generateStateToken() // Random 32-char string
    setSessionCookie(w, "oauth_state", state, 5*time.Minute)
    
    url := auth.GetLoginURL(state)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
```

**Callback Handler:**
```go
func HandleCallback(w http.ResponseWriter, r *http.Request) {
    state := r.FormValue("state")
    oauthState := getCookie(r, "oauth_state")
    
    if state != oauthState {
        http.Error(w, "Invalid state parameter", http.StatusBadRequest)
        return
    }
    
    code := r.FormValue("code")
    token, err := auth.ExchangeToken(r.Context(), code)
    if err != nil {
        http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Fetch user info
    userInfo, err := fetchGitHubUser(r.Context(), token)
    if err != nil {
        http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
        return
    }
    
    // Fetch emails
    emails, err := fetchGitHubEmails(r.Context(), token)
    if err != nil {
        http.Error(w, "Failed to fetch emails", http.StatusInternalServerError)
        return
    }
    
    // Validate primary email
    primaryEmail := findPrimaryEmail(emails)
    if !isValidStudentEmail(primaryEmail) {
        // Redirect to rejection page
        http.Redirect(w, r, "/auth/rejected?reason=invalid_email", http.StatusTemporaryRedirect)
        return
    }
    
    // Extract NIM
    nim := extractNIM(primaryEmail)
    role := assignRole(nim)
    
    // Check if user exists
    user, err := db.GetUserByGitHubID(userInfo.ID)
    if err == sql.ErrNoRows {
        // Create new user
        user, err = db.CreateUser(&db.User{
            GitHubID:   userInfo.ID,
            Email:      primaryEmail,
            DisplayName: userInfo.Name,
            AvatarURL:  userInfo.AvatarURL,
            NIM:        nim,
            Role:       role,
        })
        if err != nil {
            http.Error(w, "Failed to create user", http.StatusInternalServerError)
            return
        }
        
        // Log activity
        db.CreateActivityLog(user.ID, "account_created", nil)
        
        // Redirect to welcome screen
        http.Redirect(w, r, fmt.Sprintf("/auth/welcome?userId=%s", user.ID), http.StatusTemporaryRedirect)
        return
    }
    
    // Existing user: create session and redirect to dashboard
    session, err := createSession(user)
    if err != nil {
        http.Error(w, "Failed to create session", http.StatusInternalServerError)
        return
    }
    
    setAuthCookies(w, session)
    http.Redirect(w, r, os.Getenv("FRONTEND_URL")+"/dashboard", http.StatusTemporaryRedirect)
}
```

**GitHub API Helpers:**
```go
type GitHubUser struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Login     string `json:"login"`
    AvatarURL string `json:"avatar_url"`
    Email     string `json:"email"`
}

type GitHubEmail struct {
    Email    string `json:"email"`
    Primary  bool   `json:"primary"`
    Verified bool   `json:"verified"`
    Visibility string `json:"visibility"`
}

func fetchGitHubUser(ctx context.Context, token *oauth2.Token) (*GitHubUser, error) {
    client := auth.GetHTTPClient(ctx, token)
    resp, err := client.Get("https://api.github.com/user")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var user GitHubUser
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }
    return &user, nil
}

func fetchGitHubEmails(ctx context.Context, token *oauth2.Token) ([]GitHubEmail, error) {
    client := auth.GetHTTPClient(ctx, token)
    resp, err := client.Get("https://api.github.com/user/emails")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var emails []GitHubEmail
    if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
        return nil, err
    }
    return emails, nil
}

func findPrimaryEmail(emails []GitHubEmail) string {
    for _, email := range emails {
        if email.Primary && email.Verified {
            return email.Email
        }
    }
    return ""
}

func isValidStudentEmail(email string) bool {
    return strings.HasSuffix(email, "@student.unand.ac.id")
}

func extractNIM(email string) string {
    // email format: NIM@student.unand.ac.id
    parts := strings.Split(email, "@")
    if len(parts) == 0 {
        return ""
    }
    return parts[0]
}

func assignRole(nim string) string {
    if strings.Contains(nim, "1152") {
        return "internal"
    }
    return "external"
}
```

---

#### Task 1.5: JWT + Refresh Token System
**Estimate:** 8 hours  
**Acceptance Criteria:**
- [ ] Access token (JWT, 15min expiry) generation working
- [ ] Refresh token (opaque, 7 days) generation working
- [ ] HTTP-only cookies set correctly
- [ ] `/api/auth/refresh` endpoint rotates tokens
- [ ] Token verification middleware validates JWT
- [ ] Max 3 concurrent sessions enforced

**Implementation:**

**Token Generation:**
```go
// backend/internal/auth/jwt.go
package auth

import (
    "crypto/sha256"
    "encoding/base64"
    "encoding/hex"
    "errors"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

var (
    ErrInvalidToken = errors.New("invalid token")
    ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func GenerateAccessToken(userID, email string) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "podland",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func GenerateRefreshToken() (string, string) {
    // Generate random opaque token
    bytes := make([]byte, 32)
    rand.Read(bytes)
    refreshToken := base64.URLEncoding.EncodeToString(bytes)
    
    // Generate JTI
    jtiBytes := make([]byte, 16)
    rand.Read(jtiBytes)
    jti := hex.EncodeToString(jtiBytes)
    
    return refreshToken, jti
}

func HashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }
    
    if claims.ExpiresAt.Before(time.Now()) {
        return nil, ErrExpiredToken
    }
    
    return claims, nil
}
```

**Session Management:**
```go
// backend/internal/auth/session.go
type Session struct {
    ID           string
    UserID       string
    RefreshToken string // Raw token (hash for DB)
    JTI          string
    DeviceInfo   DeviceInfo
    CreatedAt    time.Time
    ExpiresAt    time.Time
}

type DeviceInfo struct {
    UserAgent string `json:"user_agent"`
    IP        string `json:"ip"`
}

func CreateSession(userID string, deviceInfo DeviceInfo) (*Session, error) {
    // Check concurrent session limit (max 3)
    count, err := db.GetActiveSessionCount(userID)
    if err != nil {
        return nil, err
    }
    
    if count >= 3 {
        // Revoke oldest session
        err := db.RevokeOldestSession(userID)
        if err != nil {
            return nil, err
        }
    }
    
    refreshToken, jti := GenerateRefreshToken()
    
    session := &Session{
        ID:           uuid.New().String(),
        UserID:       userID,
        RefreshToken: refreshToken,
        JTI:          jti,
        DeviceInfo:   deviceInfo,
        CreatedAt:    time.Now(),
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
    }
    
    // Store in database
    err = db.CreateSession(&db.Session{
        ID:             session.ID,
        UserID:         session.UserID,
        RefreshTokenHash: HashToken(session.RefreshToken),
        JTI:            session.JTI,
        DeviceInfo:     session.DeviceInfo,
        CreatedAt:      session.CreatedAt,
        ExpiresAt:      session.ExpiresAt,
    })
    
    if err != nil {
        return nil, err
    }
    
    return session, nil
}

func RotateRefreshToken(oldToken string) (*Session, error) {
    oldHash := HashToken(oldToken)
    
    // Find and validate old token
    oldSession, err := db.GetSessionByRefreshToken(oldHash)
    if err != nil {
        return nil, ErrInvalidToken
    }
    
    if oldSession.ExpiresAt.Before(time.Now()) {
        return nil, ErrExpiredToken
    }
    
    if oldSession.RevokedAt != nil {
        // Token reuse detected - security alert
        db.RevokeAllUserSessions(oldSession.UserID)
        return nil, errors.New("token reuse detected")
    }
    
    // Revoke old token
    err = db.RevokeSession(oldSession.ID, time.Now())
    if err != nil {
        return nil, err
    }
    
    // Create new session
    newSession, err := CreateSession(oldSession.UserID, oldSession.DeviceInfo)
    if err != nil {
        return nil, err
    }
    
    // Link old to new
    db.LinkSessionReplacement(oldSession.ID, newSession.ID)
    
    return newSession, nil
}
```

**Cookie Helpers:**
```go
// backend/handlers/auth.go
func setAuthCookies(w http.ResponseWriter, session *Session) {
    // Generate access token
    user, _ := db.GetUserByID(session.UserID)
    accessToken, _ := GenerateAccessToken(user.ID, user.Email)
    
    // Set refresh token cookie (HTTP-only)
    refreshCookie := &http.Cookie{
        Name:     "refresh_token",
        Value:    session.RefreshToken,
        Path:     "/api/auth/refresh",
        HttpOnly: true,
        Secure:   os.Getenv("ENV") == "production",
        SameSite: http.SameSiteStrictMode,
        Expires:  session.ExpiresAt,
    }
    http.SetCookie(w, refreshCookie)
    
    // Set XSRF token cookie (JavaScript-readable)
    xsrfToken := generateXSRFToken()
    xsrfCookie := &http.Cookie{
        Name:     "XSRF-TOKEN",
        Value:    xsrfToken,
        Path:     "/",
        HttpOnly: false, // JS needs to read this
        Secure:   os.Getenv("ENV") == "production",
        SameSite: http.SameSiteStrictMode,
        Expires:  session.ExpiresAt,
    }
    http.SetCookie(w, xsrfCookie)
    
    // Return access token in response body (for in-memory storage)
    // Frontend stores this in React state, NOT localStorage
}

func clearAuthCookies(w http.ResponseWriter) {
    refreshCookie := &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Path:     "/api/auth/refresh",
        HttpOnly: true,
        Secure:   os.Getenv("ENV") == "production",
        SameSite: http.SameSiteStrictMode,
        Expires:  time.Unix(0, 0),
        MaxAge:   -1,
    }
    http.SetCookie(w, refreshCookie)
    
    xsrfCookie := &http.Cookie{
        Name:     "XSRF-TOKEN",
        Value:    "",
        Path:     "/",
        HttpOnly: false,
        Secure:   os.Getenv("ENV") == "production",
        SameSite: http.SameSiteStrictMode,
        Expires:  time.Unix(0, 0),
        MaxAge:   -1,
    }
    http.SetCookie(w, xsrfCookie)
}
```

**Refresh Endpoint:**
```go
func HandleRefresh(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get refresh token from cookie
    refreshCookie, err := r.Cookie("refresh_token")
    if err != nil {
        http.Error(w, "Missing refresh token", http.StatusUnauthorized)
        return
    }
    
    // Rotate token
    newSession, err := RotateRefreshToken(refreshCookie.Value)
    if err != nil {
        http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }
    
    // Set new cookies
    setAuthCookies(w, newSession)
    
    // Return new access token
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "access_token": accessToken,
    })
}
```

**Auth Middleware:**
```go
// backend/middleware/auth.go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get access token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := ValidateAccessToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
            return
        }
        
        // Add user to context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "email", claims.Email)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// CSRF Middleware
func CSRFMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip CSRF for GET, HEAD, OPTIONS
        if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
            next.ServeHTTP(w, r)
            return
        }
        
        // Validate XSRF-TOKEN header
        xsrfHeader := r.Header.Get("X-XSRF-TOKEN")
        xsrfCookie, err := r.Cookie("XSRF-TOKEN")
        if err != nil || xsrfHeader != xsrfCookie.Value {
            http.Error(w, "Invalid CSRF token", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

### Week 2: Frontend + Dashboard

#### Task 2.1: TanStack Start Setup
**Estimate:** 4 hours  
**Acceptance Criteria:**
- [ ] TanStack Start project initialized
- [ ] Tailwind v4 configured with dark mode
- [ ] Base layout component created
- [ ] Route structure set up
- [ ] API client configured with auth handling

**Implementation:**

```bash
# Create TanStack Start app
npm create tanstack@latest start -- --template react
cd frontend
npm install
```

**Tailwind v4 Config (CSS):**
```css
/* src/styles/index.css */
@import "tailwindcss";

@theme {
  --color-primary: #3b82f6;
  --color-primary-dark: #2563eb;
}

/* Dark mode via system preference */
@media (prefers-color-scheme: dark) {
  :root {
    color-scheme: dark;
  }
}
```

**Route Structure:**
```
src/
├── routes/
│   ├── __root.tsx
│   ├── index.tsx              # Landing/login page
│   ├── auth/
│   │   ├── callback.tsx       # OAuth callback
│   │   ├── welcome.tsx        # First-time user welcome
│   │   └── rejected.tsx       # Email rejection page
│   ├── dashboard/
│   │   ├── index.tsx          # Dashboard home
│   │   └── profile.tsx        # User profile
│   └── api/
│       └── trpc/
│           └── [trpc].ts      # tRPC API routes
├── components/
│   ├── layout/
│   │   ├── Sidebar.tsx
│   │   ├── MobileTabBar.tsx
│   │   └── DashboardLayout.tsx
│   ├── dashboard/
│   │   ├── QuotaUsage.tsx
│   │   ├── VMCountCard.tsx
│   │   └── ActivityLog.tsx
│   └── auth/
│       ├── SignInButton.tsx
│       └── WelcomeScreen.tsx
├── lib/
│   ├── api.ts                 # API client
│   ├── auth.ts                # Auth hooks
│   └── utils.ts
└── styles/
    └── index.css
```

---

#### Task 2.2: Authentication UI Components
**Estimate:** 6 hours  
**Acceptance Criteria:**
- [ ] Sign in button component
- [ ] OAuth callback handler
- [ ] Welcome screen with NIM confirmation
- [ ] Rejection page with guide
- [ ] Auth hook for token management

**Implementation:**

**Sign In Button:**
```tsx
// src/components/auth/SignInButton.tsx
export function SignInButton() {
  return (
    <a
      href="/api/auth/login"
      className="inline-flex items-center gap-2 px-6 py-3 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 rounded-lg hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors"
    >
      <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
      </svg>
      Sign in with GitHub
    </a>
  );
}
```

**Welcome Screen:**
```tsx
// src/routes/auth/welcome.tsx
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import axios from 'axios'

export const Route = createFileRoute('/auth/welcome')({
  component: WelcomeScreen,
})

function WelcomeScreen() {
  const navigate = useNavigate()
  const [nim, setNim] = useState('')
  const [extractedNim, setExtractedNim] = useState('')
  const [role, setRole] = useState('')
  const [termsAccepted, setTermsAccepted] = useState(false)
  const [isEditing, setIsEditing] = useState(false)

  useEffect(() => {
    // Get user data from callback
    const params = new URLSearchParams(window.location.search)
    const userId = params.get('userId')
    if (userId) {
      // Fetch user data
      axios.get(`/api/users/${userId}`).then((res) => {
        setExtractedNim(res.data.nim)
        setNim(res.data.nim)
        setRole(res.data.role)
      })
    }
  }, [])

  const handleConfirm = async () => {
    await axios.post('/api/users/confirm-nim', { nim })
    navigate({ to: '/dashboard' })
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="max-w-md w-full p-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
          Welcome to Podland!
        </h1>
        
        <div className="space-y-4">
          <div>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Your student ID (NIM) has been extracted from your email.
            </p>
          </div>

          {!isEditing ? (
            <div className="p-4 bg-gray-100 dark:bg-gray-700 rounded-lg">
              <p className="text-lg font-semibold text-gray-900 dark:text-white">
                NIM: {extractedNim}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Role: {role === 'internal' ? 'Internal (SI UNAND)' : 'External'}
              </p>
            </div>
          ) : (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Enter your NIM
              </label>
              <input
                type="text"
                value={nim}
                onChange={(e) => setNim(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                placeholder="221152001"
              />
            </div>
          )}

          <div className="flex gap-2">
            <button
              onClick={() => setIsEditing(!isEditing)}
              className="text-sm text-primary hover:underline"
            >
              {isEditing ? 'Cancel' : 'Edit NIM'}
            </button>
          </div>

          <div className="flex items-start gap-2">
            <input
              type="checkbox"
              id="terms"
              checked={termsAccepted}
              onChange={(e) => setTermsAccepted(e.target.checked)}
              className="mt-1"
            />
            <label htmlFor="terms" className="text-sm text-gray-600 dark:text-gray-400">
              I agree to the{' '}
              <a href="/terms" target="_blank" className="text-primary hover:underline">
                Terms of Service
              </a>{' '}
              and confirm I am a current Unand student.
            </label>
          </div>

          <button
            onClick={handleConfirm}
            disabled={!termsAccepted}
            className="w-full py-2 px-4 bg-primary text-white rounded-lg hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            Activate Account
          </button>
        </div>
      </div>
    </div>
  )
}
```

**Rejection Page:**
```tsx
// src/routes/auth/rejected.tsx
import { createFileRoute, Link } from '@tanstack/react-router'

export const Route = createFileRoute('/auth/rejected')({
  component: RejectedPage,
})

function RejectedPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="max-w-md w-full p-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg text-center">
        <div className="w-16 h-16 mx-auto mb-4 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center">
          <svg className="w-8 h-8 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Access Denied
        </h1>
        
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Your GitHub email must end with <code className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded">@student.unand.ac.id</code>
        </p>

        <div className="text-left p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg mb-6">
          <h2 className="font-semibold text-gray-900 dark:text-white mb-2">
            How to fix this:
          </h2>
          <ol className="list-decimal list-inside space-y-2 text-sm text-gray-600 dark:text-gray-400">
            <li>Go to your{' '}
              <a href="https://github.com/settings/emails" target="_blank" className="text-primary hover:underline">
                GitHub Email Settings
              </a>
            </li>
            <li>Add your <code className="px-1">@student.unand.ac.id</code> email</li>
            <li>Verify the email address</li>
            <li>Set it as your primary email</li>
            <li>Click "Retry" below</li>
          </ol>
        </div>

        <a
          href="/api/auth/login"
          className="inline-block px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
        >
          Retry with Updated Email
        </a>
      </div>
    </div>
  )
}
```

---

#### Task 2.3: Dashboard Layout
**Estimate:** 6 hours  
**Acceptance Criteria:**
- [ ] Sidebar navigation (desktop)
- [ ] Bottom tab bar (mobile)
- [ ] Responsive layout working
- [ ] Dark mode auto-detection
- [ ] User avatar dropdown

**Implementation:**

**Dashboard Layout:**
```tsx
// src/components/layout/DashboardLayout.tsx
export function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Desktop Sidebar */}
      <aside className="hidden md:fixed md:inset-y-0 md:flex md:w-64 md:flex-col">
        <div className="flex flex-col flex-grow pt-5 overflow-y-auto bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
          <div className="flex items-center flex-shrink-0 px-4 mb-5">
            <h1 className="text-xl font-bold text-gray-900 dark:text-white">Podland</h1>
          </div>
          
          <nav className="flex-1 px-2 space-y-1">
            <Link
              to="/dashboard"
              className="group flex items-center px-2 py-2 text-sm font-medium rounded-md text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-700"
            >
              <span className="mr-3">📊</span>
              Dashboard
            </Link>
            <Link
              to="/dashboard/profile"
              className="group flex items-center px-2 py-2 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              <span className="mr-3">👤</span>
              Profile
            </Link>
          </nav>

          {/* User Avatar Dropdown */}
          <div className="flex items-center p-4 border-t border-gray-200 dark:border-gray-700">
            <UserDropdown />
          </div>
        </div>
      </aside>

      {/* Mobile Bottom Tab Bar */}
      <nav className="md:hidden fixed bottom-0 inset-x-0 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
        <div className="grid grid-cols-2 h-16">
          <Link
            to="/dashboard"
            className="flex flex-col items-center justify-center text-gray-900 dark:text-white"
          >
            <span className="text-xl">📊</span>
            <span className="text-xs mt-1">Dashboard</span>
          </Link>
          <Link
            to="/dashboard/profile"
            className="flex flex-col items-center justify-center text-gray-700 dark:text-gray-300"
          >
            <span className="text-xl">👤</span>
            <span className="text-xs mt-1">Profile</span>
          </Link>
        </div>
      </nav>

      {/* Main Content */}
      <main className="md:pl-64 pb-16 md:pb-0">
        <div className="px-4 py-6 sm:px-6">
          {children}
        </div>
      </main>
    </div>
  )
}
```

**User Dropdown:**
```tsx
// src/components/layout/UserDropdown.tsx
export function UserDropdown() {
  const { user, logout } = useAuth()
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 focus:outline-none"
      >
        <img
          src={user?.avatarUrl}
          alt={user?.displayName}
          className="w-8 h-8 rounded-full"
        />
        <span className="text-sm font-medium text-gray-700 dark:text-gray-300 hidden lg:block">
          {user?.displayName}
        </span>
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute bottom-full left-0 mb-2 w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg py-1 border border-gray-200 dark:border-gray-700">
          <div className="px-4 py-2 border-b border-gray-200 dark:border-gray-700">
            <p className="text-sm font-medium text-gray-900 dark:text-white">{user?.displayName}</p>
            <p className="text-xs text-gray-500 dark:text-gray-400 truncate">{user?.email}</p>
          </div>
          <button
            onClick={logout}
            className="w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            Sign out
          </button>
        </div>
      )}
    </div>
  )
}
```

---

#### Task 2.4: Dashboard Widgets
**Estimate:** 6 hours  
**Acceptance Criteria:**
- [ ] Quota usage bar component
- [ ] VM count card component
- [ ] Activity log component
- [ ] Responsive grid layout
- [ ] Dark mode styling

**Implementation:**

**Dashboard Home:**
```tsx
// src/routes/dashboard/index.tsx
export const Route = createFileRoute('/dashboard/')({
  component: DashboardHome,
})

function DashboardHome() {
  const { user } = useAuth()
  const { data: activity } = useQuery({
    queryKey: ['activity'],
    queryFn: () => axios.get('/api/activity').then(r => r.data),
  })

  const quota = user?.role === 'internal' 
    ? { cpu: 1, ram: 2048 } 
    : { cpu: 0.5, ram: 1024 }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
          Dashboard
        </h1>

        {/* Quota Usage */}
        <QuotaUsageCard 
          usedCpu={0} 
          maxCpu={quota.cpu}
          usedRam={0}
          maxRam={quota.ram}
        />

        {/* VM Count + Activity Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <VMCountCard count={0} />
          <ActivityLog activities={activity} className="lg:col-span-2" />
        </div>
      </div>
    </DashboardLayout>
  )
}
```

**Quota Usage Component:**
```tsx
// src/components/dashboard/QuotaUsage.tsx
export function QuotaUsageCard({ usedCpu, maxCpu, usedRam, maxRam }: QuotaProps) {
  const cpuPercent = (usedCpu / maxCpu) * 100
  const ramPercent = (usedRam / maxRam) * 100

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Quota Usage
      </h2>
      
      <div className="space-y-4">
        {/* CPU */}
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">CPU</span>
            <span className="text-sm text-gray-600 dark:text-gray-400">
              {usedCpu.toFixed(2)} / {maxCpu} cores
            </span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5">
            <div
              className={`h-2.5 rounded-full transition-all ${
                cpuPercent > 90 ? 'bg-red-600' : cpuPercent > 70 ? 'bg-yellow-500' : 'bg-green-500'
              }`}
              style={{ width: `${cpuPercent}%` }}
            />
          </div>
        </div>

        {/* RAM */}
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">RAM</span>
            <span className="text-sm text-gray-600 dark:text-gray-400">
              {(usedRam / 1024).toFixed(1)} / {(maxRam / 1024).toFixed(1)} GB
            </span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5">
            <div
              className={`h-2.5 rounded-full transition-all ${
                ramPercent > 90 ? 'bg-red-600' : ramPercent > 70 ? 'bg-yellow-500' : 'bg-green-500'
              }`}
              style={{ width: `${ramPercent}%` }}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
```

**VM Count Card:**
```tsx
// src/components/dashboard/VMCountCard.tsx
export function VMCountCard({ count }: { count: number }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <div className="flex items-center">
        <div className="flex-shrink-0 w-12 h-12 bg-blue-100 dark:bg-blue-900/20 rounded-lg flex items-center justify-center">
          <span className="text-2xl">💻</span>
        </div>
        <div className="ml-4">
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400">VMs Running</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{count}</p>
        </div>
      </div>
    </div>
  )
}
```

**Activity Log:**
```tsx
// src/components/dashboard/ActivityLog.tsx
export function ActivityLog({ activities, className }: ActivityProps) {
  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg shadow p-6 ${className}`}>
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Recent Activity
      </h2>
      
      <div className="space-y-3">
        {activities?.map((activity: Activity) => (
          <div key={activity.id} className="flex items-start gap-3">
            <div className="w-2 h-2 mt-2 bg-gray-400 rounded-full" />
            <div className="flex-1">
              <p className="text-sm text-gray-900 dark:text-white">
                {formatActivityText(activity.action)}
              </p>
              <p className="text-xs text-gray-500 dark:text-gray-400">
                {formatDistance(new Date(activity.createdAt), new Date())} ago
              </p>
            </div>
          </div>
        ))}
        
        {activities?.length === 0 && (
          <p className="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
            No recent activity
          </p>
        )}
      </div>
    </div>
  )
}
```

---

#### Task 2.5: Profile Page
**Estimate:** 4 hours  
**Acceptance Criteria:**
- [ ] Profile displays all user fields
- [ ] Avatar shows correctly
- [ ] NIM and role displayed
- [ ] Read-only fields
- [ ] Responsive layout

**Implementation:**

```tsx
// src/routes/dashboard/profile.tsx
export const Route = createFileRoute('/dashboard/profile')({
  component: ProfilePage,
})

function ProfilePage() {
  const { user } = useAuth()

  return (
    <DashboardLayout>
      <div className="max-w-2xl">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
          Profile
        </h1>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          {/* Avatar */}
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center gap-4">
              <img
                src={user?.avatarUrl}
                alt={user?.displayName}
                className="w-20 h-20 rounded-full"
              />
              <div>
                <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                  {user?.displayName}
                </h2>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {user?.email}
                </p>
              </div>
            </div>
          </div>

          {/* Profile Fields */}
          <div className="p-6 space-y-4">
            <ProfileField label="Display Name" value={user?.displayName} />
            <ProfileField label="Email" value={user?.email} readOnly />
            <ProfileField label="NIM" value={user?.nim} readOnly />
            <ProfileField 
              label="Role" 
              value={user?.role === 'internal' ? 'Internal (SI UNAND)' : 'External'} 
              readOnly 
            />
            <ProfileField 
              label="Member Since" 
              value={formatDate(user?.createdAt)} 
              readOnly 
            />
          </div>
        </div>
      </div>
    </DashboardLayout>
  )
}

function ProfileField({ label, value, readOnly = true }: ProfileFieldProps) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
        {label}
      </label>
      <input
        type="text"
        value={value || ''}
        readOnly={readOnly}
        className={`w-full px-3 py-2 border rounded-md ${
          readOnly 
            ? 'bg-gray-50 dark:bg-gray-700 text-gray-500 dark:text-gray-400' 
            : 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white'
        }`}
      />
    </div>
  )
}
```

---

#### Task 2.6: Auth Hook + API Client
**Estimate:** 4 hours  
**Acceptance Criteria:**
- [ ] `useAuth()` hook provides user context
- [ ] Auto refresh at 50% expiry
- [ ] 401 interceptor triggers refresh
- [ ] Logout clears state and cookies
- [ ] Axios interceptors handle tokens

**Implementation:**

**API Client:**
```ts
// src/lib/api.ts
import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  withCredentials: true, // Send cookies
})

// Add access token to requests
api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  
  // Add CSRF token
  const xsrfToken = document.cookie
    .split('; ')
    .find(row => row.startsWith('XSRF-TOKEN='))
    ?.split('=')[1]
  
  if (xsrfToken) {
    config.headers['X-XSRF-TOKEN'] = xsrfToken
  }
  
  return config
})

// Handle 401 responses
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      
      try {
        const response = await axios.post('/api/auth/refresh', null, {
          withCredentials: true,
        })
        
        const { access_token } = response.data
        setAccessToken(access_token)
        
        originalRequest.headers.Authorization = `Bearer ${access_token}`
        return api(originalRequest)
      } catch (refreshError) {
        // Refresh failed - redirect to login
        window.location.href = '/api/auth/login'
        return Promise.reject(refreshError)
      }
    }
    
    return Promise.reject(error)
  }
)

export default api
```

**Auth Hook:**
```ts
// src/lib/auth.ts
import { create } from 'zustand'
import api from './api'

interface User {
  id: string
  githubId: string
  email: string
  displayName: string
  avatarUrl: string
  nim: string
  role: 'internal' | 'external' | 'superadmin'
  createdAt: string
}

interface AuthState {
  user: User | null
  isLoading: boolean
  login: () => void
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

let accessToken: string | null = null
let refreshTimer: NodeJS.Timeout | null = null

export const useAuth = create<AuthState>((set, get) => ({
  user: null,
  isLoading: true,
  
  login: () => {
    window.location.href = '/api/auth/login'
  },
  
  logout: async () => {
    await api.post('/auth/logout')
    accessToken = null
    if (refreshTimer) clearTimeout(refreshTimer)
    set({ user: null, isLoading: false })
  },
  
  refreshUser: async () => {
    try {
      const { data } = await api.get('/users/me')
      set({ user: data, isLoading: false })
      
      // Schedule silent refresh at 50% expiry (7.5 minutes for 15-min token)
      if (refreshTimer) clearTimeout(refreshTimer)
      refreshTimer = setTimeout(() => {
        api.post('/auth/refresh').then(({ data }) => {
          accessToken = data.access_token
        })
      }, 7.5 * 60 * 1000)
    } catch (error) {
      set({ user: null, isLoading: false })
    }
  },
}))

// Helper functions
export function setAccessToken(token: string) {
  accessToken = token
}

export function getAccessToken(): string | null {
  return accessToken
}
```

---

### Week 3: Integration + Testing

#### Task 3.1: Avatar Download System
**Estimate:** 4 hours  
**Acceptance Criteria:**
- [ ] Avatar downloaded on sign-in
- [ ] Stored in `/uploads/avatars/{userId}.{ext}`
- [ ] GitHub URL hash compared for changes
- [ ] Fallback to hotlink if download fails

**Implementation:**

```go
// backend/internal/avatar/service.go
package avatar

import (
    "crypto/md5"
    "encoding/hex"
    "io"
    "net/http"
    "os"
    "path/filepath"
)

const AvatarDir = "./uploads/avatars"

type Service struct {
    baseURL string
}

func NewService() *Service {
    os.MkdirAll(AvatarDir, 0755)
    return &Service{baseURL: "https://avatars.githubusercontent.com"}
}

func (s *Service) SyncAvatar(userID, githubAvatarURL string) (string, error) {
    // Extract hash from GitHub URL
    githubHash := extractAvatarHash(githubAvatarURL)
    
    // Check local file
    localPath := filepath.Join(AvatarDir, userID+".jpg")
    localHash, err := s.getFileHash(localPath)
    
    if err == nil && localHash == githubHash {
        // Already synced
        return "/uploads/avatars/" + userID + ".jpg", nil
    }
    
    // Download new avatar
    resp, err := http.Get(githubAvatarURL)
    if err != nil {
        // Fallback to hotlink
        return githubAvatarURL, nil
    }
    defer resp.Body.Close()
    
    file, err := os.Create(localPath)
    if err != nil {
        return githubAvatarURL, nil
    }
    defer file.Close()
    
    _, err = io.Copy(file, resp.Body)
    if err != nil {
        return githubAvatarURL, nil
    }
    
    return "/uploads/avatars/" + userID + ".jpg", nil
}

func extractAvatarHash(url string) string {
    // GitHub URL: https://avatars.githubusercontent.com/u/12345?v=4&hash=abc123
    // Extract hash or use full URL hash
    h := md5.Sum([]byte(url))
    return hex.EncodeToString(h[:])
}

func (s *Service) getFileHash(path string) (string, error) {
    file, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    hash := md5.New()
    io.Copy(hash, file)
    return hex.EncodeToString(hash.Sum(nil)), nil
}
```

---

#### Task 3.2: Activity Logging
**Estimate:** 2 hours  
**Acceptance Criteria:**
- [ ] Activity logged on account creation
- [ ] Activity logged on sign-in
- [ ] Activity logged on sign-out
- [ ] Activity API endpoint returns last 50 entries

**Implementation:**

```go
// backend/internal/activity/service.go
package activity

type Action string

const (
    AccountCreated Action = "account_created"
    SignedIn       Action = "signed_in"
    SignedOut      Action = "signed_out"
)

func LogActivity(userID string, action Action, metadata map[string]interface{}) error {
    query := `
        INSERT INTO activity_logs (user_id, action, metadata, created_at)
        VALUES ($1, $2, $3, NOW())
    `
    _, err := db.Exec(query, userID, action, metadata)
    return err
}

func GetUserActivity(userID string, limit int) ([]Activity, error) {
    query := `
        SELECT id, user_id, action, metadata, created_at
        FROM activity_logs
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2
    `
    rows, err := db.Query(query, userID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var activities []Activity
    for rows.Next() {
        var a Activity
        err := rows.Scan(&a.ID, &a.UserID, &a.Action, &a.Metadata, &a.CreatedAt)
        if err != nil {
            return nil, err
        }
        activities = append(activities, a)
    }
    return activities, nil
}
```

---

#### Task 3.3: End-to-End Testing
**Estimate:** 8 hours  
**Acceptance Criteria:**
- [ ] OAuth flow tested (happy path + rejection)
- [ ] Session persistence tested (browser refresh)
- [ ] Token refresh tested (silent + on-demand)
- [ ] Dashboard renders correctly (mobile + desktop)
- [ ] Profile displays correct data
- [ ] Logout invalidates session

**Test Cases:**

```go
// backend/tests/auth_test.go
func TestGitHubOAuth_Success(t *testing.T) {
    // Mock GitHub API
    mockGitHubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/user" {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "id": "12345",
                "name": "Test User",
                "avatar_url": "https://avatars.githubusercontent.com/u/12345",
            })
        } else if r.URL.Path == "/user/emails" {
            json.NewEncoder(w).Encode([]map[string]interface{}{
                {"email": "221152001@student.unand.ac.id", "primary": true, "verified": true},
            })
        }
    }))
    defer mockGitHubServer.Close()

    // Test OAuth callback
    resp := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/api/auth/github/callback?code=test&state=test", nil)
    
    handler := http.HandlerFunc(HandleCallback)
    handler.ServeHTTP(resp, req)
    
    assert.Equal(t, http.StatusTemporaryRedirect, resp.Code)
    assert.Contains(t, resp.Header().Get("Location"), "/dashboard")
    
    // Verify cookies set
    cookies := resp.Result().Cookies()
    assert.Len(t, cookies, 2) // refresh_token + XSRF-TOKEN
}

func TestGitHubOAuth_RejectedEmail(t *testing.T) {
    // Mock GitHub API with non-student email
    mockGitHubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/user/emails" {
            json.NewEncoder(w).Encode([]map[string]interface{}{
                {"email": "user@gmail.com", "primary": true, "verified": true},
            })
        }
    }))
    defer mockGitHubServer.Close()

    resp := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/api/auth/github/callback?code=test&state=test", nil)
    
    handler := http.HandlerFunc(HandleCallback)
    handler.ServeHTTP(resp, req)
    
    assert.Equal(t, http.StatusTemporaryRedirect, resp.Code)
    assert.Contains(t, resp.Header().Get("Location"), "/auth/rejected")
}

func TestTokenRefresh(t *testing.T) {
    // Create user and session
    user := createUser(t)
    session := createSession(t, user.ID)
    
    // Request refresh with valid token
    resp := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/api/auth/refresh", nil)
    req.AddCookie(&http.Cookie{
        Name:  "refresh_token",
        Value: session.RefreshToken,
    })
    
    handler := http.HandlerFunc(HandleRefresh)
    handler.ServeHTTP(resp, req)
    
    assert.Equal(t, http.StatusOK, resp.Code)
    
    var data map[string]string
    json.NewDecoder(resp.Body).Decode(&data)
    assert.NotEmpty(t, data["access_token"])
}
```

---

#### Task 3.4: Responsive Testing
**Estimate:** 3 hours  
**Acceptance Criteria:**
- [ ] Dashboard tested at 320px (mobile)
- [ ] Dashboard tested at 768px (tablet)
- [ ] Dashboard tested at 1920px (desktop)
- [ ] Bottom tab bar visible on mobile only
- [ ] Sidebar visible on desktop only
- [ ] All widgets stack correctly on mobile

**Manual Testing Checklist:**

- [ ] Mobile (320px): Bottom tab bar visible, widgets stacked vertically
- [ ] Mobile (320px): User dropdown works
- [ ] Mobile (320px): Profile page readable
- [ ] Tablet (768px): Sidebar visible, 2-column widget grid
- [ ] Desktop (1920px): Sidebar visible, 2-column widget grid
- [ ] Dark mode: All components styled correctly
- [ ] Light mode: All components styled correctly

---

#### Task 3.5: Bug Fixes + Polish
**Estimate:** 5 hours  
**Acceptance Criteria:**
- [ ] All console errors fixed
- [ ] Loading states added
- [ ] Error messages user-friendly
- [ ] Accessibility checked (keyboard nav, screen reader)
- [ ] Performance optimized (lazy loading, code splitting)

---

## Risk Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| GitHub OAuth rate limiting | Medium | Low | Cache user info, exponential backoff |
| k3s cluster instability | High | Low | Use stable k3s version, test locally first |
| TanStack Start bugs | Medium | Low | Fallback to Vite + React Router if needed |
| JWT token leakage | Critical | Low | HTTP-only cookies, CSRF protection, security audit |
| Email validation false positives | Medium | Low | Manual override via admin panel (Phase 5) |

---

## Definition of Done

Phase 1 is complete when:

- ✅ All 10 requirements implemented (AUTH-01 through AUTH-06, DASH-01, DASH-03, DASH-04)
- ✅ All 9 success criteria verified
- ✅ All tests passing (unit + integration + e2e)
- ✅ No critical/high severity security issues
- ✅ Dashboard usable on mobile (320px) and desktop (1920px)
- ✅ Code reviewed and merged to main branch

---

## Next Phase

After Phase 1 completion, proceed to **Phase 2: Core VM** (VM lifecycle management + quotas).

---

*Plan created: 2026-03-25*  
*Ready for: Implementation*

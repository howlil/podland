# ARCHITECTURE.md — Integration Patterns for v1.1

**Milestone:** v1.1 — Hardening & Polish
**Date:** 2026-03-30
**Current Architecture:** Modular monolith (Go 1.25 + chi + React 18 + k3s)

---

## Executive Summary

v1.1 introduces **6 new integration patterns** across the existing architecture. This document details where each feature fits, how data flows through the system, new vs modified components, and the suggested build order.

| Feature | Integration Point | New Components | Modified Components |
|---------|------------------|----------------|---------------------|
| Rate Limiting | Middleware layer | `ratelimit.go` | `cmd/main.go` |
| GDPR Deletion | Use case + repository | `account_deletion.go` | `user.go`, migrations |
| OpenAPI | Handler annotations + docs | `docs/` directory | All handlers |
| WCAG AA | Frontend components | `a11y.ts`, test utils | All components, CSS |
| Lighthouse CI | CI/CD pipeline | Workflow, config | Build process |
| Load Testing | Test infrastructure | Enhanced k6 scripts | Test environment |

---

## 1. Rate Limiting Architecture

### Where It Fits

```
┌─────────────────────────────────────────────────────────────┐
│                      Request Flow                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Client → Cloudflare → Traefik → [Rate Limit Middleware]   │
│                                              │              │
│                                              ▼              │
│                                    ┌─────────────────┐     │
│                                    │  Auth Endpoints │     │
│                                    │  (5 req/min)    │     │
│                                    └─────────────────┘     │
│                                              │              │
│                                              ▼              │
│                                    ┌─────────────────┐     │
│                                    │  Other Endpoints│     │
│                                    │  (100 req/min)  │     │
│                                    └─────────────────┘     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Component Integration

```
apps/backend/
├── cmd/
│   └── main.go              # MODIFY: Add rate limit middleware
├── internal/
│   ├── middleware/
│   │   ├── middleware.go    # Existing: CORS, CSRF, Auth
│   │   └── ratelimit.go     # NEW: Rate limiting middleware
│   └── handler/
│       └── auth/
│           └── oauth.go     # PROTECTED: Rate limited endpoint
```

### Middleware Stack Order

```go
// cmd/main.go
r := chi.NewRouter()

// Order matters! Middleware executes in registration order:
r.Use(middleware.CORSMiddleware)           // 1. CORS (outermost)
r.Use(middleware.CSRFMiddleware)           // 2. CSRF
r.Use(middleware.RecoveryMiddleware)       // 3. Panic recovery

// Rate limiting applied selectively:
r.Route("/api/auth", func(r chi.Router) {
    r.Use(middleware.NewAuthRateLimiter()) // 4. Auth rate limit (5/min)
    r.Get("/login", auth.LoginHandler)
    r.Get("/callback", auth.CallbackHandler)
})

r.Route("/api", func(r chi.Router) {
    r.Use(middleware.NewAPIRateLimiter())  // 5. API rate limit (100/min)
    r.Use(middleware.AuthMiddleware)       // 6. Auth validation
    // ... other routes
})
```

### Data Flow: Rate Limited Request

```
1. Request arrives at /api/auth/login
2. Rate limiter middleware extracts IP from X-Forwarded-For
3. Check counter in memory store
   ├─ If under limit: Increment counter, pass to handler
   └─ If over limit: Return 429 Too Many Requests
4. Handler processes request (if allowed)
5. Response includes rate limit headers:
   - X-RateLimit-Limit: 5
   - X-RateLimit-Remaining: 3
   - Retry-After: 60 (if limited)
```

### New vs Modified Components

| Component | Type | Changes |
|-----------|------|---------|
| `internal/middleware/ratelimit.go` | NEW | Rate limiter factory functions |
| `cmd/main.go` | MODIFY | Add middleware to routes |
| `internal/config/config.go` | MODIFY | Add rate limit configuration |

---

## 2. GDPR Account Deletion Flow

### System-Wide Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Account Deletion Flow                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  User Request                                                   │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────┐                                           │
│  │ Settings Page   │                                           │
│  │ "Delete Account"│                                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Confirmation    │                                           │
│  │ Modal + Email   │                                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ API: DELETE     │                                           │
│  │ /api/account    │                                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────────────────────────────┐                   │
│  │      AccountDeletionService             │                   │
│  │  ┌───────────────────────────────────┐  │                   │
│  │  │ 1. Hard delete sessions           │  │                   │
│  │  │ 2. Hard delete API keys           │  │                   │
│  │  │ 3. Delete user VMs + namespaces   │  │                   │
│  │  │ 4. Anonymize user record          │  │                   │
│  │  │ 5. Log audit event                │  │                   │
│  │  │ 6. Send confirmation email        │  │                   │
│  │  └───────────────────────────────────┘  │                   │
│  └─────────────────────────────────────────┘                   │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Response: 204   │                                           │
│  │ No Content      │                                           │
│  └─────────────────┘                                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Database Schema Changes

```sql
-- migrations/006_account_deletion.sql

-- Add soft delete columns to users table
ALTER TABLE users 
    ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN anonymized BOOLEAN DEFAULT FALSE,
    ADD COLUMN deletion_requested_at TIMESTAMP WITH TIME ZONE;

-- Add index for cleanup job
CREATE INDEX idx_users_deleted_at ON users(deleted_at) 
    WHERE deleted_at IS NOT NULL AND anonymized = FALSE;

-- Add cascade delete for related data
ALTER TABLE sessions 
    ADD CONSTRAINT fk_sessions_user 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE api_keys 
    ADD CONSTRAINT fk_api_keys_user 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
```

### Component Integration

```
apps/backend/
├── cmd/
│   └── main.go                  # MODIFY: Add deletion route
├── internal/
│   ├── entity/
│   │   └── user.go              # MODIFY: Add deleted_at, anonymized
│   ├── repository/
│   │   └── user_repository.go   # MODIFY: Add deletion methods
│   ├── usecase/
│   │   └── account_deletion.go  # NEW: Deletion service
│   └── handler/
│       └── account/
│           └── deletion.go      # NEW: HTTP handler
└── migrations/
    └── 006_account_deletion.sql # NEW: Schema migration

apps/frontend/
├── src/
│   └── routes/
│       └── settings/
│           └── account.tsx      # MODIFY: Add delete button
```

### Service Implementation Pattern

```go
// internal/usecase/account_deletion.go

type AccountDeletionService struct {
    db            *sql.DB
    auditLogger   *AuditLogger
    emailService  *EmailService
    vmService     *VMService
}

func (s *AccountDeletionService) DeleteAccount(ctx context.Context, userID string) error {
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
    if err != nil {
        return fmt.Errorf("failed to start transaction: %w", err)
    }
    defer tx.Rollback()

    // 1. Hard delete sessions (immediate invalidation)
    _, err = tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
    if err != nil {
        return fmt.Errorf("failed to delete sessions: %w", err)
    }

    // 2. Hard delete API keys
    _, err = tx.ExecContext(ctx, `DELETE FROM api_keys WHERE user_id = $1`, userID)
    if err != nil {
        return fmt.Errorf("failed to delete API keys: %w", err)
    }

    // 3. Delete all user VMs (triggers k8s cleanup)
    err = s.vmService.DeleteAllUserVMs(ctx, tx, userID)
    if err != nil {
        return fmt.Errorf("failed to delete VMs: %w", err)
    }

    // 4. Anonymize user record (preserve referential integrity)
    _, err = tx.ExecContext(ctx, `
        UPDATE users 
        SET email = CONCAT('deleted-', id, '@deleted.local'),
            github_id = NULL,
            display_name = 'Deleted User',
            avatar_url = NULL,
            anonymized = TRUE,
            deleted_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `, userID)
    if err != nil {
        return fmt.Errorf("failed to anonymize user: %w", err)
    }

    // 5. Log audit event
    s.auditLogger.Log(ctx, AuditEvent{
        Type:      "account_deleted",
        UserID:    userID,
        Timestamp: time.Now(),
        Metadata:  map[string]interface{}{"method": "user_request"},
    })

    // 6. Send confirmation email (best effort, don't fail transaction)
    go func() {
        user, _ := s.getUserByID(ctx, userID) // Will get anonymized data
        s.emailService.SendDeletionConfirmation(user.Email)
    }()

    return tx.Commit()
}
```

### New vs Modified Components

| Component | Type | Changes |
|-----------|------|---------|
| `internal/usecase/account_deletion.go` | NEW | Deletion service logic |
| `internal/handler/account/deletion.go` | NEW | HTTP handler |
| `internal/entity/user.go` | MODIFY | Add deleted_at, anonymized |
| `internal/repository/user_repository.go` | MODIFY | Add deletion queries |
| `migrations/006_account_deletion.sql` | NEW | Schema changes |
| `apps/frontend/src/routes/settings/account.tsx` | MODIFY | Add delete UI |

---

## 3. OpenAPI/Swagger Integration

### Documentation Generation Flow

```
┌─────────────────────────────────────────────────────────────┐
│              OpenAPI Generation Pipeline                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Go Source Files (with swaggo comments)                    │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────┐                                       │
│  │ swag init       │                                       │
│  │ (build step)    │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────────────────────────────┐               │
│  │ docs/                                   │               │
│  │ ├── docs.go      (Go package)           │               │
│  │ ├── swagger.json (OpenAPI spec)         │               │
│  │ └── swagger.yaml (YAML spec)            │               │
│  └────────┬────────────────────────────────┘               │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ Swagger UI      │                                       │
│  │ /api/docs       │                                       │
│  └─────────────────┘                                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Component Integration

```
apps/backend/
├── cmd/
│   └── main.go                  # MODIFY: Add Swagger route, @title annotation
├── internal/
│   └── handler/
│       ├── auth/
│       │   └── oauth.go         # MODIFY: Add @Router, @Success annotations
│       ├── vm/
│       │   └── vm.go            # MODIFY: Add annotations
│       └── ...                  # All handlers need annotations
└── docs/
    ├── docs.go                  # GENERATED: swag init
    ├── swagger.json             # GENERATED: OpenAPI spec
    └── swagger.yaml             # GENERATED: YAML format
```

### Annotation Pattern

```go
// cmd/main.go

// @title           Podland API
// @version         1.1.0
// @description     Multi-tenant PaaS for students
// @termsOfService  https://podland.app/terms

// @contact.name   Podland Support
// @contact.email  support@podland.app

// @license.name   MIT
// @license.url    https://opensource.org/licenses/MIT

// @host      api.podland.app
// @BasePath  /api

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT token: "Bearer {token}"

// @externalDocs.description  User Guide
// @externalDocs.url          https://docs.podland.app
```

```go
// internal/handler/auth/oauth.go

// LoginHandler initiates GitHub OAuth flow
// @Summary      Initiate GitHub OAuth login
// @Description  Redirects user to GitHub OAuth authorization page
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      302  {string}  string  "Redirect to GitHub"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /api/auth/login [get]
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
    // ... implementation
}

// CallbackHandler handles GitHub OAuth callback
// @Summary      Handle OAuth callback
// @Description  Processes OAuth code, creates session, returns JWT
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        code  query  string  true  "OAuth authorization code"
// @Param        state  query  string  true  "OAuth state token"
// @Success      200  {object}  LoginResponse  "JWT tokens"
// @Failure      400  {object}  ErrorResponse  "Invalid code or state"
// @Router       /api/auth/callback [get]
func (h *Handler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
    // ... implementation
}
```

### Build Integration

```yaml
# .github/workflows/deploy.yml (modified)

- name: Generate OpenAPI docs
  run: |
    go install github.com/swaggo/swag/cmd/swag@v1.16.2
    cd apps/backend
    swag init -g cmd/main.go -o ./docs

- name: Commit generated docs
  run: |
    git add apps/backend/docs/
    git commit -m "docs: regenerate OpenAPI spec" || true
```

### New vs Modified Components

| Component | Type | Changes |
|-----------|------|---------|
| `docs/` directory | NEW | Generated OpenAPI files |
| `cmd/main.go` | MODIFY | Add @title annotations, Swagger route |
| All handler files | MODIFY | Add swaggo annotations |
| `.github/workflows/deploy.yml` | MODIFY | Add swag init step |

---

## 4. WCAG AA Integration

### Frontend Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              Accessibility Integration                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Development Mode                                           │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────┐                                       │
│  │ @axe-core/react │                                       │
│  │ (real-time      │                                       │
│  │  warnings)      │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────────────────────────────┐               │
│  │ React Components                        │               │
│  │ ├── Add ARIA labels                     │               │
│  │ ├── Fix focus management                │               │
│  │ ├── Ensure semantic HTML                │               │
│  │ └── Add skip links                      │               │
│  └────────┬────────────────────────────────┘               │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ CSS Updates     │                                       │
│  │ ├── Focus rings │                                       │
│  │ ├── Contrast    │                                       │
│  │ └── Skip links  │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ Lighthouse CI   │                                       │
│  │ (validation)    │                                       │
│  └─────────────────┘                                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Component Integration

```
apps/frontend/
├── src/
│   ├── main.tsx               # MODIFY: Add @axe-core/react in dev
│   ├── styles.css             # MODIFY: Focus indicators, contrast
│   ├── lib/
│   │   └── a11y.ts            # NEW: Utility functions
│   └── components/
│       ├── layout/
│       │   ├── Header.tsx     # MODIFY: Skip link, landmarks
│       │   └── Sidebar.tsx    # MODIFY: ARIA navigation
│       ├── vm/
│       │   ├── VMList.tsx     # MODIFY: ARIA labels, focus
│       │   └── VMCard.tsx     # MODIFY: Semantic HTML
│       └── common/
│           ├── Button.tsx     # MODIFY: Focus styles
│           └── Input.tsx      # MODIFY: Label association
└── .github/
    └── workflows/
        └── lighthouse.yml     # NEW: CI validation
```

### CSS Integration Pattern

```css
/* src/styles.css */

/* === Focus Indicators (WCAG 2.4.7, 2.4.11) === */
*:focus {
  outline: 3px solid #3b82f6;  /* Blue-500 */
  outline-offset: 2px;
}

/* Focus within for composite elements */
.focus-within:focus-within {
  outline: 3px solid #3b82f6;
  outline-offset: 2px;
}

/* Remove default outline only when custom focus exists */
button:focus:not(:focus-visible) {
  outline: none;
}

button:focus-visible {
  outline: 3px solid #3b82f6;
  outline-offset: 2px;
}

/* === Focus Not Obscured (WCAG 2.4.11) === */
html {
  scroll-padding-top: 80px;  /* Account for fixed header */
}

*:focus {
  scroll-margin-top: 90px;
}

/* === Color Contrast (WCAG 1.4.3) === */
.text-primary {
  color: #1f2937;  /* Gray-900, 16:1 on white */
}

.text-secondary {
  color: #4b5563;  /* Gray-600, 5.7:1 on white */
}

.text-muted {
  color: #6b7280;  /* Gray-500, 4.5:1 on white (minimum) */
}

/* Links - not color-only */
a {
  color: #2563eb;  /* Blue-600, 4.5:1 on white */
  text-decoration: underline;
  text-underline-offset: 2px;
}

/* === Skip Link (WCAG 2.4.1) === */
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  background: #3b82f6;
  color: white;
  padding: 8px 16px;
  z-index: 100;
  transition: top 0.2s;
}

.skip-link:focus {
  top: 0;
}

/* === Reduced Motion (WCAG 2.3.3) === */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

### React Component Pattern

```tsx
// src/components/vm/VMList.tsx

import { useEffect, useRef } from 'react';

export function VMList({ vms }: { vms: VM[] }) {
  const listRef = useRef<HTMLUListElement>(null);

  return (
    <section aria-labelledby="vm-list-heading">
      <h2 id="vm-list-heading">Your Virtual Machines</h2>
      
      {vms.length === 0 ? (
        <p role="status">No virtual machines found.</p>
      ) : (
        <ul 
          ref={listRef}
          role="list" 
          aria-label="Virtual machine list"
        >
          {vms.map((vm) => (
            <li key={vm.id}>
              <article aria-labelledby={`vm-title-${vm.id}`}>
                <h3 id={`vm-title-${vm.id}`}>
                  {vm.name}
                </h3>
                <p>
                  Status: <span aria-label={`Status: ${vm.status}`}>
                    {vm.status}
                  </span>
                </p>
                <a 
                  href={`/vms/${vm.id}`}
                  aria-label={`View details for ${vm.name}`}
                >
                  View Details
                </a>
              </article>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
```

### New vs Modified Components

| Component | Type | Changes |
|-----------|------|---------|
| `src/lib/a11y.ts` | NEW | Utility functions |
| `src/main.tsx` | MODIFY | Add @axe-core/react |
| `src/styles.css` | MODIFY | Focus, contrast, skip link |
| All components | MODIFY | ARIA labels, semantic HTML |
| `.github/workflows/lighthouse.yml` | NEW | CI validation |

---

## 5. Lighthouse CI Integration

### CI/CD Pipeline Integration

```
┌─────────────────────────────────────────────────────────────┐
│              Lighthouse CI Pipeline                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  GitHub Push/PR                                             │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────┐                                       │
│  │ Build Frontend  │                                       │
│  │ npm run build   │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ Start Server    │                                       │
│  │ npm run preview │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ LHCI Autorun    │                                       │
│  │ - Collect       │                                       │
│  │ - Assert        │                                       │
│  │ - Upload        │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ GitHub Status   │                                       │
│  │ Check           │                                       │
│  └─────────────────┘                                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Configuration File

```javascript
// lighthouserc.js (root of repo)

module.exports = {
  ci: {
    collect: {
      startServerCommand: 'npm run preview --workspace=@podland/frontend',
      startServerReadyPattern: 'ready',
      startServerReadyTimeout: 30000,
      url: [
        'http://localhost:4173/',
        'http://localhost:4173/dashboard',
        'http://localhost:4173/dashboard/vms',
        'http://localhost:4173/admin',
      ],
      numberOfRuns: 3,
      settings: {
        chromeFlags: '--no-sandbox',
        onlyCategories: 'performance,accessibility,best-practices,seo',
      },
    },
    assert: {
      preset: 'lighthouse:recommended',
      assertions: {
        // Relaxed thresholds for initial setup
        'categories:performance': ['warn', { minScore: 0.8 }],
        'categories:accessibility': ['error', { minScore: 0.9 }],
        'categories:best-practices': ['warn', { minScore: 0.8 }],
        'categories:seo': ['warn', { minScore: 0.8 }],
        // Core Web Vitals
        'first-contentful-paint': ['warn', { maxNumericValue: 1800 }],
        'largest-contentful-paint': ['warn', { maxNumericValue: 2500 }],
        'cumulative-layout-shift': ['warn', { maxNumericValue: 0.1 }],
        'total-blocking-time': ['warn', { maxNumericValue: 300 }],
        // Disable specific audits that may not apply
        'uses-rel-preload': 'off',
        'uses-rel-preconnect': 'off',
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
};
```

### GitHub Actions Workflow

```yaml
# .github/workflows/lighthouse.yml

name: Lighthouse CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lighthouse:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Build frontend
        run: npm run build:frontend

      - name: Run Lighthouse CI
        run: |
          npm install -g @lhci/cli@0.15.x
          lhci autorun
        env:
          LHCI_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Upload Lighthouse reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: lighthouse-reports
          path: .lighthouseci/
```

### New vs Modified Components

| Component | Type | Changes |
|-----------|------|---------|
| `lighthouserc.js` | NEW | LHCI configuration |
| `.github/workflows/lighthouse.yml` | NEW | CI workflow |
| `apps/frontend/package.json` | MODIFY | Add preview script |
| `apps/frontend/vite.config.ts` | MODIFY | Ensure preview works |

---

## 6. Load Testing Architecture

### Test Execution Flow

```
┌─────────────────────────────────────────────────────────────┐
│              Load Testing Architecture                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Test Environment (Separate from Production)               │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────┐                                       │
│  │ k6 Runner       │                                       │
│  │ (GitHub Actions │                                       │
│  │  or Local)      │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────────────────────────────┐               │
│  │ Test Scenarios                          │               │
│  │ ├── Smoke (10 VUs, 1m)                  │               │
│  │ ├── Load (50 VUs, 9m)                   │               │
│  │ └── Stress (200 VUs, 15m)               │               │
│  └────────┬────────────────────────────────┘               │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ Target: Podland │                                       │
│  │ Test Instance   │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ Results         │                                       │
│  │ - Metrics       │                                       │
│  │ - Thresholds    │                                       │
│  │ - Errors        │                                       │
│  └─────────────────┘                                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Enhanced k6 Script

```javascript
// tests/load/critical-paths.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const loginTime = new Trend('login_duration');
const vmCreateTime = new Trend('vm_create_duration');

export const options = {
  scenarios: {
    smoke: {
      executor: 'constant-vus',
      vus: 10,
      duration: '1m',
      tags: { test_type: 'smoke' },
    },
    load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },
        { duration: '5m', target: 50 },
        { duration: '2m', target: 0 },
      ],
      tags: { test_type: 'load' },
    },
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '5m', target: 100 },
        { duration: '5m', target: 200 },
        { duration: '5m', target: 0 },
      ],
      tags: { test_type: 'stress' },
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
    errors: ['rate<0.01'],
    login_duration: ['p(95)<300'],
    vm_create_duration: ['p(95)<2000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const startTime = Date.now();
  
  // 1. Auth flow
  const loginRes = http.post(`${BASE_URL}/api/auth/login`);
  loginTime.add(Date.now() - startTime);
  
  // ... rest of test
}
```

### New vs Modified Components

| Component | Type | Changes |
|-----------|------|---------|
| `tests/load/critical-paths.js` | MODIFY | Enhanced scenarios |
| `.github/workflows/load-test.yml` | NEW | CI workflow |
| `tests/load/README.md` | MODIFY | Documentation |

---

## Suggested Build Order

```
Week 1 (T+7): Security
├── Day 1-2: Add ulule/limiter dependency
├── Day 3-4: Create ratelimit.go middleware
├── Day 5: Apply to auth endpoints
└── Day 6-7: Test rate limiting

Week 2 (T+14): Documentation
├── Day 1-2: Install swaggo, add annotations to auth handlers
├── Day 3-4: Annotate VM handlers
├── Day 5: Annotate remaining handlers
├── Day 6: Add Swagger UI route
└── Day 7: Generate docs, test

Week 3 (T+21): Accessibility + Performance
├── Day 1-2: CSS updates (focus, contrast)
├── Day 3-4: Component ARIA updates
├── Day 5: Add @axe-core/react
├── Day 6-7: Lighthouse CI setup

Week 4 (T+30): Compliance + Testing
├── Day 1-2: Database migration for GDPR
├── Day 3-4: Account deletion service
├── Day 5: Frontend deletion UI
├── Day 6-7: Load test enhancement
```

---

## Summary: New vs Modified Components

| Feature | New Files | Modified Files |
|---------|-----------|----------------|
| Rate Limiting | 1 | 2 |
| GDPR Deletion | 3 | 4 |
| OpenAPI | 3 (generated) | ~10 (handlers) |
| WCAG AA | 1 | ~15 (components + CSS) |
| Lighthouse CI | 2 | 2 |
| Load Testing | 1 | 2 |

**Total:** 11 new files, ~35 modified files

---

*Research completed: 2026-03-30*
*Next: Review PITFALLS.md for common mistakes*

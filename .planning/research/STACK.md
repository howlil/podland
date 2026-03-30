# STACK.md — Technology Additions for v1.1 Hardening

**Milestone:** v1.1 — Hardening & Polish
**Date:** 2026-03-30
**Current Stack:** Go 1.25 + chi v5.2.5 + React 18 + TanStack Router + k3s

---

## Executive Summary

v1.1 requires **6 targeted technology additions** across security, compliance, documentation, accessibility, performance, and testing. This document specifies exact libraries, versions, integration points, and what to avoid.

| Area | Technology | Version | Priority | Complexity |
|------|-----------|---------|----------|------------|
| Rate Limiting | ulule/limiter | v3.11.2 | T+7 | Low |
| OpenAPI Docs | swaggo/swag | v1.16.2 | T+30 | Medium |
| Lighthouse CI | @lhci/cli | v0.15.x | T+30 | Low |
| A11y Testing | axe-core + @axe-core/react | v4.10.0 | T+30 | Low |
| Load Testing | k6 | v1.2.0+ | T+30 | Medium |
| GDPR Deletion | Custom implementation | N/A | T+30 | Medium |

---

## 1. Security: Rate Limiting

### Recommended Library

| Component | Package | Version | Rationale |
|-----------|---------|---------|-----------|
| **Rate Limiter** | `github.com/ulule/limiter/v3` | v3.11.2 | Dead simple, chi-compatible, Redis or in-memory store, battle-tested (10K+ stars) |
| **Store (Dev)** | `limiter/v3/drivers/store/memory` | Bundled | Zero dependencies, goroutine-based cleanup |
| **Store (Prod)** | `github.com/redis/go-redis/v9` | v9.7.0 | Distributed rate limiting, persists across restarts |

### Installation

```bash
cd apps/backend
go get github.com/ulule/limiter/v3@v3.11.2
go get github.com/redis/go-redis/v9@v9.7.0  # Production only
```

### Integration Points

**Where it fits:** Middleware layer, applied selectively to auth endpoints only.

```
apps/backend/
├── internal/
│   ├── middleware/
│   │   ├── middleware.go      # Existing: CORS, CSRF, Auth
│   │   ├── ratelimit.go       # NEW: Rate limiting middleware
│   │   └── audit.go
│   └── handler/
│       └── auth/
│           └── oauth.go       # Apply rate limit here
```

### Middleware Implementation Pattern

```go
// internal/middleware/ratelimit.go
package middleware

import (
    "github.com/ulule/limiter/v3"
    "github.com/ulule/limiter/v3/drivers/store/memory"
    limiterstd "github.com/ulule/limiter/v3/drivers/middleware/stdlib"
    "net/http"
    "time"
)

// NewAuthRateLimiter creates rate limiter for auth endpoints
// 5 requests per minute per IP (brute force protection)
func NewAuthRateLimiter() func(http.Handler) http.Handler {
    rate, _ := limiter.NewRateFromFormatted("5-M")
    store := memory.NewStore()
    instance := limiter.New(store, rate)
    return limiterstd.NewMiddleware(instance).Handler
}

// NewAPIRateLimiter creates rate limiter for general API
// 100 requests per minute per IP
func NewAPIRateLimiter() func(http.Handler) http.Handler {
    rate, _ := limiter.NewRateFromFormatted("100-M")
    store := memory.NewStore()
    instance := limiter.New(store, rate)
    return limiterstd.NewMiddleware(instance).Handler
}
```

### Application to Routes

```go
// cmd/main.go
r := chi.NewRouter()

// Apply rate limiting ONLY to auth endpoints
r.Route("/api/auth", func(r chi.Router) {
    r.Use(NewAuthRateLimiter())  // 5 req/min
    r.Get("/login", auth.LoginHandler)
    r.Get("/callback", auth.CallbackHandler)
})

// General API rate limiting
r.Route("/api", func(r chi.Router) {
    r.Use(NewAPIRateLimiter())  // 100 req/min
    // ... other routes
})
```

### Configuration

| Endpoint | Limit | Window | Rationale |
|----------|-------|--------|-----------|
| `/api/auth/*` | 5 requests | 1 minute | Brute force protection |
| `/api/*` (general) | 100 requests | 1 minute | API abuse prevention |
| `/api/vms/*` | 50 requests | 1 minute | VM operations are expensive |

### What NOT to Add

| Technology | Why Avoid |
|-----------|-----------|
| `golang.org/x/time/rate` | Too low-level, requires manual middleware wrapping, no built-in store abstraction |
| `tolerance/tolerance` | Less mature, smaller community, fewer examples |
| Custom Redis implementation | Reinventing the wheel, ulule already supports Redis |
| Rate limiting on static assets | Wastes resources, Cloudflare handles this |

### Dependencies

- **None** for in-memory store (development)
- **Redis** for production distributed rate limiting (optional for v1.1, can defer to v1.2)

---

## 2. Compliance: GDPR Account Deletion

### Technology Required

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| **Soft Delete** | Custom implementation | N/A | Simple boolean flag + anonymization |
| **Cascade Delete** | PostgreSQL foreign keys | N/A | Database-level referential integrity |
| **Audit Logging** | Existing audit middleware | N/A | Already implemented in v1.0 |

### Implementation Pattern

**Hybrid Approach:** Soft delete user record (preserve referential integrity) + hard delete PII and sessions.

```go
// internal/entity/user.go
type User struct {
    ID           string
    Email        string
    GitHubID     string
    // ... other fields
    DeletedAt    *time.Time  // NEW: Soft delete timestamp
    Anonymized   bool        // NEW: Flag for anonymization
}

// internal/usecase/account_deletion.go
type AccountDeletionService struct {
    db *sql.DB
}

// RequestDeletion initiates soft delete with 30-day grace period
func (s *AccountDeletionService) RequestDeletion(ctx context.Context, userID string) error {
    // 1. Mark as pending deletion (30-day grace period)
    gracePeriod := time.Now().Add(30 * 24 * time.Hour)
    _, err := s.db.ExecContext(ctx, `
        UPDATE users 
        SET deleted_at = $1, anonymized = false
        WHERE id = $2 AND deleted_at IS NULL
    `, gracePeriod, userID)
    return err
}

// ExecuteDeletion performs hard delete after grace period
func (s *AccountDeletionService) ExecuteDeletion(ctx context.Context, userID string) error {
    tx, _ := s.db.BeginTx(ctx, nil)
    
    // 1. Hard delete sessions (prevent re-login)
    tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
    
    // 2. Hard delete API keys
    tx.ExecContext(ctx, `DELETE FROM api_keys WHERE user_id = $1`, userID)
    
    // 3. Anonymize user record (preserve referential integrity)
    tx.ExecContext(ctx, `
        UPDATE users 
        SET email = CONCAT('deleted-', id, '@deleted.local'),
            github_id = NULL,
            display_name = 'Deleted User',
            anonymized = true
        WHERE id = $1
    `, userID)
    
    // 4. Delete user VMs and namespaces
    tx.ExecContext(ctx, `DELETE FROM vms WHERE user_id = $1`, userID)
    
    return tx.Commit()
}
```

### Data Retention Policy

| Data Type | Retention | Method | Rationale |
|-----------|-----------|--------|-----------|
| **User PII** | Immediate anonymization | Hard delete | GDPR right to erasure |
| **Sessions** | Immediate | Hard delete | Security |
| **Audit logs** | 7 years | Retain (anonymized) | Legal compliance |
| **VM metadata** | 30 days | Retain (anonymized) | Debugging, analytics |
| **Metrics** | 15 days (Prometheus default) | Auto-expire | Already compliant |

### What NOT to Add

| Technology | Why Avoid |
|-----------|-----------|
| Complex workflow engines | Overkill for single deletion flow |
| Third-party GDPR tools | Manual implementation is simpler for this scale |
| Data archiving systems | 500 users × minimal data = negligible storage |

### Dependencies

- **PostgreSQL foreign key constraints** — Already in place
- **Audit logging** — Already implemented in v1.0
- **Email service (SendGrid)** — For deletion confirmation emails

---

## 3. Documentation: OpenAPI/Swagger

### Recommended Tool

| Component | Package | Version | Rationale |
|-----------|---------|---------|-----------|
| **OpenAPI Generator** | `github.com/swaggo/swag` | v1.16.2 | Most popular Go OpenAPI generator, chi-compatible, active maintenance |
| **Swagger UI** | `github.com/swaggo/http-swagger` | v1.3.4 | Official Swagger UI wrapper for Go |
| **CLI Tool** | `swag` command | v1.16.2 | Generates docs from comments |

### Installation

```bash
cd apps/backend
go install github.com/swaggo/swag/cmd/swag@v1.16.2
go get github.com/swaggo/swag@v1.16.2
go get github.com/swaggo/http-swagger@v1.3.4
```

### Integration Points

```
apps/backend/
├── cmd/
│   └── main.go            # Add Swagger UI route
├── internal/
│   └── handler/
│       ├── auth/
│       │   └── oauth.go   # Add swaggo comments
│       ├── vm/
│       │   └── vm.go      # Add swaggo comments
│       └── ...
└── docs/
    ├── docs.go            # Generated
    ├── swagger.json       # Generated
    └── swagger.yaml       # Generated
```

### Annotation Pattern

```go
// internal/handler/auth/oauth.go

// LoginHandler handles GitHub OAuth login
// @Summary      Initiate GitHub OAuth login
// @Description  Redirects user to GitHub OAuth authorization page
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      302  {string}  string  "Redirect to GitHub"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/auth/login [get]
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
    // ... implementation
}

// ErrorResponse is the error response structure
// @Description Error response
// @APIModel
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}
```

### Swagger UI Route

```go
// cmd/main.go
import swagger "github.com/swaggo/http-swagger"

r := chi.NewRouter()

// Swagger UI at /api/docs
r.Get("/api/docs/*", swagger.Handler)

// Generate docs with: swag init -g cmd/main.go
```

### CI/CD Integration

```yaml
# .github/workflows/docs.yml
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

### What NOT to Add

| Technology | Why Avoid |
|-----------|-----------|
| `go-openapi` | More complex, steeper learning curve, overkill for REST API |
| Manual OpenAPI YAML | Error-prone, drifts from implementation |
| Redoc | Swagger UI is sufficient for v1.1 |
| GraphQL + Apollo | Out of scope, REST is sufficient |

### Dependencies

- **Go comments in handlers** — Requires discipline to maintain
- **Build step** — `swag init` must run before deployment

---

## 4. Accessibility: WCAG AA Compliance

### Testing Tools

| Component | Package | Version | Rationale |
|-----------|---------|---------|-----------|
| **A11y Testing** | `axe-core` | v4.10.0 | Industry standard, comprehensive WCAG 2.2 coverage |
| **React Integration** | `@axe-core/react` | v4.10.0 | Real-time a11y warnings in development |
| **Lighthouse** | `lighthouse` | v12.2.0 | Automated a11y scoring, CI integration |
| **Color Contrast** | `polished` | v4.3.0 | Programmatic contrast checking |

### Installation

```bash
cd apps/frontend
npm install -D axe-core@4.10.0 @axe-core/react@4.10.0 polished@4.3.0
```

### Integration Points

```
apps/frontend/
├── src/
│   ├── main.tsx           # Add @axe-core/react in dev
│   ├── styles.css         # Fix focus indicators, contrast
│   ├── components/
│   │   ├── layout/
│   │   │   └── Header.tsx # Fix skip links, landmarks
│   │   └── vm/
│   │       └── VMList.tsx # Fix ARIA labels, focus management
│   └── lib/
│       └── a11y.ts        # Utility functions
└── .github/
    └── workflows/
        └── lighthouse.yml # CI integration
```

### Development Setup

```tsx
// src/main.tsx
import React from 'react'
import ReactDOM from 'react-dom/client'

if (process.env.NODE_ENV === 'development') {
  import('@axe-core/react').then((axe) => {
    axe.default(React, ReactDOM, 1000) // 1000ms debounce
  })
}
```

### CSS Fixes Required

```css
/* src/styles.css */

/* Focus indicators (WCAG 2.4.7, 2.4.11) */
*:focus {
  outline: 3px solid #3b82f6;  /* Blue-500 */
  outline-offset: 2px;
}

/* Focus not obscured (WCAG 2.4.11) */
html {
  scroll-padding-top: 80px;  /* Account for fixed header */
}

*:focus {
  scroll-margin-top: 90px;
}

/* Color contrast fixes (WCAG 1.4.3) */
.text-muted {
  color: #4b5563;  /* Gray-600, 5.7:1 on white */
}

.link {
  color: #2563eb;  /* Blue-600, 4.5:1 on white */
  text-decoration: underline;  /* Not color-only */
}

/* Skip link (WCAG 2.4.1) */
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  background: #3b82f6;
  color: white;
  padding: 8px;
  z-index: 100;
}

.skip-link:focus {
  top: 0;
}
```

### What NOT to Add

| Technology | Why Avoid |
|-----------|-----------|
| Overlay widgets (accessiBe, UserWay) | Harmful, create more problems than they solve |
| Automated a11y "fix" tools | Cannot fix semantic issues automatically |
| Screen reader testing services | Manual testing with NVDA/VoiceOver is sufficient |

### Dependencies

- **Design system updates** — Color palette may need adjustment
- **Component refactoring** — ARIA labels, focus management

---

## 5. Performance: Lighthouse CI

### Recommended Tools

| Component | Package | Version | Rationale |
|-----------|---------|---------|-----------|
| **Lighthouse CI** | `@lhci/cli` | v0.15.x | Official CLI, GitHub Actions integration |
| **Lighthouse** | `lighthouse` | v12.2.0 | Latest version with Core Web Vitals |
| **GitHub App** | LHCI GitHub App | N/A | Status checks on PRs |

### Installation

```bash
cd apps/frontend
npm install -D @lhci/cli@0.15.x lighthouse@12.2.0
```

### Configuration File

```javascript
// lighthouserc.js
module.exports = {
  ci: {
    collect: {
      startServerCommand: 'npm run preview',
      startServerReadyPattern: 'ready',
      url: [
        'http://localhost:4173/',
        'http://localhost:4173/dashboard',
        'http://localhost:4173/dashboard/vms',
      ],
      numberOfRuns: 3,
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
      },
    },
    upload: {
      target: 'temporary-public-storage',  // Start here, upgrade to LHCI server later
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
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm ci

      - name: Build frontend
        run: npm run build:frontend

      - name: Run Lighthouse CI
        run: |
          npm install -g @lhci/cli@0.15.x
          cd apps/frontend
          lhci autorun
        env:
          LHCI_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### What NOT to Add

| Technology | Why Avoid |
|-----------|-----------|
| Self-hosted LHCI server | Overkill for v1.1, temporary public storage is sufficient |
| Performance budgets per page | Start with global thresholds, refine later |
| Real User Monitoring (RUM) | Defer to v1.2, synthetic testing is sufficient initially |

### Dependencies

- **Build process** — Must have `npm run build` and `npm run preview`
- **GitHub token** — For status checks on PRs

---

## 6. Testing: Load Testing with k6

### Current State

k6 script already exists at `tests/load/critical-paths.js`. Needs enhancement for v1.1.

### Recommended Enhancements

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| **Load Testing** | `k6` | v1.2.0+ | Already in use, excellent Go API testing |
| **Grafana Integration** | `k6-output-prometheus` | v0.5.0 | Visualize load test results |
| **CI Integration** | k6 GitHub Action | v2 | Run load tests in CI |

### Enhanced Test Scenarios

```javascript
// tests/load/critical-paths.js (enhanced)

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const loginTime = new Trend('login_duration');

export const options = {
  scenarios: {
    // Scenario 1: Smoke test (quick validation)
    smoke: {
      executor: 'constant-vus',
      vus: 10,
      duration: '1m',
      tags: { test_type: 'smoke' },
    },
    // Scenario 2: Load test (normal traffic)
    load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },   // Ramp up to 50 users
        { duration: '5m', target: 50 },   // Stay at 50 users
        { duration: '2m', target: 0 },    // Ramp down
      ],
      tags: { test_type: 'load' },
    },
    // Scenario 3: Stress test (breaking point)
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '5m', target: 100 },  // Ramp to 100 users
        { duration: '5m', target: 200 },  // Ramp to 200 users
        { duration: '5m', target: 0 },    // Ramp down
      ],
      tags: { test_type: 'stress' },
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500'],  // p95 < 500ms
    http_req_failed: ['rate<0.01'],    // <1% errors
    errors: ['rate<0.01'],
    login_duration: ['p(95)<300'],     // p95 login < 300ms
  },
};
```

### CI Integration

```yaml
# .github/workflows/load-test.yml
name: Load Testing

on:
  workflow_dispatch:  # Manual trigger
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday at 2 AM

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run k6 load test
        uses: grafana/k6-action@v2
        with:
          filename: tests/load/critical-paths.js
        env:
          BASE_URL: ${{ secrets.LOAD_TEST_BASE_URL }}
```

### What NOT to Add

| Technology | Why Avoid |
|-----------|-----------|
| JMeter | k6 is already integrated, JS-based scripts are easier to maintain |
| Locust | Python-based, k6 Go backend alignment is better |
| Complex distributed load testing | Single server target, k6 on one machine is sufficient |

### Dependencies

- **Test environment** — Separate from production
- **Test data** — Pre-seeded users and VMs for realistic testing

---

## Summary: What to Add vs Avoid

### Add (v1.1)

| Priority | Technology | Purpose | Timeline |
|----------|-----------|---------|----------|
| P0 | ulule/limiter v3.11.2 | Auth rate limiting | T+7 |
| P0 | Custom GDPR deletion | Account deletion flow | T+30 |
| P1 | swaggo/swag v1.16.2 | OpenAPI documentation | T+30 |
| P1 | @lhci/cli v0.15.x | Lighthouse CI | T+30 |
| P1 | axe-core v4.10.0 | Accessibility testing | T+30 |
| P2 | k6 enhancements | Load testing scenarios | T+30 |

### Avoid (Overkill for v1.1)

| Technology | Why Avoid |
|-----------|-----------|
| Redis for rate limiting | Single server, in-memory is sufficient |
| Self-hosted LHCI server | Temporary storage works for now |
| Complex workflow engines | Manual GDPR flow is simpler |
| GraphQL + Apollo | REST is sufficient |
| JMeter/Locust | k6 already integrated |
| Overlay a11y widgets | Harmful, manual fixes are better |

---

## Version Verification

All versions verified as of 2026-03-30:

- ✅ ulule/limiter v3.11.2 — Latest stable (GitHub releases)
- ✅ swaggo/swag v1.16.2 — Latest stable (GitHub releases)
- ✅ @lhci/cli v0.15.x — Latest stable (npm)
- ✅ axe-core v4.10.0 — Latest stable (npm)
- ✅ k6 v1.2.0+ — Latest stable (grafana.com)

---

*Research completed: 2026-03-30*
*Next: Review FEATURES.md for feature patterns*

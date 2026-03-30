# FEATURES.md — Feature Patterns for v1.1 Hardening

**Milestone:** v1.1 — Hardening & Polish
**Date:** 2026-03-30
**Context:** Building on v1.0 foundation (48 requirements shipped)

---

## Executive Summary

v1.1 adds **6 feature categories** focused on production hardening. Each category has table stakes (must-have), differentiators (nice-to-have), and anti-features (avoid). This document provides patterns, complexity notes, and dependencies.

| Category | Table Stakes | Differentiators | Anti-Features | Complexity |
|----------|-------------|-----------------|---------------|------------|
| Security (Rate Limiting) | Auth endpoint limits | Per-user limits | Global rate limiting | Low |
| Compliance (GDPR) | Account deletion | Grace period | Data export | Medium |
| Documentation (OpenAPI) | Swagger UI | Interactive examples | Manual YAML | Medium |
| Accessibility (WCAG AA) | Focus indicators, contrast | Screen reader testing | Overlay widgets | Medium |
| Performance (Lighthouse) | CI integration | Performance budgets | RUM (v2+) | Low |
| Testing (Load) | Critical path tests | Automated CI | Distributed testing | Medium |

---

## 1. Security: Rate Limiting

### Table Stakes (Must-Have)

| Feature | Description | Implementation | Priority |
|---------|-------------|----------------|----------|
| **Auth endpoint rate limiting** | 5 requests/minute per IP on `/api/auth/*` | ulule/limiter middleware | P0 |
| **IP-based identification** | Use client IP as rate limit key | `X-Forwarded-For` (trusted proxy) | P0 |
| **429 Too Many Requests** | Return proper HTTP status with `Retry-After` header | Standard middleware response | P0 |
| **Configurable limits** | Environment variable configuration | `RATE_LIMIT_AUTH=5-M` | P1 |

### Differentiators (Nice-to-Have)

| Feature | Description | Complexity | Timeline |
|---------|-------------|------------|----------|
| **Per-user rate limiting** | Higher limits for authenticated users | Medium | v1.2 |
| **Redis-backed store** | Distributed rate limiting across instances | Low | v1.2 |
| **Rate limit headers** | `X-RateLimit-Limit`, `X-RateLimit-Remaining` | Low | v1.2 |
| **Graduated response** | Add warning headers before hard limit | Medium | v1.2 |

### Anti-Features (Avoid)

| Anti-Feature | Why Avoid | Alternative |
|--------------|-----------|-------------|
| **Global rate limiting** | Punishes all users for one attacker's actions | Per-IP or per-user limits |
| **Rate limiting on static assets** | Wastes resources, Cloudflare handles this | Skip static files |
| **Complex rate limit rules** | Hard to debug, maintain | Simple per-endpoint limits |
| **Rate limiting in application logic** | Scattered, hard to audit | Centralized middleware |

### Complexity Notes

**Overall: Low**

- ulule/limiter is well-documented and chi-compatible
- In-memory store requires zero infrastructure
- ~50 lines of middleware code
- No database changes required

### Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| chi router v5.2.5 | ✅ Shipped (v1.0) | Middleware compatible |
| Cloudflare proxy | ✅ Configured | Provides trusted IP headers |
| Redis | ❌ Not required | Can use in-memory for v1.1 |

---

## 2. Compliance: GDPR Account Deletion

### Table Stakes (Must-Have)

| Feature | Description | Implementation | Priority |
|---------|-------------|----------------|----------|
| **Account deletion request** | User can request account deletion from settings | UI button + API endpoint | P0 |
| **PII anonymization** | Email, name, GitHub ID anonymized on deletion | Database UPDATE | P0 |
| **Session invalidation** | All sessions deleted immediately | Hard delete sessions | P0 |
| **VM cleanup** | All user VMs deleted with account | Cascade delete | P0 |
| **Audit logging** | Deletion action logged for compliance | Existing audit middleware | P0 |
| **Confirmation email** | Email sent confirming deletion | SendGrid integration | P1 |

### Differentiators (Nice-to-Have)

| Feature | Description | Complexity | Timeline |
|---------|-------------|------------|----------|
| **30-day grace period** | Account marked for deletion, user can cancel | Medium | v1.2 |
| **Data export** | User can download their data before deletion | Medium | v1.2 |
| **Deletion dashboard** | Admin view of pending deletions | Low | v1.2 |
| **Scheduled cleanup job** | Cron job executes pending deletions | Low | v1.2 |

### Anti-Features (Avoid)

| Anti-Feature | Why Avoid | Alternative |
|--------------|-----------|-------------|
| **Immediate hard delete of everything** | Breaks referential integrity, audit trail | Soft delete + anonymize |
| **Complex approval workflows** | Overkill for student PaaS | Self-service deletion |
| **Third-party GDPR tools** | Expensive, complex for 500 users | Manual implementation |
| **Data retention beyond legal requirements** | GDPR violation | 7 years for audit logs only |

### Complexity Notes

**Overall: Medium**

- Requires database schema changes (`deleted_at`, `anonymized` columns)
- Cascade delete logic must be carefully tested
- Email service integration required
- Admin UI for viewing deletion status

### Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| PostgreSQL | ✅ Shipped (v1.0) | Foreign key constraints |
| SendGrid | ✅ Shipped (v1.0) | Deletion confirmation emails |
| Audit logging | ✅ Shipped (v1.0) | Log deletion actions |
| VM management | ✅ Shipped (v1.0) | VM cleanup on deletion |

---

## 3. Documentation: OpenAPI/Swagger

### Table Stakes (Must-Have)

| Feature | Description | Implementation | Priority |
|---------|-------------|----------------|----------|
| **Swagger UI** | Interactive API documentation at `/api/docs` | swaggo/http-swagger | P0 |
| **Auto-generated spec** | OpenAPI spec generated from code comments | swaggo/swag CLI | P0 |
| **All endpoints documented** | Auth, VM, Domain, Quota, Admin endpoints | Handler annotations | P0 |
| **Example requests/responses** | Sample payloads in Swagger UI | swaggo annotations | P1 |
| **Authentication docs** | JWT token usage documented | Security scheme definition | P1 |

### Differentiators (Nice-to-Have)

| Feature | Description | Complexity | Timeline |
|---------|-------------|------------|----------|
| **Try-it-out functionality** | Execute API calls from Swagger UI | Low | v1.2 |
| **API versioning** | Multiple API versions in docs | Medium | v1.2 |
| **Downloadable spec** | JSON/YAML download link | Low | v1.2 |
| **Custom branding** | Podland logo, colors in Swagger UI | Low | v1.2 |

### Anti-Features (Avoid)

| Anti-Feature | Why Avoid | Alternative |
|--------------|-----------|-------------|
| **Manual OpenAPI YAML** | Drifts from implementation, error-prone | Auto-generate from code |
| **Separate API docs site** | Maintenance burden, URL fragmentation | Embedded Swagger UI |
| **Over-documented internal endpoints** | Clutters docs | Document public API only |
| **Redoc instead of Swagger UI** | Swagger UI is more familiar | Stick with Swagger UI |

### Complexity Notes

**Overall: Medium**

- Requires annotating all handler functions (~200 lines of comments)
- Build step integration (`swag init`)
- Learning curve for swaggo annotation syntax
- Ongoing maintenance discipline required

### Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| Go handlers | ✅ Shipped (v1.0) | Add swaggo comments |
| Build pipeline | ✅ Shipped (v1.0) | Add `swag init` step |
| chi router | ✅ Shipped (v1.0) | Compatible with swaggo |

---

## 4. Accessibility: WCAG AA Compliance

### Table Stakes (Must-Have)

| Feature | Description | Implementation | Priority |
|---------|-------------|----------------|----------|
| **Color contrast 4.5:1** | All text meets WCAG 1.4.3 minimum | Update Tailwind colors | P0 |
| **Focus indicators 3:1** | Visible focus rings on all interactive elements | CSS `:focus` styles | P0 |
| **Skip links** | Keyboard users can skip to main content | ARIA skip link | P0 |
| **Form labels** | All inputs have associated labels | HTML `label` elements | P0 |
| **Image alt text** | All informative images have alt text | `alt` attributes | P0 |
| **Heading hierarchy** | Proper h1 → h2 → h3 structure | Semantic HTML | P0 |
| **ARIA landmarks** | `main`, `nav`, `header`, `footer` regions | HTML5 landmarks | P1 |

### Differentiators (Nice-to-Have)

| Feature | Description | Complexity | Timeline |
|---------|-------------|------------|----------|
| **Screen reader testing** | Manual testing with NVDA/VoiceOver | Low | v1.2 |
| **Reduced motion support** | Respect `prefers-reduced-motion` | Low | v1.2 |
| **High contrast mode** | Alternative color scheme | Medium | v1.2 |
| **Keyboard navigation guide** | Documentation for keyboard users | Low | v1.2 |

### Anti-Features (Avoid)

| Anti-Feature | Why Avoid | Alternative |
|--------------|-----------|-------------|
| **Accessibility overlay widgets** | Create more problems than they solve | Manual fixes |
| **Automated "fix all" tools** | Cannot fix semantic issues | Manual refactoring |
| **Removing focus outlines** | Keyboard users lose navigation | Style them better |
| **Color-only state indicators** | Colorblind users can't perceive | Add icons/text |

### Complexity Notes

**Overall: Medium**

- CSS changes across multiple components
- Some components may need structural refactoring
- Testing requires manual effort (screen readers)
- Design system updates may be needed

### Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| Tailwind CSS v4 | ✅ Shipped (v1.0) | Update color palette |
| React components | ✅ Shipped (v1.0) | Add ARIA attributes |
| axe-core | ❌ New | Development testing |

---

## 5. Performance: Lighthouse CI

### Table Stakes (Must-Have)

| Feature | Description | Implementation | Priority |
|---------|-------------|----------------|----------|
| **Lighthouse CI integration** | Automated Lighthouse scores on every PR | @lhci/cli + GitHub Actions | P0 |
| **Core Web Vitals tracking** | LCP, FID, CLS monitored | Lighthouse audits | P0 |
| **Performance budgets** | Minimum scores enforced | LHCI assertions | P0 |
| **Multi-page testing** | Home, Dashboard, VM list tested | Multiple URLs in config | P1 |
| **Public report storage** | Reports accessible via URL | Temporary public storage | P1 |

### Differentiators (Nice-to-Have)

| Feature | Description | Complexity | Timeline |
|---------|-------------|------------|----------|
| **LHCI server** | Self-hosted report dashboard | Medium | v1.2 |
| **Performance trends** | Historical performance tracking | Medium | v1.2 |
| **PR comments** | Lighthouse scores in PR description | Low | v1.2 |
| **Mobile testing** | Emulated mobile Lighthouse runs | Low | v1.2 |

### Anti-Features (Avoid)

| Anti-Feature | Why Avoid | Alternative |
|--------------|-----------|-------------|
| **100% score requirement** | Unrealistic, blocks PRs for minor issues | 80-90% thresholds |
| **Real User Monitoring (RUM)** | Overkill for v1.1 | Synthetic testing first |
| **Per-page budgets initially** | Too complex to maintain | Global thresholds first |
| **Blocking all PRs on performance** | Start with warnings, escalate later | `warn` before `error` |

### Complexity Notes

**Overall: Low**

- Configuration file setup (~30 lines)
- GitHub Actions workflow (~40 lines)
- Build process integration
- Minimal code changes required

### Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| Build process | ✅ Shipped (v1.0) | `npm run build` |
| Preview server | ⚠️ Needs setup | `npm run preview` |
| GitHub Actions | ✅ Shipped (v1.0) | Add new workflow |

---

## 6. Testing: Load Testing

### Table Stakes (Must-Have)

| Feature | Description | Implementation | Priority |
|---------|-------------|----------------|----------|
| **Critical path tests** | Login, VM create/start/stop/delete | k6 scripts (existing) | P0 |
| **Load scenario** | 50 concurrent users, 5 minutes | k6 ramping VUs | P0 |
| **Stress scenario** | 200 concurrent users, breaking point | k6 stress test | P0 |
| **Performance thresholds** | p95 < 500ms, <1% errors | k6 thresholds | P0 |
| **Test documentation** | How to run, interpret results | README updates | P1 |

### Differentiators (Nice-to-Have)

| Feature | Description | Complexity | Timeline |
|---------|-------------|------------|----------|
| **Automated CI runs** | Weekly scheduled load tests | Low | v1.2 |
| **Grafana visualization** | Load test metrics in Grafana | Medium | v1.2 |
| **API-specific tests** | Individual endpoint benchmarks | Low | v1.2 |
| **Database load testing** | Query performance under load | Medium | v1.2 |

### Anti-Features (Avoid)

| Anti-Feature | Why Avoid | Alternative |
|--------------|-----------|-------------|
| **Testing against production** | Risk of data corruption, performance impact | Separate test environment |
| **Distributed load testing** | Overkill for single-server target | Single k6 instance |
| **JMeter migration** | k6 already integrated, JS is easier | Enhance existing k6 |
| **100% endpoint coverage** | Diminishing returns | Focus on critical paths |

### Complexity Notes

**Overall: Medium**

- Existing k6 script needs enhancement
- Test environment setup required
- Test data seeding needed
- Result interpretation requires training

### Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| k6 | ⚠️ Partial | Script exists, needs enhancement |
| Test environment | ❌ Needed | Separate from production |
| Test users | ❌ Needed | Pre-seeded accounts for testing |

---

## Feature Dependencies Graph

```
v1.1 Feature Dependencies:

Security (Rate Limiting)
├── Depends on: chi middleware (✅ v1.0)
└── Enables: Auth endpoint protection

Compliance (GDPR)
├── Depends on: PostgreSQL (✅ v1.0), SendGrid (✅ v1.0), Audit logging (✅ v1.0)
├── Depends on: VM management (✅ v1.0)
└── Enables: User data deletion

Documentation (OpenAPI)
├── Depends on: Go handlers (✅ v1.0)
└── Enables: API discoverability

Accessibility (WCAG AA)
├── Depends on: React components (✅ v1.0), Tailwind (✅ v1.0)
└── Enables: Inclusive user experience

Performance (Lighthouse)
├── Depends on: Build process (✅ v1.0)
└── Enables: Performance monitoring

Testing (Load)
├── Depends on: k6 (⚠️ partial), Test environment (❌ needed)
└── Enables: Performance validation
```

---

## Build Order Recommendation

```
T+7 (Week 1):
└── Security: Rate limiting on auth endpoints
    └── ulule/limiter integration
    └── Middleware testing

T+14 (Week 2):
└── Documentation: OpenAPI/Swagger
    └── swaggo annotations
    └── Swagger UI setup

T+21 (Week 3):
├── Accessibility: WCAG AA fixes
│   ├── Color contrast updates
│   ├── Focus indicators
│   └── ARIA labels
└── Performance: Lighthouse CI
    └── Configuration setup
    └── GitHub Actions workflow

T+30 (Week 4):
├── Compliance: GDPR account deletion
│   ├── Database schema changes
│   ├── Deletion flow implementation
│   └── Admin UI
└── Testing: Load testing execution
    ├── k6 script enhancement
    └── Test environment setup
```

---

## Summary by Category

| Category | Table Stakes Count | Differentiators Count | Anti-Features Count | Complexity |
|----------|-------------------|----------------------|---------------------|------------|
| Security | 4 | 4 | 4 | Low |
| Compliance | 6 | 4 | 4 | Medium |
| Documentation | 5 | 4 | 4 | Medium |
| Accessibility | 7 | 4 | 4 | Medium |
| Performance | 5 | 4 | 4 | Low |
| Testing | 5 | 4 | 4 | Medium |

**Total:** 32 table stakes, 24 differentiators, 24 anti-features identified

---

*Research completed: 2026-03-30*
*Next: Review ARCHITECTURE.md for integration patterns*

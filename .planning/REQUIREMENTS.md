# Requirements: Podland

**Defined:** 2026-03-30
**Core Value:** Students can deploy and run applications with zero DevOps knowledge — just create a "VM", get auto-configured domain and tunnel, and focus on code.

---

## v1.0 Requirements (Shipped)

All 48 requirements from v1.0 have been implemented and validated.

**Summary by Category:**
- ✅ Authentication (6/6) — AUTH-01 through AUTH-06
- ✅ VM Management (11/11) — VM-01 through VM-11
- ✅ Domain & Networking (6/6) — DOMAIN-01 through DOMAIN-06
- ✅ Resource Quotas (5/5) — QUOTA-01 through QUOTA-05
- ✅ Monitoring & Observability (6/6) — MON-01 through MON-06
- ✅ User Dashboard (4/4) — DASH-01 through DASH-04
- ✅ Admin Panel (5/5) — ADMIN-01 through ADMIN-05
- ✅ API (4/4) — API-01 through API-04

**Status:** 48/48 complete (100%)

---

## v1.1 Requirements (Current Milestone)

Requirements for v1.1 — Hardening & Polish. Each maps to roadmap phases.

### Security (Rate Limiting)

- [ ] **SEC-01**: System enforces rate limit of 5 requests/minute on `/api/auth/*` endpoints per IP
- [ ] **SEC-02**: System returns HTTP 429 with `Retry-After` header when rate limit exceeded
- [ ] **SEC-03**: System uses `X-Forwarded-For` header for IP identification (Cloudflare trusted proxy)
- [ ] **SEC-04**: Rate limit threshold configurable via `RATE_LIMIT_AUTH` environment variable

### Compliance (GDPR)

- [ ] **GDPR-01**: User can request account deletion from settings page
- [ ] **GDPR-02**: System anonymizes PII (email, name, GitHub ID) on account deletion
- [ ] **GDPR-03**: System invalidates all user sessions immediately on deletion
- [ ] **GDPR-04**: System deletes all user VMs with account (cascade delete)
- [ ] **GDPR-05**: System logs account deletion to audit log for compliance
- [ ] **GDPR-06**: System sends confirmation email after account deletion completes

### Documentation (OpenAPI + Guides)

- [ ] **DOC-01**: System provides OpenAPI/Swagger specification for all API endpoints
- [ ] **DOC-02**: Swagger UI accessible at `/api/docs` with interactive testing
- [ ] **DOC-03**: API documentation includes authentication, error codes, rate limits
- [ ] **DOC-04**: User guide covers sign in, create VM, access via domain workflows
- [ ] **DOC-05**: User guide accessible from dashboard help menu

### Accessibility (WCAG AA)

- [ ] **A11Y-01**: All interactive elements have visible focus indicators (3px outline minimum)
- [ ] **A11Y-02**: Color contrast ratio meets WCAG AA (4.5:1 for text, 3:1 for UI components)
- [ ] **A11Y-03**: All images have descriptive alt text
- [ ] **A11Y-04**: Forms have associated labels for screen readers
- [ ] **A11Y-05**: Page structure uses semantic HTML (headings, landmarks, ARIA where needed)
- [ ] **A11Y-06**: Keyboard navigation works for all interactive elements (tab order, escape)
- [ ] **A11Y-07**: Screen reader announcements for dynamic content (loading, errors, success)

### Performance (Lighthouse CI)

- [ ] **PERF-01**: Lighthouse CI integrated into CI/CD pipeline
- [ ] **PERF-02**: Performance score threshold ≥80% enforced on PRs
- [ ] **PERF-03**: Accessibility score threshold ≥90% enforced on PRs
- [ ] **PERF-04**: Core Web Vitals dashboard tracks LCP, FID, CLS metrics
- [ ] **PERF-05**: Bundle size monitoring with budget alerts (max 500KB initial load)

### Testing (Load & Integration)

- [ ] **TEST-01**: k6 load test scripts cover critical paths (auth, VM create, VM start)
- [ ] **TEST-02**: Load tests simulate 50 concurrent users with sub-second response times
- [ ] **TEST-03**: Integration tests verify cross-phase flows (auth → create VM → access domain)
- [ ] **TEST-04**: End-to-end tests cover primary user workflows
- [ ] **TEST-05**: Load test results logged to CI/CD artifacts for trend analysis

---

## v2 Requirements (Deferred)

Deferred to future releases. Tracked but not in current roadmap.

### Advanced Security

- **SEC-05**: Per-user rate limiting (higher limits for authenticated users)
- **SEC-06**: Redis-backed distributed rate limiting
- **SEC-07**: Rate limit dashboard for admin monitoring

### Advanced Compliance

- **GDPR-07**: 30-day grace period with cancellation option
- **GDPR-08**: Data export before deletion (download user data)
- **GDPR-09**: Admin dashboard for pending deletions

### Advanced Documentation

- **DOC-06**: Interactive API examples with sample requests/responses
- **DOC-07**: Video tutorials for common workflows
- **DOC-08**: API changelog with version history

### Advanced Accessibility

- **A11Y-08**: Automated a11y testing in CI/CD (axe-core integration)
- **A11Y-09**: User testing with screen reader users
- **A11Y-10**: Accessibility statement page

### Advanced Performance

- **PERF-06**: Real User Monitoring (RUM) for production metrics
- **PERF-07**: Performance budgets per route
- **PERF-08**: Automatic image optimization pipeline

### Advanced Testing

- **TEST-06**: Automated load test execution in CI/CD
- **TEST-07**: Distributed load testing across multiple regions
- **TEST-08**: Chaos engineering experiments (VM failures, network partitions)

---

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Real VM (qemu/kvm) | Contradicts shared-resource model, high overhead |
| Multi-server cluster | Single-server constraint for v1.x |
| GPU acceleration | Niche use case, hardware limitation |
| Managed database service | Out of scope, users can run their own in VMs |
| Mobile app | Web-first, mobile later |
| Windows OS templates | Licensing complexity, resource-heavy |
| Third-party GDPR tools | Overkill for 500 users, manual implementation sufficient |
| Rate limiting on static assets | Cloudflare handles this at edge |
| Complex rate limit rules | Hard to debug, maintain — simple per-endpoint preferred |
| Overlay accessibility widgets | Ineffective, real fixes required |
| Real User Monitoring (RUM) | Defer to v2, Lighthouse CI sufficient for v1.1 |

---

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| SEC-01 | Phase 6 | Pending |
| SEC-02 | Phase 6 | Pending |
| SEC-03 | Phase 6 | Pending |
| SEC-04 | Phase 6 | Pending |
| DOC-01 | Phase 7 | Pending |
| DOC-02 | Phase 7 | Pending |
| DOC-03 | Phase 7 | Pending |
| DOC-04 | Phase 7 | Pending |
| DOC-05 | Phase 7 | Pending |
| A11Y-01 | Phase 8 | Pending |
| A11Y-02 | Phase 8 | Pending |
| A11Y-03 | Phase 8 | Pending |
| A11Y-04 | Phase 8 | Pending |
| A11Y-05 | Phase 8 | Pending |
| A11Y-06 | Phase 8 | Pending |
| A11Y-07 | Phase 8 | Pending |
| PERF-01 | Phase 9 | Pending |
| PERF-02 | Phase 9 | Pending |
| PERF-03 | Phase 9 | Pending |
| PERF-04 | Phase 9 | Pending |
| PERF-05 | Phase 9 | Pending |
| GDPR-01 | Phase 10 | Pending |
| GDPR-02 | Phase 10 | Pending |
| GDPR-03 | Phase 10 | Pending |
| GDPR-04 | Phase 10 | Pending |
| GDPR-05 | Phase 10 | Pending |
| GDPR-06 | Phase 10 | Pending |
| TEST-01 | Phase 11 | Pending |
| TEST-02 | Phase 11 | Pending |
| TEST-03 | Phase 11 | Pending |
| TEST-04 | Phase 11 | Pending |
| TEST-05 | Phase 11 | Pending |

**Coverage:**
- v1.1 requirements: 31 total
- Mapped to phases: 31 ✓
- Unmapped: 0 ✓

---

*Requirements defined: 2026-03-30*
*Last updated: 2026-03-30 after v1.1 roadmap created*

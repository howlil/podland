# Roadmap: Podland

**Created:** 2026-03-25
**Current Milestone:** v1.1 — Hardening & Polish 🔄 IN PROGRESS

---

## Milestones

- ✅ **v1.0** — Phases 1-5 (shipped 2026-03-30) — [Archive](milestones/v1.0-ROADMAP.md)
- 📋 **v1.1** — Phases 6-11 (planning) — Hardening & Polish

---

## Phases

<details>
<summary>✅ v1.0 — Foundation to Admin + Polish (Phases 1-5) — SHIPPED 2026-03-30</summary>

- [x] Phase 1: Foundation — Complete
- [x] Phase 2: Core VM — Complete
- [x] Phase 3: Networking — Complete
- [x] Phase 4: Observability — Complete
- [x] Phase 5: Admin + Polish — Complete

**Details:** See [v1.0 Archive](milestones/v1.0-ROADMAP.md)

</details>

<details>
<summary>📋 v1.1 — Hardening & Polish (Phases 6-11) — PLANNING</summary>

- [ ] Phase 6: Security (Rate Limiting) — 4 requirements
- [ ] Phase 7: Documentation (OpenAPI) — 5 requirements
- [ ] Phase 8: Accessibility (WCAG AA) — 7 requirements
- [ ] Phase 9: Performance (Lighthouse CI) — 5 requirements
- [ ] Phase 10: Compliance (GDPR) — 6 requirements
- [ ] Phase 11: Testing (Load & Integration) — 5 requirements

**Details:** See below

</details>

---

## Progress

| Phase | Name | Milestone | Status | Completed |
|-------|------|-----------|--------|-----------|
| 1 | Foundation | v1.0 | ✅ Complete | 2026-03-27 |
| 2 | Core VM | v1.0 | ✅ Complete | 2026-03-28 |
| 3 | Networking | v1.0 | ✅ Complete | 2026-03-28 |
| 4 | Observability | v1.0 | ✅ Complete | 2026-03-29 |
| 5 | Admin + Polish | v1.0 | ✅ Complete | 2026-03-30 |
| 6 | Security | v1.1 | 📋 Planned | — |
| 7 | Documentation | v1.1 | 📋 Planned | — |
| 8 | Accessibility | v1.1 | 📋 Planned | — |
| 9 | Performance | v1.1 | 📋 Planned | — |
| 10 | Compliance | v1.1 | 📋 Planned | — |
| 11 | Testing | v1.1 | 📋 Planned | — |

**Total:** 5/11 phases complete (45%)

---

## v1.1 Phase Details

### Phase 6: Security (Rate Limiting)

**Goal:** Protect auth endpoints from brute force attacks with rate limiting.

**Requirements:**
- SEC-01: Rate limit 5 req/min on `/api/auth/*` per IP
- SEC-02: HTTP 429 with `Retry-After` header
- SEC-03: `X-Forwarded-For` IP identification
- SEC-04: Configurable via `RATE_LIMIT_AUTH` env var

**Success Criteria:**
1. Auth endpoint returns 429 after 5 requests within 1 minute from same IP
2. Response includes `Retry-After: 60` header
3. Legitimate users behind NAT not affected (Cloudflare IP headers respected)
4. Rate limit threshold can be changed without code deployment

**Timeline:** T+7

---

### Phase 7: Documentation (OpenAPI)

**Goal:** Provide interactive API documentation for developers.

**Requirements:**
- DOC-01: OpenAPI/Swagger specification for all endpoints
- DOC-02: Swagger UI at `/api/docs`
- DOC-03: Auth, error codes, rate limits documented
- DOC-04: User guide (sign in, create VM, domain access)
- DOC-05: Help menu link to user guide

**Success Criteria:**
1. All 20+ API endpoints documented with request/response schemas
2. Swagger UI renders at `/api/docs` with Try It Out functionality
3. User guide covers 3 primary workflows with screenshots
4. Help menu in dashboard links to documentation

**Timeline:** T+14

---

### Phase 8: Accessibility (WCAG AA)

**Goal:** Achieve WCAG AA compliance for inclusive user experience.

**Requirements:**
- A11Y-01: Visible focus indicators (3px outline minimum)
- A11Y-02: Color contrast ≥4.5:1 text, ≥3:1 UI
- A11Y-03: Alt text for all images
- A11Y-04: Form labels for screen readers
- A11Y-05: Semantic HTML structure
- A11Y-06: Full keyboard navigation
- A11Y-07: Screen reader announcements

**Success Criteria:**
1. axe-core reports 0 violations on all pages
2. Tab navigation reaches all interactive elements with visible focus
3. Color contrast analyzer passes all text/UI elements
4. Screen reader (NVDA/VoiceOver) can navigate all workflows
5. Keyboard-only user can complete VM creation flow

**Timeline:** T+21

---

### Phase 9: Performance (Lighthouse CI)

**Goal:** Monitor and enforce performance standards in CI/CD.

**Requirements:**
- PERF-01: Lighthouse CI in CI/CD pipeline
- PERF-02: Performance ≥80% threshold
- PERF-03: Accessibility ≥90% threshold
- PERF-04: Core Web Vitals dashboard
- PERF-05: Bundle size ≤500KB initial load

**Success Criteria:**
1. Lighthouse CI runs on every PR with status checks
2. PRs failing performance/a11y thresholds blocked from merge
3. Grafana dashboard shows LCP <2.5s, FID <100ms, CLS <0.1
4. Bundle analyzer reports included in build artifacts
5. Performance trends visible over time

**Timeline:** T+21

---

### Phase 10: Compliance (GDPR)

**Goal:** Enable GDPR right to erasure with account deletion.

**Requirements:**
- GDPR-01: Account deletion request from settings
- GDPR-02: PII anonymization on deletion
- GDPR-03: Session invalidation
- GDPR-04: VM cascade delete
- GDPR-05: Audit logging
- GDPR-06: Confirmation email

**Success Criteria:**
1. User can click "Delete Account" in settings with confirmation dialog
2. After deletion, email/name/GitHub ID show as "[deleted]"
3. All sessions invalidated (user must re-authenticate)
4. All user VMs deleted with account
5. Audit log shows "Account deleted by user [timestamp]"
6. Confirmation email received within 5 minutes

**Timeline:** T+30

---

### Phase 11: Testing (Load & Integration)

**Goal:** Validate system performance under load with comprehensive testing.

**Requirements:**
- TEST-01: k6 scripts for critical paths
- TEST-02: 50 concurrent users simulation
- TEST-03: Integration tests for cross-phase flows
- TEST-04: E2E tests for primary workflows
- TEST-05: Load test results in CI artifacts

**Success Criteria:**
1. k6 scripts cover auth, VM create, VM start endpoints
2. Load test shows <1s response time at 50 concurrent users
3. Integration tests verify auth → create VM → domain access flow
4. E2E tests cover sign in, create VM, view dashboard, delete VM
5. Load test reports archived in CI/CD for trend analysis

**Timeline:** T+30

---

## Requirement Traceability

All 31 v1.1 requirements mapped to phases:

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

_For v1.0 roadmap details, see [v1.0 Archive](milestones/v1.0-ROADMAP.md)_

# PITFALLS.md — Common Mistakes for v1.1 Hardening

**Milestone:** v1.1 — Hardening & Polish
**Date:** 2026-03-30
**Context:** Adding production hardening features to existing v1.0 codebase

---

## Executive Summary

Each v1.1 feature area has common pitfalls that can cause security issues, compliance failures, or degraded user experience. This document identifies mistakes, warning signs, prevention strategies, and which phase should address each pitfall.

| Feature Area | Critical Pitfalls | Common Pitfalls | Prevention Priority |
|--------------|------------------|-----------------|---------------------|
| Rate Limiting | Wrong middleware order, IP spoofing | Over-limiting, no headers | T+7 |
| GDPR Deletion | Incomplete cascade, audit gaps | No grace period, UX issues | T+30 |
| OpenAPI | Annotation drift, security gaps | Over-documentation | T+30 |
| WCAG AA | Overlay widgets, removed focus | Color-only states | T+30 |
| Lighthouse CI | Unrealistic thresholds, no budgets | Ignoring mobile scores | T+30 |
| Load Testing | Production testing, no baseline | Ignoring percentiles | T+30 |

---

## 1. Rate Limiting Pitfalls

### Critical Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Middleware order wrong** | Rate limiting bypassed | Logs show requests skipping rate limit | Apply rate limit BEFORE auth middleware | T+7 |
| **Trusting X-Forwarded-For blindly** | IP spoofing, bypass | Same IP with multiple X-Forwarded-For values | Only trust headers from Cloudflare (CF-Connecting-IP) | T+7 |
| **Rate limiting after authentication** | Wasted resources on invalid requests | High CPU on rejected requests | Rate limit at edge, before auth logic | T+7 |
| **No rate limit headers** | Users can't debug 429 errors | Support tickets about "random" failures | Include `X-RateLimit-*` and `Retry-After` headers | T+7 |

### Common Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Rate limiting static assets** | Wasted resources, slow pages | High rate limit counter on CSS/JS | Skip rate limiting on `/static/*`, `/*.js`, `/*.css` | T+7 |
| **Global rate limit too low** | Legitimate users blocked | 429 errors on dashboard load | Set API limit to 100/min, auth to 5/min | T+7 |
| **In-memory store in production** | Limits reset on restart | Attackers can wait for restart | Use Redis for production (v1.2) | T+30 |
| **No monitoring on rate limits** | Attacks undetected | No alerts on sustained 429s | Add metrics: `rate_limit_hits_total` | T+30 |

### Prevention Checklist

```go
// ✅ Correct middleware order
r.Use(CORSMiddleware)           // Outermost
r.Use(CSRFMiddleware)
r.Use(RecoveryMiddleware)

// Rate limit BEFORE auth
r.Route("/api/auth", func(r chi.Router) {
    r.Use(NewAuthRateLimiter()) // 5 req/min - applied first
    r.Use(AuthMiddleware)       // Auth validated after rate limit
    r.Get("/login", LoginHandler)
})

// ✅ Use Cloudflare header, not X-Forwarded-For
func getClientIP(r *http.Request) string {
    // Cloudflare sets CF-Connecting-IP
    if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
        return ip
    }
    // Fallback to RemoteAddr
    return r.RemoteAddr
}

// ✅ Include rate limit headers
w.Header().Set("X-RateLimit-Limit", "5")
w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
w.Header().Set("Retry-After", strconv.Itoa(resetSeconds))
```

### Red Flags

🚩 **Stop and fix immediately if you see:**
- Rate limiting applied AFTER authentication
- Using `X-Forwarded-For` without verifying Cloudflare proxy
- No `Retry-After` header on 429 responses
- Rate limiting on `/api/docs` or static files

---

## 2. GDPR Account Deletion Pitfalls

### Critical Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Hard delete user record** | Broken foreign keys, lost audit trail | Database errors, missing audit logs | Soft delete + anonymize, preserve user ID | T+30 |
| **Incomplete cascade deletion** | Orphaned data, privacy violation | Sessions exist for deleted users | Explicitly delete sessions, API keys, VMs | T+30 |
| **No audit logging of deletion** | Compliance failure | Can't prove deletion occurred | Log deletion event with timestamp, method | T+30 |
| **Deletion doesn't invalidate sessions** | User can still access account | Deleted user sessions still valid | Delete sessions FIRST, before anonymizing | T+30 |

### Common Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **No confirmation step** | Accidental deletions | Support tickets: "I clicked by mistake" | Add confirmation modal + email confirmation | T+30 |
| **Deleting audit logs** | Compliance violation | Audit trail gaps | Retain audit logs (anonymized) for 7 years | T+30 |
| **No VM cleanup** | Resource leak, orphaned namespaces | k8s namespaces without owners | Delete VMs and Kubernetes namespaces | T+30 |
| **External service data not cleaned** | Privacy violation | Profile images still accessible | Delete from SendGrid, blob storage, etc. | T+30 |

### Prevention Checklist

```go
// ✅ Correct deletion order
func DeleteAccount(userID string) error {
    tx, _ := db.BeginTx(ctx, nil)
    
    // 1. FIRST: Delete sessions (immediate invalidation)
    tx.Exec(`DELETE FROM sessions WHERE user_id = $1`, userID)
    
    // 2. Delete API keys
    tx.Exec(`DELETE FROM api_keys WHERE user_id = $1`, userID)
    
    // 3. Delete VMs (triggers k8s cleanup)
    vmService.DeleteAllUserVMs(ctx, tx, userID)
    
    // 4. THEN: Anonymize user (preserve referential integrity)
    tx.Exec(`
        UPDATE users 
        SET email = CONCAT('deleted-', id, '@deleted.local'),
            github_id = NULL,
            display_name = 'Deleted User',
            anonymized = TRUE,
            deleted_at = NOW()
        WHERE id = $1
    `, userID)
    
    // 5. Log audit event
    auditLogger.Log(AuditEvent{
        Type: "account_deleted",
        UserID: userID,
    })
    
    return tx.Commit()
}

// ✅ Retain audit logs (anonymized)
// DO NOT delete from audit_logs table
// Instead, the audit log references user_id which now points to anonymized record
```

### Red Flags

🚩 **Stop and fix immediately if you see:**
- `DELETE FROM users WHERE id = ?` (hard delete)
- Sessions not deleted before user anonymization
- Audit logs being deleted
- VMs not cleaned up (check k8s namespaces)
- No confirmation step before deletion

---

## 3. OpenAPI/Swagger Pitfalls

### Critical Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Annotations drift from code** | Docs show wrong parameters | 400 errors from "documented" requests | Run `swag init` in CI, fail on drift | T+30 |
| **Security schemes not documented** | Users can't authenticate | Swagger UI "Try it out" fails | Add `@securityDefinitions` annotation | T+30 |
| **Sensitive data in examples** | API keys, tokens exposed | Swagger JSON contains secrets | Use placeholder values: `"token": "Bearer {token}"` | T+30 |
| **Internal endpoints documented** | Attack surface exposed | `/admin/*` endpoints in public docs | Use `@Router` only for public endpoints | T+30 |

### Common Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Manual OpenAPI YAML editing** | Drift from implementation | YAML doesn't match actual API | Auto-generate only, never manual edit | T+30 |
| **Missing error responses** | Users don't know error formats | Only 200 responses documented | Document 400, 401, 403, 404, 500 for each endpoint | T+30 |
| **No request/response examples** | Users guess payload format | Support tickets about payload format | Add `@Example` annotations | T+30 |
| **Swagger UI exposed without auth** | API docs public to attackers | `/api/docs` accessible without login | Add auth middleware to Swagger route | T+30 |

### Prevention Checklist

```go
// ✅ Security scheme definition
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT token: "Bearer {token}"

// ✅ Document all responses
// @Success  200  {object}  VMResponse
// @Failure  400  {object}  ErrorResponse  "Invalid request body"
// @Failure  401  {object}  ErrorResponse  "Missing or invalid token"
// @Failure  403  {object}  ErrorResponse  "Insufficient permissions"
// @Failure  404  {object}  ErrorResponse  "VM not found"
// @Failure  500  {object}  ErrorResponse  "Internal server error"

// ✅ Use placeholder values (NEVER real secrets)
// @Param  Authorization  header  string  true  "Bearer {your_token}"

// ✅ CI check for annotation drift
// .github/workflows/docs.yml
// - name: Check OpenAPI drift
//   run: |
//     swag init -g cmd/main.go -o ./docs
//     git diff --exit-code apps/backend/docs/ || \
//       (echo "OpenAPI docs out of sync! Run 'swag init'" && exit 1)
```

### Red Flags

🚩 **Stop and fix immediately if you see:**
- Manual edits to `docs/swagger.json`
- Real API keys or tokens in examples
- `/api/docs` accessible without authentication
- Only success responses documented (no errors)
- `swag init` not in CI pipeline

---

## 4. WCAG AA Accessibility Pitfalls

### Critical Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Using overlay widgets** | Creates more barriers | axe-core shows new violations | Never use accessiBe, UserWay, etc. | T+30 |
| **Removing focus outlines** | Keyboard users lost | Can't see which element is focused | Style focus, don't remove (`outline: none`) | T+30 |
| **Color-only state indicators** | Colorblind users can't perceive | Error states only use red | Add icons + text: "Error: Invalid email" | T+30 |
| **Focus obscured by fixed header** | Users can't see focused element | Tab navigation loses position | Add `scroll-margin-top` to focused elements | T+30 |

### Common Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Light gray text on white** | Fails 4.5:1 contrast | Text hard to read, especially on mobile | Use Gray-600 (#4b5563) minimum | T+30 |
| **Links without underline** | Users can't identify links | Links look like regular text | Always underline links or use distinct style | T+30 |
| **Images without alt text** | Screen readers say "image" | Accessibility score drops | Add descriptive alt or `alt=""` for decorative | T+30 |
| **Form inputs without labels** | Screen readers can't identify | `<input>` without associated `<label>` | Use `<label for="id">` or `aria-label` | T+30 |
| **Heading levels skipped** | Screen reader navigation broken | h1 → h3 (no h2) | Maintain h1 → h2 → h3 hierarchy | T+30 |

### Prevention Checklist

```css
/* ✅ Correct focus indicator (DON'T remove) */
*:focus {
  outline: 3px solid #3b82f6;  /* Blue-500 */
  outline-offset: 2px;
}

/* ✅ Focus not obscured */
html {
  scroll-padding-top: 80px;  /* Fixed header height */
}

*:focus {
  scroll-margin-top: 90px;  /* Header + padding */
}

/* ✅ Sufficient contrast */
.text-muted {
  color: #4b5563;  /* Gray-600, 5.7:1 on white */
  /* NOT #999999 (2.8:1 - FAILS) */
}

/* ✅ Links not color-only */
a {
  color: #2563eb;
  text-decoration: underline;  /* Always underline */
  text-underline-offset: 2px;
}

/* ✅ Skip link for keyboard users */
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  background: #3b82f6;
  color: white;
  padding: 8px 16px;
  z-index: 100;
}

.skip-link:focus {
  top: 0;  /* Visible on focus */
}
```

```tsx
// ✅ Form with proper labels
<label htmlFor="email">Email Address</label>
<input 
  type="email" 
  id="email" 
  name="email"
  aria-required="true"
  aria-invalid={errors.email ? "true" : "false"}
  aria-describedby={errors.email ? "email-error" : undefined}
/>
{errors.email && (
  <span id="email-error" role="alert">
    <ExclamationIcon /> Invalid email format
  </span>
)}

// ✅ Image with alt text
<img 
  src="/chart.png" 
  alt="Sales increased 45% in Q4 2024, from $10K to $14.5K" 
/>

// ✅ Decorative image (screen readers skip)
<img 
  src="/decorative-border.png" 
  alt="" 
  role="presentation" 
/>
```

### Red Flags

🚩 **Stop and fix immediately if you see:**
- `outline: none` or `outline: 0` without replacement
- Gray text lighter than #6b7280 on white backgrounds
- Links that only differ by color (no underline)
- `<img>` tags without `alt` attributes
- Heading jumps (h1 → h4)
- Any accessibility overlay widget in package.json

---

## 5. Lighthouse CI Pitfalls

### Critical Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **100% score requirement** | PRs blocked for minor issues | Constant CI failures, team frustration | Start with 80% performance, 90% a11y | T+30 |
| **Testing only homepage** | Other pages degrade unnoticed | Dashboard has poor performance | Test multiple URLs: home, dashboard, VM list | T+30 |
| **No mobile testing** | Mobile users have poor experience | Desktop scores good, mobile unusable | Use `--emulated-form-factor=mobile` | T+30 |
| **Ignoring Core Web Vitals** | Real users experience poor performance | Good scores but slow page loads | Set LCP < 2.5s, CLS < 0.1, FID < 100ms | T+30 |

### Common Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Running on every commit** | CI slowdown, wasted resources | 30+ min CI times | Run on PR only, not every commit | T+30 |
| **No performance budgets** | Gradual degradation unnoticed | Scores slowly drop over months | Set budgets, alert on regression >10% | T+30 |
| **Temporary storage forever** | Reports lost after 30 days | Can't track historical trends | Plan LHCI server for v1.2 | T+30 |
| **Blocking PRs on day one** | Team resistance to Lighthouse | "Lighthouse is blocking my PR" | Start with `warn`, escalate to `error` later | T+30 |

### Prevention Checklist

```javascript
// ✅ Realistic thresholds
module.exports = {
  ci: {
    assert: {
      assertions: {
        // Start with warnings, not errors
        'categories:performance': ['warn', { minScore: 0.8 }],
        'categories:accessibility': ['error', { minScore: 0.9 }],
        
        // Core Web Vitals
        'largest-contentful-paint': ['warn', { maxNumericValue: 2500 }],
        'cumulative-layout-shift': ['warn', { maxNumericValue: 0.1 }],
        'total-blocking-time': ['warn', { maxNumericValue: 300 }],
        
        // Disable non-applicable audits
        'uses-rel-preload': 'off',
      },
    },
  },
};
```

```yaml
# ✅ Run on PR, not every commit
on:
  pull_request:
    branches: [main]
  # Not on: push (every commit)

# ✅ Test multiple pages
collect:
  url:
    - http://localhost:4173/
    - http://localhost:4173/dashboard
    - http://localhost:4173/dashboard/vms
    - http://localhost:4173/admin
```

### Red Flags

🚩 **Stop and fix immediately if you see:**
- `minScore: 1.0` or `minScore: 0.95` for performance
- Only testing homepage
- No Core Web Vitals thresholds
- Lighthouse running on every commit (CI slowdown)
- Blocking PRs before team is trained

---

## 6. Load Testing Pitfalls

### Critical Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Testing against production** | Data corruption, user impact | Test VMs appear in production | Use separate test environment | T+30 |
| **No baseline metrics** | Can't tell if performance regressed | "Is 500ms p95 good?" | Run baseline test before changes | T+30 |
| **Only testing happy path** | Errors under load unnoticed | 0% error rate in test, 10% in prod | Include error scenarios, rate limits | T+30 |
| **Ignoring percentiles** | Tail latency hidden | Average looks good, users complain | Track p95, p99, not just average | T+30 |

### Common Mistakes

| Mistake | Impact | Warning Signs | Prevention | Phase |
|---------|--------|---------------|------------|-------|
| **Short test duration** | Memory leaks, resource exhaustion missed | Test passes, prod crashes after 1 hour | Run load test for 15+ minutes | T+30 |
| **No ramp-up period** | Cold start skews results | First requests slow, rest fast | Use `ramping-vus` with gradual increase | T+30 |
| **Testing without auth** | Doesn't reflect real usage | All requests unauthenticated | Include login flow in test | T+30 |
| **No test data cleanup** | Database grows unbounded | Test DB larger than prod | Clean up test data after runs | T+30 |

### Prevention Checklist

```javascript
// ✅ Multiple scenarios with ramp-up
export const options = {
  scenarios: {
    smoke: {
      executor: 'constant-vus',
      vus: 10,
      duration: '1m',  // Quick validation
    },
    load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },   // Ramp up
        { duration: '5m', target: 50 },   // Sustain
        { duration: '2m', target: 0 },    // Ramp down
      ],
    },
    stress: {
      executor: 'ramping-vus',
      stages: [
        { duration: '5m', target: 100 },
        { duration: '5m', target: 200 },  // Breaking point
        { duration: '5m', target: 0 },
      ],
    },
  },
  thresholds: {
    // Track percentiles, not just average
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],  // <1% errors
  },
};

// ✅ Include auth flow
export default function () {
  // Login first
  const loginRes = http.post(`${BASE_URL}/api/auth/login`);
  const token = loginRes.json('access_token');
  const headers = { 'Authorization': `Bearer ${token}` };
  
  // Then test authenticated endpoints
  http.get(`${BASE_URL}/api/vms`, { headers });
}
```

### Red Flags

🚩 **Stop and fix immediately if you see:**
- `BASE_URL` pointing to production
- Only tracking `http_req_duration` average (no percentiles)
- Test duration < 5 minutes for load test
- No authentication in test scripts
- Test data not cleaned up after runs

---

## Pitfall Prevention Summary by Phase

| Phase | Feature | Critical Pitfalls to Prevent |
|-------|---------|------------------------------|
| **T+7** | Rate Limiting | Middleware order, IP spoofing, no headers |
| **T+14** | OpenAPI | Annotation drift, security gaps, sensitive data exposure |
| **T+21** | WCAG AA | Overlay widgets, removed focus, color-only states |
| **T+21** | Lighthouse CI | Unrealistic thresholds, no mobile testing |
| **T+30** | GDPR Deletion | Hard delete, incomplete cascade, session invalidation |
| **T+30** | Load Testing | Production testing, no baseline, ignoring percentiles |

---

## Red Flag Summary: Stop Immediately If You See

### Security (Rate Limiting)
- 🚩 Rate limit AFTER auth middleware
- 🚩 Using `X-Forwarded-For` without Cloudflare verification
- 🚩 No `Retry-After` header on 429

### Compliance (GDPR)
- 🚩 `DELETE FROM users` (hard delete)
- 🚩 Sessions not deleted first
- 🚩 Audit logs being deleted
- 🚩 VMs not cleaned up

### Documentation (OpenAPI)
- 🚩 Manual edits to `swagger.json`
- 🚩 Real secrets in examples
- 🚩 `/api/docs` without authentication

### Accessibility (WCAG AA)
- 🚩 `outline: none` without replacement
- 🚩 Gray text lighter than #6b7280
- 🚩 Links without underline
- 🚩 Accessibility overlay widgets

### Performance (Lighthouse)
- 🚩 `minScore: 1.0` requirements
- 🚩 Only testing homepage
- 🚩 Blocking PRs before team training

### Testing (Load)
- 🚩 Testing against production URL
- 🚩 No percentiles (p95, p99)
- 🚩 Test duration < 5 minutes

---

## Testing Checklist Before Each Feature Launch

### Rate Limiting (T+7)
- [ ] Middleware order verified (rate limit before auth)
- [ ] IP extraction uses Cloudflare header
- [ ] 429 responses include `Retry-After`
- [ ] Static assets excluded from rate limiting
- [ ] Load test shows rate limiting works under load

### GDPR Deletion (T+30)
- [ ] Sessions deleted first
- [ ] User record anonymized, not hard deleted
- [ ] Audit log entry created
- [ ] VMs and namespaces cleaned up
- [ ] External services notified (SendGrid, blob storage)

### OpenAPI (T+14)
- [ ] `swag init` in CI pipeline
- [ ] Security schemes documented
- [ ] No real secrets in examples
- [ ] All error responses documented
- [ ] Swagger UI requires authentication

### WCAG AA (T+21)
- [ ] axe-core shows 0 critical violations
- [ ] Keyboard navigation works (Tab through all pages)
- [ ] Color contrast passes 4.5:1 for text
- [ ] Focus indicators visible on all elements
- [ ] Screen reader test with NVDA or VoiceOver

### Lighthouse CI (T+21)
- [ ] Multiple pages tested
- [ ] Thresholds realistic (80% performance, 90% a11y)
- [ ] Core Web Vitals tracked
- [ ] Reports uploaded and accessible
- [ ] Team trained on interpreting results

### Load Testing (T+30)
- [ ] Test environment separate from production
- [ ] Baseline metrics recorded
- [ ] Auth flow included in tests
- [ ] p95 and p99 tracked
- [ ] Test data cleaned up after runs

---

*Research completed: 2026-03-30*
*All 4 v1.1 research reports complete: STACK.md, FEATURES.md, ARCHITECTURE.md, PITFALLS.md*

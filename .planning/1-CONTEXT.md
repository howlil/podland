# Phase 1 Context: Foundation

**Phase:** 1 — Authentication + Basic Dashboard  
**Created:** 2026-03-25  
**Status:** Decisions locked, ready for research and planning

---

## Decision Summary

All implementation decisions for Phase 1 requirements. Downstream agents (researcher, planner) use this to know what choices are locked.

---

## 1. OAuth Flow & Email Validation

**Requirements:** AUTH-01, AUTH-02, AUTH-03

### Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Multiple emails** | Require primary = student email | Simplest implementation, clear user guidance |
| **Rejection flow** | Detailed guide with retry link | Educational, reduces support burden |
| **First-time onboarding** | Welcome screen with terms acceptance | Legal protection, confirms user understands platform |
| **NIM extraction** | Extract + user confirmation UI | Handles edge cases (email typos, format changes) |

### Implementation Notes

- **OAuth scope:** Request `user:email` scope to access all GitHub emails
- **Primary email check:** GitHub API returns `primary` flag — use this
- **Rejection page:** Show exact error ("Email must end with @student.unand.ac.id"), link to GitHub email settings (`https://github.com/settings/emails`), "Retry" button that re-validates without re-authorizing
- **Welcome screen:** Display extracted NIM, role (Internal/External), terms checkbox, "Activate Account" button
- **NIM confirmation:** If user edits NIM, store edited value (don't re-extract from email)

### User Flow

```
1. User clicks "Sign in with GitHub"
2. GitHub OAuth redirect → backend exchanges code for token
3. Backend fetches user info + emails from GitHub API
4. Check if primary email ends with @student.unand.ac.id
   - NO → Show rejection page with guide + retry button
   - YES → Extract NIM, proceed to step 5
5. Check if user exists in database
   - NEW → Show welcome screen (NIM, role, terms), create account on accept
   - EXISTS → Create session, redirect to dashboard
```

---

## 2. Profile Data

**Requirements:** AUTH-05

### Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Display name** | Always fetch from GitHub | Single source of truth, no sync issues |
| **Email** | Store from GitHub (read-only) | Needed for NIM extraction, stable reference |
| **Avatar** | Download on sign-in, store locally | Control over caching, no broken hotlinks |
| **NIM** | Extract once, store confirmed value | Performance, user can correct if wrong |
| **Profile editability** | All fields read-only | GitHub is source of truth, simplifies logic |

### Data Model

```typescript
interface User {
  id: string
  githubId: string
  email: string           // Stored from GitHub, read-only
  displayName: string     // Fetched from GitHub on each load
  avatarUrl: string       // Local storage path (downloaded on sign-in)
  nim: string             // Extracted + confirmed by user
  role: 'internal' | 'external' | 'superadmin'  // Auto-assigned from NIM
  createdAt: Date
  updatedAt: Date
}
```

### Avatar Strategy

- **Storage:** Save to `/uploads/avatars/{userId}.{ext}`
- **Sync:** On each sign-in, compare GitHub avatar URL hash with local copy
  - Changed → Download new avatar, update local file
  - Unchanged → Use cached copy
- **Fallback:** If download fails, hotlink GitHub URL temporarily

### NIM Confirmation Flow

```
1. User signs in with valid student email
2. System extracts NIM: 221152001@... → NIM = "221152001"
3. Welcome screen shows: "Your NIM: 221152001 — Is this correct?"
   - [Yes, Confirm] → Store NIM, activate account
   - [Edit] → Text input, user types correct NIM, [Confirm]
4. NIM stored in database, cannot be changed without admin intervention
```

### Role Assignment Logic

```
NIM contains "1152" → Internal (SI UNAND student)
NIM does not contain "1152" → External (other Unand student)
```

---

## 3. Dashboard Layout

**Requirements:** DASH-01, DASH-03, DASH-04

### Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Navigation** | Dashboard + Profile only (Phase 1) | VMs menu hidden until Phase 2 |
| **Widget priority** | 1. Quota bar, 2. VM count, 3. Activity | Quota most important for resource awareness |
| **Dark mode** | System preference auto-detect only | No manual toggle needed for Phase 1 |
| **Mobile layout** | Bottom tab bar on mobile, sidebar on desktop | Mobile-first UX, familiar pattern |

### Navigation Structure

**Desktop (sidebar):**
```
┌─────────────────────────────────────────┐
│  Podland           [User Avatar] ▼     │
├──────────┬──────────────────────────────┤
│          │                              │
│  ────────│  Dashboard Home              │
│  📊 Dash │                              │
│  👤 Prof │                              │
│          │                              │
│          │                              │
└──────────┴──────────────────────────────┘
```

**Mobile (bottom tab bar):**
```
┌─────────────────────────┐
│  Dashboard Content      │
│  (stacked widgets)      │
│                         │
│                         │
├──────────┬──────────────┤
│ 📊 Dash  │  👤 Profile  │
└──────────┴──────────────┘
```

### Widget Layout (Desktop)

```
┌─────────────────────────────────────────────────────┐
│  Dashboard                                          │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │  Quota Usage                                │   │
│  │  ████████░░░░░░░░░░░░░░░░  0.3/1 CPU       │   │
│  │  ████████░░░░░░░░░░░░░░░░  512MB/2GB RAM   │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  ┌──────────────┐  ┌──────────────────────────┐   │
│  │  0 VMs       │  │  Recent Activity         │   │
│  │  Running     │  │  • Account created       │   │
│  │              │  │  • Signed in             │   │
│  └──────────────┘  └──────────────────────────┘   │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Responsive Behavior

- **Desktop (≥1024px):** Sidebar + 2-column widget grid
- **Tablet (768px-1023px):** Sidebar + 1-column widget stack
- **Mobile (<768px):** Bottom tab bar + single column widgets

### Dark Mode

- Use CSS `@media (prefers-color-scheme: dark)` for auto-detection
- No manual toggle in Phase 1
- Tailwind v4 `dark:` variants with `class` strategy (system class via JS)

---

## 4. Session & Refresh

**Requirements:** AUTH-04, AUTH-06

### Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Refresh strategy** | Both: silent + on-demand fallback | Best UX, handles edge cases |
| **Token storage** | HTTP-only cookie | Secure, CSRF protection required |
| **Invalidation** | Logout → current session only | Expected behavior, simpler implementation |
| **Concurrent sessions** | Limit to 3 per user | Prevent abuse, allow multi-device |

### Token Architecture

```
Access Token:
- Type: JWT
- Expiry: 15 minutes
- Storage: Memory (React state)
- Usage: API requests

Refresh Token:
- Type: Opaque (random string)
- Expiry: 7 days
- Storage: HTTP-only cookie (secure, sameSite=strict)
- Usage: Get new access token
```

### Refresh Flow

```
1. User signs in → backend returns access token + refresh cookie
2. Frontend stores access token in memory
3. Silent refresh: At 50% expiry (7.5 min), call /api/auth/refresh
   - Success → New access token, update state
   - Failure → Proceed to step 4
4. On-demand refresh: If API returns 401:
   - Call /api/auth/refresh
   - Retry original request with new token
   - If refresh fails → Redirect to login
```

### CSRF Protection

- Use double-submit cookie pattern
- Backend sets `XSRF-TOKEN` cookie (JavaScript-readable)
- Frontend reads token, sends in `X-XSRF-TOKEN` header
- Backend validates header matches cookie

### Session Management

```
Session Table:
- id (PK)
- userId (FK)
- refreshToken (hash)
- deviceInfo (user agent, IP)
- createdAt
- expiresAt

Constraints:
- Max 3 active sessions per user
- Oldest session revoked when 4th device signs in
- Logout → Delete current session only
```

### Session Invalidation Triggers

| Trigger | Action |
|---------|--------|
| User clicks logout | Delete current session, clear cookie |
| Session expires (7 days) | Auto-delete, user must re-login |
| Admin bans user | Delete all sessions (future enhancement) |
| User changes password | No action (Phase 1) |

---

## Code Context

### Integration Points

**Backend (Go):**
- OAuth callback handler (`/api/auth/github/callback`)
- Token generation (JWT access + opaque refresh)
- Email validation middleware
- NIM extraction utility
- Session management (create, validate, revoke)

**Frontend (TanStack Start):**
- Sign in button component
- Welcome screen (NIM confirmation + terms)
- Dashboard layout (sidebar + widgets)
- Profile page (read-only fields)
- Auth hook (token refresh logic)

**Database (PostgreSQL):**
- `users` table (id, github_id, email, display_name, avatar_url, nim, role, created_at, updated_at)
- `sessions` table (id, user_id, refresh_token_hash, device_info, created_at, expires_at)

**Storage:**
- `/uploads/avatars/{userId}.{ext}` — User avatar files

---

## Deferred Ideas

These were suggested but are **out of scope** for Phase 1:

- Custom avatar upload (Phase 2+)
- Manual dark mode toggle (Phase 2+)
- Session management UI (show active devices, revoke) (Phase 5)
- Bio/faculty/major fields (Phase 2+)
- Editable display name (Phase 2+)

---

## Next Steps

1. **Research:** Investigate GitHub OAuth Go libraries, TanStack Start auth patterns, CSRF protection strategies
2. **Planning:** Create detailed implementation plan with tasks, estimates, and acceptance criteria
3. **Implementation:** Execute plan, write tests, verify against success criteria

---

*Context created: 2026-03-25*  
*Ready for: Research or Planning*

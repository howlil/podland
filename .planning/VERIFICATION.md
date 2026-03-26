# Phase 1 Verification Checklist

## Success Criteria Verification

### Authentication (AUTH-01 to AUTH-06)

- [ ] **AUTH-01**: User can sign in via GitHub OAuth
  - Test: Click "Sign in with GitHub", complete OAuth flow
  - Expected: Redirected to dashboard after authorization

- [ ] **AUTH-02**: System validates GitHub email matches @student.unand.ac.id
  - Test: Sign in with non-student email
  - Expected: Rejection page with instructions

- [ ] **AUTH-03**: System extracts NIM and assigns role
  - Test: Sign in with 221152001@student.unand.ac.id
  - Expected: Role = "internal"
  - Test: Sign in with 221153001@student.unand.ac.id
  - Expected: Role = "external"

- [ ] **AUTH-04**: Session persists across browser refresh
  - Test: Sign in, refresh browser
  - Expected: Still logged in, dashboard visible

- [ ] **AUTH-05**: User can view their profile
  - Test: Navigate to Profile page
  - Expected: Display name, email, NIM, role visible

- [ ] **AUTH-06**: User can sign out
  - Test: Click sign out in user dropdown
  - Expected: Redirected to landing page, session cleared

### Dashboard (DASH-01, DASH-03, DASH-04)

- [ ] **DASH-01**: Dashboard shows VM summary and quota usage
  - Test: View dashboard home
  - Expected: Quota bar (0/0.5 CPU, 0/1GB RAM for External)
  - Expected: VM count card (0 VMs)

- [ ] **DASH-03**: Recent activity log visible
  - Test: View dashboard after sign in
  - Expected: "Account created" and "Signed in" entries

- [ ] **DASH-04**: Dashboard is responsive
  - Test: Resize browser to 320px width
  - Expected: Bottom tab bar visible, widgets stacked
  - Test: Resize to 1920px width
  - Expected: Sidebar visible, 2-column grid

## Manual Testing Steps

### 1. Fresh User Sign Up Flow

```
1. Open http://localhost:3000
2. Click "Sign in with GitHub"
3. Authorize application on GitHub
4. If email is valid:
   - See welcome screen with extracted NIM
   - Accept terms and confirm NIM
   - Redirected to dashboard
5. If email is invalid:
   - See rejection page
   - Instructions to add student email
   - "Retry" button
```

### 2. Returning User Sign In

```
1. Open http://localhost:3000 (already signed in)
2. Should auto-redirect to dashboard
3. Or click "Sign in" if logged out
4. GitHub OAuth (skip if already authorized)
5. Redirected to dashboard
```

### 3. Profile Page

```
1. Navigate to Profile (sidebar or mobile tab)
2. Verify fields:
   - Display name (from GitHub)
   - Email (read-only)
   - NIM (read-only)
   - Role (read-only)
   - Member since (auto date)
```

### 4. Sign Out

```
1. Click user avatar in sidebar
2. Dropdown opens
3. Click "Sign out"
4. Redirected to landing page
5. Session cookies cleared
```

### 5. Responsive Testing

**Mobile (320px):**
- [ ] Bottom tab bar visible
- [ ] Dashboard widgets stack vertically
- [ ] Profile page readable
- [ ] User dropdown works

**Tablet (768px):**
- [ ] Sidebar visible
- [ ] 2-column widget grid
- [ ] All interactions work

**Desktop (1920px):**
- [ ] Sidebar visible
- [ ] 2-column widget grid
- [ ] User dropdown positioned correctly

### 6. Dark Mode

- [ ] System dark mode detected
- [ ] All components styled for dark mode
- [ ] No white flashes on load

## API Testing

### Health Check
```bash
curl http://localhost:8080/api/health
# Expected: {"status":"ok",...}
```

### Auth Flow (Manual)
```bash
# 1. Initiate login
curl -L http://localhost:8080/api/auth/login
# Expected: Redirect to GitHub

# 2. After OAuth, check cookies
curl -v http://localhost:8080/api/users/me \
  -H "Authorization: Bearer <token>"
# Expected: User JSON
```

## Performance Checks

- [ ] Page loads in < 2 seconds
- [ ] OAuth redirect completes in < 5 seconds
- [ ] Dashboard renders in < 1 second
- [ ] No console errors

## Security Checks

- [ ] Refresh token cookie is HTTP-only
- [ ] Refresh token cookie has Secure flag (production)
- [ ] Refresh token cookie has SameSite=Strict
- [ ] CSRF token required for POST requests
- [ ] JWT expires after 15 minutes
- [ ] Refresh token expires after 7 days
- [ ] Max 3 concurrent sessions enforced

## Browser Compatibility

- [ ] Chrome/Edge (latest)
- [ ] Firefox (latest)
- [ ] Safari (latest)

## Known Issues

| Issue | Severity | Workaround | Status |
|-------|----------|------------|--------|
| None | - | - | - |

---

**Phase 1 Sign-off:**

- [ ] All success criteria verified
- [ ] All manual tests passed
- [ ] No critical bugs
- [ ] Ready for Phase 2

*Verification date: _______________*
*Verified by: _______________*

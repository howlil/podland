# Quick Task: Frontend Gap Analysis & Fixes

**Requested:** FE Architect/Engineer analysis with UI/UX max, product engineer mindset
**Context:** PaaS platform @apps/frontend
**Mode:** Quick (ad-hoc, atomic commits, STATE.md tracking)

---

## Gap Analysis (as of 2026-03-30)

### 🎨 UI/UX Issues

1. **Inconsistent Loading States**
   - Root route shows static loading spinner, no skeleton screens
   - VM table shows "Loading VMs..." text instead of skeleton rows
   - No optimistic UI updates for mutations

2. **Accessibility Gaps**
   - Missing ARIA labels on sort buttons in VM table
   - Color contrast issues with some badge variants
   - No keyboard navigation hints for interactive elements
   - Screen reader announcements missing for dynamic content

3. **Visual Polish**
   - Hardcoded emoji icons in DashboardLayout (📊, 💻, 👤) instead of Lucide icons
   - Inconsistent button variants (some use Tailwind classes, some use Button component)
   - No hover/active states on table rows for better affordance
   - Missing empty state illustrations

4. **Mobile UX**
   - Table overflows on small screens without horizontal scroll container
   - Action buttons in VM table may be too small for touch targets (min 44px)
   - Mobile nav doesn't show notification bell

### 🔧 Technical Debt

1. **Type Safety**
   - `any` usage in vite.config.ts plugins array
   - Loose typing in observability route search params
   - Missing error types for API calls

2. **Error Handling**
   - Generic error messages throughout
   - No error boundary components
   - Missing retry logic for failed queries
   - No toast/notification system for errors

3. **Performance**
   - VM polling at 5s interval regardless of tab visibility
   - No pagination for VM list (will degrade with many VMs)
   - No memoization for expensive computations (formatBytes called repeatedly)

4. **Code Quality**
   - Duplicated formatBytes function (in VMs route and wizard)
   - Magic numbers for polling intervals, refresh timers
   - Hardcoded tier data in CreateVMWizard instead of API fetch
   - Inconsistent naming: `vms` route uses `-vms` prefix

5. **Security**
   - XSS risk: VM name interpolated into domain preview without sanitization
   - No rate limiting UI feedback for create actions
   - CSRF token extraction from cookies could fail silently

### 📦 Missing Features (PaaS Context)

1. **Observability**
   - AlertHistory component is a stub
   - No real metrics integration (placeholder only)
   - LogViewer likely not implemented

2. **User Experience**
   - No onboarding flow for first-time users
   - No quota visualization before VM creation
   - Missing VM connection guide after creation
   - No bulk actions for VMs

3. **Admin Features**
   - Admin routes exist but likely incomplete
   - No audit log viewer implementation
   - Health check page is basic

---

## Priority Fixes (Quick Wins)

### P0 - Critical
1. Add error boundary and proper error handling
2. Fix accessibility (ARIA labels, focus management)
3. Add proper loading skeletons
4. Implement toast notification system

### P1 - High
1. Consolidate utility functions (formatBytes)
2. Add Visibility API for polling optimization
3. Implement proper empty states
4. Replace emoji with Lucide icons

### P2 - Medium
1. Add pagination to VM list
2. Implement real metrics/logs fetching
3. Add onboarding flow
4. Mobile table improvements

---

## Execution Plan

1. **Setup** - Add toast library, error boundary
2. **Accessibility** - Fix ARIA, focus states
3. **UX Polish** - Skeletons, empty states, icons
4. **Code Quality** - Dedupe utils, add types
5. **Performance** - Visibility API, memoization

# Quick Task: Fix Frontend Import Issues

## Task Description
Fix missing/broken imports in the frontend application identified through codebase analysis.

## Issues Found

### 1. Path Alias Mismatch (Critical)
- **Problem:** Admin routes use `~/*` alias but only `@/*` is configured in tsconfig
- **Files affected:**
  - `src/routes/admin/audit-log.tsx`
  - `src/routes/admin/health.tsx`
  - `src/routes/admin/index.tsx`
  - `src/routes/admin/users.tsx`

### 2. Missing UI Components
- `src/components/ui/badge.tsx` - used in admin routes
- `src/components/ui/progress.tsx` - used in health route

### 3. Wrong Relative Imports
- `src/routes/dashboard/observability/index.tsx` uses incorrect relative paths

### 4. Unused Imports
- `Filter` in `LogViewer.tsx`
- `refcase` in `MetricsSummary.tsx`

### 5. Event Listener Type Issues
- `CreateVMWizard.tsx` has keydown event type mismatch

## Resolution Plan

1. Add `~/*` path alias to tsconfig.json (mapping to `./src/*`)
2. Create missing UI components (badge.tsx, progress.tsx)
3. Fix relative imports in observability route
4. Remove unused imports
5. Fix event listener types

## Execution

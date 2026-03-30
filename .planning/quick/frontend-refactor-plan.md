# Frontend Refactor Plan - Clean Code, DRY, YAGNI, SOLID

**Date:** 2026-03-31  
**Mode:** Quick Task  
**Focus:** Code Quality & Architecture Excellence

---

## Executive Summary

Reviewed the frontend codebase and identified key areas for refactoring to achieve clean code principles:

### ✅ What's Already Good

1. **Custom Hooks Pattern** - `useVMs`, `useVM`, `useDashboard`, `useAdminUsers`, `useObservability`
2. **Zustand Store** - `useUIStore` for client state
3. **Component Architecture** - Container/Presentational pattern in core pages
4. **TypeScript** - Proper interfaces and type safety
5. **TanStack Query** - Server state management with mutations
6. **Toast Notifications** - Consistent error/success handling

### 🔴 Critical Issues Found (FIXED)

1. ✅ **Duplicate `export default`** - Fixed in 4 page files
2. ✅ **TypeScript errors (38 → 0)** - All compilation errors resolved
3. ✅ **Unused imports/variables** - Cleaned up across all files
4. ✅ **Route import mismatches** - Fixed named vs default exports

---

## Code Quality Issues Identified

### 1. **DRY Violations**

#### Issue: Duplicate Status Color Logic
**Location:** Multiple components
- `VMHeader.tsx` (lines 33-45)
- `VMTable.tsx` (lines 77-89)
- `VMsPage.tsx` (lines 167-174)
- `ResourceMetrics.tsx`

**Current:**
```typescript
// Repeated in 4+ places
const getStatusColor = (status: string) => {
  switch (status) {
    case "running": return "bg-green-100 text-green-800...";
    case "stopped": return "bg-gray-100 text-gray-800...";
    // ...
  }
};
```

**Solution:** Create utility function
```typescript
// src/lib/vm-utils.ts
export function getVMStatusStyles(status: VM['status']): string {
  const styles = {
    running: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
    stopped: "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300",
    pending: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400",
    error: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
  };
  return styles[status] || styles.stopped;
}
```

**Impact:** 🟡 Medium  
**Effort:** 🟢 Low (1 hour)

---

#### Issue: Duplicate Date Formatting
**Location:** Multiple pages
- `AdminUsersPage.tsx`: `new Date(user.created_at).toLocaleDateString()`
- `VMsPage.tsx`: `new Date(vm.created_at).toLocaleDateString()`
- `AdminAuditLogPage.tsx`: `new Date(log.created_at).toLocaleString()`

**Solution:**
```typescript
// src/lib/utils.ts
export function formatDate(date: string | Date, format: 'date' | 'datetime' = 'date'): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return format === 'datetime' 
    ? d.toLocaleString() 
    : d.toLocaleDateString();
}
```

**Impact:** 🟢 Low  
**Effort:** 🟢 Low (30 min)

---

### 2. **SOLID Violations**

#### SRP Violation: ObservabilityPage
**Location:** `pages/ObservabilityPage.tsx`

**Current:** 150+ lines, handles:
- VM selection logic
- Tab navigation
- Metrics display
- Logs display  
- Alerts display
- Grafana integration

**Solution:** Extract container component
```typescript
// ObservabilityPage.tsx (Container)
export default function ObservabilityPage() {
  const { vmId } = useRouteParams(); // Get from route
  const { activeTab } = useObservability(vmId);
  
  return (
    <DashboardLayout>
      <ObservabilityHeader vmId={vmId} />
      <TabNav activeTab={activeTab} />
      <TabContent activeTab={activeTab} vmId={vmId} />
    </DashboardLayout>
  );
}

// New: components/observability/ObservabilityHeader.tsx
// New: components/observability/TabContent.tsx
```

**Impact:** 🟡 Medium  
**Effort:** 🟡 Medium (2 hours)

---

#### OCP Violation: VM Actions
**Location:** `components/vm/VMActions.tsx`

**Current:** Hardcoded action buttons
```typescript
export function VMActions({ vm, onStart, onStop, onDelete }) {
  return (
    <div>
      {vm.status === "stopped" && <button onClick={onStart}>Start</button>}
      {vm.status === "running" && (
        <>
          <button onClick={onStop}>Stop</button>
          <button onClick={onDelete}>Delete</button>
        </>
      )}
    </div>
  );
}
```

**Problem:** Adding new actions requires modifying the component

**Solution:** Action configuration pattern
```typescript
interface VMAction {
  id: string;
  label: string;
  icon: LucideIcon;
  variant: 'default' | 'destructive' | 'outline';
  isVisible: (vm: VM) => boolean;
  isEnabled: (vm: VM) => boolean;
  onClick: (vm: VM) => void;
}

const defaultActions: VMAction[] = [
  {
    id: 'start',
    label: 'Start',
    icon: Play,
    variant: 'default',
    isVisible: (vm) => vm.status === 'stopped',
    isEnabled: (vm) => true,
    onClick: (vm) => onStart(vm.id),
  },
  // ...
];

export function VMActions({ vm, actions = defaultActions }) {
  return actions
    .filter(action => action.isVisible(vm))
    .map(action => <ActionButton key={action.id} action={action} vm={vm} />);
}
```

**Impact:** 🟢 Low (future-proofing)  
**Effort:** 🟡 Medium (3 hours)

---

### 3. **YAGNI Violations**

#### Over-engineering: VMsPage Filters
**Location:** `pages/VMsPage.tsx`

**Current:**
```typescript
const [sortField, setSortField] = useState("created_at");
const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
```

**Problem:** Sort functionality exists but is NOT USED anywhere in the UI
**Evidence:** Removed unused variables, but sorting logic still in comments

**Solution:** Remove entirely until needed
```typescript
// Keep only status filter - actually used
const [statusFilter, setStatusFilter] = useState<string>("all");
```

**Impact:** 🟢 Low  
**Effort:** 🟢 Low (30 min)

---

#### Unused Store Features
**Location:** `stores/uiStore.ts`

**Current:**
```typescript
vmStatusFilter, vmSortField, vmSortOrder, vmCurrentPage
setVMStatusFilter, setVMSortField, setVMSortOrder, setVMCurrentPage, resetVMFilters
```

**Problem:** Store exists but is NOT CONNECTED to any component
**Evidence:** VMsPage uses local state instead

**Solution:** Either:
1. **Remove store** (if not needed) - YAGNI
2. **Connect to VMsPage** - Replace local state with store

**Recommendation:** Remove for now, add when pagination/filtering is implemented

**Impact:** 🟢 Low  
**Effort:** 🟢 Low (30 min)

---

### 4. **Clean Code Issues**

#### Magic Numbers
**Location:** Multiple files

```typescript
// useVMs.ts
refetchInterval: 5000,  // What is this?

// useObservability.ts  
refetchInterval: 30000,  // Why 30s?
refetchInterval: activeTab === "logs" ? 5000 : false,  // Magic number

// AdminHealthPage.tsx
refetchInterval: 30000, // Refresh every 30s
```

**Solution:** Use constants
```typescript
// src/lib/constants.ts
export const REFRESH_INTERVALS = {
  VM_STATUS: 5000,      // 5 seconds
  METRICS: 30000,       // 30 seconds
  LOGS_LIVE: 5000,      // 5 seconds when active
  HEALTH: 30000,        // 30 seconds
  AUDIT_LOG: 60000,     // 1 minute
} as const;
```

**Impact:** 🟡 Medium (readability)  
**Effort:** 🟢 Low (1 hour)

---

#### Inconsistent Error Messages
**Location:** All hooks

```typescript
// useVMs.ts
toast.error(`Failed to start VM: ${error.response?.data?.message || "Unknown error"}`);

// useAdminUsers.ts
toast.error(`Failed to update role: ${error.response?.data?.message || "Unknown error"}`);

// Pattern repeated 15+ times
```

**Solution:** Create error handler utility
```typescript
// src/lib/errorHandler.ts
export function getErrorMessage(error: any, defaultAction: string): string {
  return `${defaultAction}: ${error.response?.data?.message || "Unknown error"}`;
}

// Usage in hooks
onError: (error: any) => {
  toast.error(getErrorMessage(error, "Failed to start VM"));
};
```

**Impact:** 🟢 Low (consistency)  
**Effort:** 🟢 Low (1 hour)

---

#### Component Size Violations

| Component | Lines | Target | Status |
|-----------|-------|--------|--------|
| VMsPage | 191 | <150 | 🔴 |
| ObservabilityPage | 153 | <150 | 🔴 |
| CreateVMWizard | ~200 | <150 | 🔴 |
| AdminUsersPage | 71 | <150 | ✅ |
| DashboardPage | 60 | <150 | ✅ |
| VMDetailPage | 76 | <150 | ✅ |

**Solution:** Extract sub-components (see SOLID section)

---

## Refactor Roadmap

### Phase 1: Quick Wins (2-3 hours) 🟢

**Goal:** Maximum impact, minimum effort

1. **Create utility functions** (1h)
   - `getVMStatusStyles()` - Eliminate duplicate status logic
   - `formatDate()` - Consistent date formatting
   - `getErrorMessage()` - Unified error handling

2. **Remove dead code** (30m)
   - Delete unused `uiStore.ts` or document intended use
   - Remove sorting code from VMsPage
   - Remove unused constants exports

3. **Add constants** (30m)
   - Move magic numbers to `constants.ts`
   - Document why each interval was chosen

**Deliverables:**
- ✅ `src/lib/vm-utils.ts`
- ✅ Updated `src/lib/utils.ts`
- ✅ Updated `src/lib/constants.ts`
- ✅ Cleaner codebase

---

### Phase 2: Component Extraction (3-4 hours) 🟡

**Goal:** Improve maintainability

1. **Refactor ObservabilityPage** (2h)
   - Extract `ObservabilityHeader`
   - Extract `TabContent` switch logic
   - Connect vmId to route params

2. **Refactor VMsPage** (1-2h)
   - Extract filters to `VMFilters` component
   - Extract pagination to `Pagination` component
   - Consider using `VMTable` properly

**Deliverables:**
- ✅ `components/observability/ObservabilityHeader.tsx`
- ✅ `components/observability/TabContent.tsx`
- ✅ `components/vm/VMFilters.tsx`
- ✅ `components/ui/Pagination.tsx`

---

### Phase 3: Architecture Improvements (4-5 hours) 🟠

**Goal:** Future-proof codebase

1. **VM Actions configuration** (3h)
   - Implement action configuration pattern
   - Make actions extensible
   - Add action permissions (optional)

2. **Error boundary improvements** (1h)
   - Add error boundaries to routes
   - Better error recovery UI

3. **Loading states standardization** (1h)
   - Create `LoadingState` component
   - Replace all animate-pulse patterns

**Deliverables:**
- ✅ Configurable VM actions
- ✅ Route-level error boundaries
- ✅ Consistent loading states

---

## Success Metrics

| Metric | Before | Target | Measurement |
|--------|--------|--------|-------------|
| Duplicate status logic | 4 instances | 1 utility | Code search |
| Component size >150 lines | 3 components | 0 components | LOC count |
| Magic numbers | 10+ instances | 0 instances | Code review |
| DRY violations | High | Low | Subjective |
| TypeScript errors | 38 | 0 | `tsc --noEmit` |
| Build errors | 1 (vite config) | 0 | `npm run build` |

---

## Files to Create

### Utilities
```
src/lib/vm-utils.ts         - VM-specific utilities
src/lib/errorHandler.ts     - Error handling utilities
```

### Components
```
src/components/ui/Pagination.tsx              - Reusable pagination
src/components/ui/LoadingState.tsx            - Standardized loading
src/components/observability/ObservabilityHeader.tsx
src/components/observability/TabContent.tsx
src/components/vm/VMFilters.tsx
src/components/vm/VMStatusBadge.tsx           - Status badge component
```

---

## Files to Modify

### High Priority
- `src/pages/VMsPage.tsx` - Remove dead code, extract components
- `src/pages/ObservabilityPage.tsx` - Extract sub-components
- `src/lib/constants.ts` - Add refresh intervals
- `src/lib/utils.ts` - Add formatDate

### Medium Priority
- `src/components/vm/VMActions.tsx` - Configurable actions
- `src/components/vm/VMHeader.tsx` - Use VMStatusBadge
- `src/components/vm/VMTable.tsx` - Use VMStatusBadge
- All hooks - Use errorHandler

---

## Recommendations

### Do Now (Phase 1)
- ✅ Fix all TypeScript errors (DONE)
- ✅ Create utility functions
- ✅ Remove dead code
- ✅ Add constants

### Do Soon (Phase 2)
- Extract large components
- Improve ObservabilityPage routing

### Do Later (Phase 3)
- Configurable VM actions
- Error boundaries
- Performance optimization

### Not Recommended (YAGNI)
- Complex state management (Zustand is enough)
- Over-engineered action systems
- Premature optimization

---

## Estimated Effort

| Phase | Time | Impact | Priority |
|-------|------|--------|----------|
| Phase 1: Quick Wins | 2-3h | High | 🔴 Critical |
| Phase 2: Component Extraction | 3-4h | Medium | 🟡 High |
| Phase 3: Architecture | 4-5h | Medium | 🟢 Medium |
| **Total** | **9-12h** | **High** | |

---

## Next Steps

1. **Review this plan** with team
2. **Prioritize phases** based on timeline
3. **Create GitHub issues** for each task
4. **Start with Phase 1** (quick wins)
5. **Measure progress** against success metrics

---

**Prepared by:** EZ Agents Quick Task  
**Review Status:** Ready for implementation  
**Last Updated:** 2026-03-31

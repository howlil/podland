# Quick Task: Phase 2 - Component Extraction

**Date:** 2026-03-31
**Mode:** EZ Agents Quick Task
**Task:** Execute Phase 2 of Frontend Refactor Plan - Component Extraction

---

## Objective

Extract large components to improve maintainability and reduce file sizes below 150 lines target.

---

## Current State Analysis

### Components Exceeding 150 Lines Target

| Component | Lines | Target | Priority |
|-----------|-------|--------|----------|
| VMsPage | 191 | <150 | 🔴 High |
| ObservabilityPage | 153 | <150 | 🔴 High |
| CreateVMWizard | ~200 | <150 | 🟡 Medium |

### Existing Components (Already Created)

**Observability:**
- ✅ `TabNav.tsx` - Tab navigation
- ✅ `MetricsDashboard.tsx` - Metrics display
- ✅ `LogViewer.tsx` - Log viewer

**VM:**
- ✅ `VMHeader.tsx` - VM header with status
- ✅ `VMActions.tsx` - Action buttons
- ✅ `VMTable.tsx` - VM list table
- ✅ `ResourceMetrics.tsx` - Resource cards
- ✅ `ConnectionInfo.tsx` - Connection details

**Dashboard:**
- ✅ `StatsGrid.tsx` - Stats cards
- ✅ `QuotaUsage.tsx` - Quota bars
- ✅ `RecentVMs.tsx` - Recent VMs list

**UI:**
- ✅ `StatsCard.tsx`, `badge.tsx`, `button.tsx`, `Input.tsx`, `Select.tsx`
- ✅ `EmptyState.tsx`, `alert.tsx`, `progress.tsx`, `card.tsx`, `skeleton.tsx`

---

## Tasks

### Task 1: Extract ObservabilityHeader Component
**File:** `src/components/observability/ObservabilityHeader.tsx`
**Source:** `ObservabilityPage.tsx` lines 40-68 (header section)
**Effort:** 🟢 Low (30 min)

**Props:**
```typescript
interface ObservabilityHeaderProps {
  vmId: string;
  onOpenGrafana: () => void;
}
```

---

### Task 2: Extract TabContent Component
**File:** `src/components/observability/TabContent.tsx`
**Source:** `ObservabilityPage.tsx` tab content switch logic
**Effort:** 🟢 Low (30 min)

**Props:**
```typescript
interface TabContentProps {
  activeTab: string;
  metrics: any;
  alerts: any;
  isLoading: boolean;
  timeRange: string;
  vmId: string;
}
```

---

### Task 3: Extract VMFilters Component
**File:** `src/components/vm/VMFilters.tsx`
**Source:** `VMsPage.tsx` filter section
**Effort:** 🟡 Medium (1 hour)

**Props:**
```typescript
interface VMFiltersProps {
  statusFilter: string;
  onStatusFilterChange: (status: string) => void;
  onCreateVM: () => void;
}
```

---

### Task 4: Extract Pagination Component
**File:** `src/components/ui/Pagination.tsx`
**Source:** `VMsPage.tsx` pagination logic
**Effort:** 🟡 Medium (1 hour)

**Props:**
```typescript
interface PaginationProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  totalItems?: number;
  itemsPerPage?: number;
}
```

---

### Task 5: Refactor ObservabilityPage
**File:** `src/pages/ObservabilityPage.tsx`
**Target:** <100 lines
**Effort:** 🟢 Low (30 min)

**New structure:**
```typescript
export default function ObservabilityPage() {
  const [vmId] = useState<string>("");
  const { activeTab, setActiveTab, metrics, timeRange, alerts, isLoading } = useObservability(vmId);

  if (!vmId) {
    return <EmptyState ... />;
  }

  return (
    <DashboardLayout>
      <ObservabilityHeader vmId={vmId} onOpenGrafana={...} />
      <TabNav activeTab={activeTab} onTabChange={setActiveTab} alertsCount={alerts?.length || 0} />
      <TabContent
        activeTab={activeTab}
        metrics={metrics}
        alerts={alerts}
        isLoading={isLoading}
        timeRange={timeRange}
        vmId={vmId}
      />
    </DashboardLayout>
  );
}
```

---

### Task 6: Refactor VMsPage
**File:** `src/pages/VMsPage.tsx`
**Target:** <150 lines
**Effort:** 🟡 Medium (1 hour)

**New structure:**
```typescript
export default function VMsPage() {
  const [statusFilter, setStatusFilter] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const { data: vms = [], isLoading } = useQuery(...);

  // Filtered VMs logic
  const filteredVMs = useMemo(() => ..., [vms, statusFilter]);
  
  // Pagination logic
  const ITEMS_PER_PAGE = 10;
  const totalPages = Math.ceil(filteredVMs.length / ITEMS_PER_PAGE);
  const paginatedVMs = filteredVMs.slice(...);

  return (
    <DashboardLayout>
      <VMFilters
        statusFilter={statusFilter}
        onStatusFilterChange={setStatusFilter}
        onCreateVM={() => setIsWizardOpen(true)}
      />
      <VMTable vms={paginatedVMs} isLoading={isLoading} />
      <Pagination
        currentPage={currentPage}
        totalPages={totalPages}
        onPageChange={setCurrentPage}
        totalItems={filteredVMs.length}
        itemsPerPage={ITEMS_PER_PAGE}
      />
      <CreateVMWizard isOpen={isWizardOpen} onClose={() => setIsWizardOpen(false)} />
    </DashboardLayout>
  );
}
```

---

## Success Criteria

| Metric | Before | Target | Status |
|--------|--------|--------|--------|
| ObservabilityPage lines | 153 | <100 | ⏳ Pending |
| VMsPage lines | 191 | <150 | ⏳ Pending |
| Reusable components created | 0 | 4 | ⏳ Pending |
| TypeScript compilation | ✅ Pass | ✅ Pass | ⏳ Pending |

---

## Files to Create

1. `src/components/observability/ObservabilityHeader.tsx`
2. `src/components/observability/TabContent.tsx`
3. `src/components/vm/VMFilters.tsx`
4. `src/components/ui/Pagination.tsx`

---

## Files to Modify

1. `src/pages/ObservabilityPage.tsx` - Extract components
2. `src/pages/VMsPage.tsx` - Extract components

---

## Execution Plan

1. Create `ObservabilityHeader.tsx`
2. Create `TabContent.tsx`
3. Refactor `ObservabilityPage.tsx`
4. Create `VMFilters.tsx`
5. Create `Pagination.tsx`
6. Refactor `VMsPage.tsx`
7. Run TypeScript check
8. Update STATE.md
9. Commit changes

---

**Estimated Time:** 3-4 hours
**Priority:** High
**Status:** Ready to execute

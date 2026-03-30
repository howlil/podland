# Frontend Deep Refactor Plan

**Date:** 2026-03-30  
**Goal:** Complete frontend architecture refactor with proper separation of concerns, UI/UX excellence, and product engineering best practices

---

## 📊 Current State Analysis

### ✅ What's Done (Architecture Foundation)

```
apps/frontend/src/
├── hooks/              ✅ Business logic layer
│   └── useVMs.ts       ✅ VM data fetching + mutations
├── stores/             ✅ Client state layer
│   └── uiStore.ts      ✅ UI filters, pagination (Zustand)
├── components/         ✅ Reusable presentational components
│   ├── vm/
│   │   └── VMTable.tsx ✅ Presentational VM list
│   └── ui/
│       └── StatsCard.tsx ✅ Stats, QuotaBar, EmptyState
├── pages/              ⚠️ PARTIAL - VMsPage done, others pending
│   └── VMsPage.tsx     ✅ Using hooks + components
└── routes/             ✅ Route config only (thin)
```

### ❌ What Needs Refactoring

| Page | Current Issues | Priority |
|------|---------------|----------|
| `DashboardPage.tsx` | Direct API calls, mixed logic/UI, no reusable components | P0 |
| `VMDetailPage.tsx` | Inline mutations, no custom hook usage | P0 |
| `ObservabilityPage.tsx` | Monolithic component, no separation | P1 |
| `AdminPage.tsx` | Duplicate logic with other pages | P1 |
| `AdminUsersPage.tsx` | Direct API calls, no reusable table | P1 |
| `AdminHealthPage.tsx` | Hardcoded metrics, no components | P2 |
| `AdminAuditLogPage.tsx` | No pagination component | P2 |

---

## 🎯 Refactor Goals

### 1. Architecture Excellence

- [ ] **100% Container/Presentational Pattern**
  - Pages = Container (orchestration only)
  - Components = Presentational (pure UI)
  - Hooks = Business logic (reusable)

- [ ] **Complete State Management Strategy**
  - Server State: TanStack Query (all data fetching)
  - Client State: Zustand (UI preferences, filters)
  - Route State: TanStack Router (search params)

- [ ] **Zero Direct API Calls in Pages**
  - All API calls through custom hooks
  - Consistent error handling
  - Unified loading states

### 2. UI/UX Excellence

- [ ] **Design System Implementation**
  - Consistent color palette (Tailwind config)
  - Reusable component library (shadcn/ui pattern)
  - Typography scale (h1-h6, body, caption)
  - Spacing system (4px grid)

- [ ] **Accessibility (WCAG 2.1 AA)**
  - ARIA labels on all interactive elements
  - Keyboard navigation (tab index, focus states)
  - Screen reader announcements
  - Color contrast compliance

- [ ] **Responsive Design**
  - Mobile-first approach
  - Breakpoints: sm (640px), md (768px), lg (1024px), xl (1280px)
  - Touch targets: min 44x44px
  - Safe areas for notched devices

- [ ] **Loading States**
  - Skeleton screens for all data fetching
  - Optimistic UI updates
  - Progressive enhancement

- [ ] **Error States**
  - User-friendly error messages
  - Retry mechanisms
  - Fallback UIs

- [ ] **Empty States**
  - Illustrations for key pages
  - Clear CTAs
  - Onboarding hints

### 3. Product Engineering

- [ ] **User Onboarding Flow**
  - First-time user guide
  - Feature tooltips
  - Progress indicators

- [ ] **Analytics Integration**
  - Page view tracking
  - Action tracking (VM create, start, stop, delete)
  - Error tracking

- [ ] **Performance Optimization**
  - Code splitting per route
  - Lazy loading for heavy components
  - Image optimization
  - Bundle size < 500KB

- [ ] **SEO (if applicable)**
  - Meta tags
  - Open Graph
  - Structured data

---

## 📋 Refactor Roadmap

### Phase 2A: Core Pages (P0) - **ESTIMATED: 8 hours**

#### Task 2A.1: Refactor DashboardPage (2h)

**Current Issues:**
- Direct API calls in component
- Inline stats calculation
- Mixed concerns (logic + UI)

**Target Structure:**
```tsx
// hooks/useDashboard.ts
export function useDashboard() {
  const { data: vms } = useVMs();
  const { data: quota } = useQuota();
  
  const stats = useMemo(() => ({
    totalVMs: vms.length,
    runningVMs: vms.filter(vm => vm.status === 'running').length,
    totalCPU: vms.reduce((sum, vm) => sum + vm.cpu, 0),
    totalRAM: vms.reduce((sum, vm) => sum + vm.ram, 0),
  }), [vms]);
  
  return { stats, quota, vms };
}

// pages/DashboardPage.tsx
export default function DashboardPage() {
  const { stats, quota, vms } = useDashboard();
  
  return (
    <DashboardLayout>
      <StatsGrid stats={stats} />
      <QuotaUsage quota={quota} />
      <RecentVMs vms={vms.slice(0, 5)} />
    </DashboardLayout>
  );
}
```

**Components to Create:**
- `StatsGrid.tsx` - Grid layout for stats
- `QuotaUsage.tsx` - Quota progress bars
- `RecentVMs.tsx` - Recent VM list

---

#### Task 2A.2: Refactor VMDetailPage (2h)

**Current Issues:**
- Inline mutations
- No custom hook usage
- Repetitive code

**Target Structure:**
```tsx
// hooks/useVM.ts (already exists, extend it)
export function useVM(id: string) {
  const { data: vm, ...query } = useQuery({
    queryKey: ['vm', id],
    queryFn: () => api.get(`/vms/${id}`),
  });
  
  const mutations = useVMMutations(id); // start, stop, delete, pin
  
  return { vm, ...query, ...mutations };
}

// pages/VMDetailPage.tsx
export default function VMDetailPage() {
  const { id } = useParams({ from: '/dashboard/vms/$id' });
  const { vm, startVM, stopVM, deleteVM, isLoading } = useVM(id);
  
  if (isLoading) return <SkeletonVMDetail />;
  if (!vm) return <NotFound />;
  
  return (
    <DashboardLayout>
      <VMHeader vm={vm} />
      <ResourceMetrics vmId={id} />
      <ConnectionInfo vm={vm} />
      <VMActions vm={vm} onStart={startVM} onStop={stopVM} onDelete={deleteVM} />
    </DashboardLayout>
  );
}
```

**Components to Create:**
- `SkeletonVMDetail.tsx` - Loading skeleton
- `VMHeader.tsx` - VM name, status, pin
- `ResourceMetrics.tsx` - CPU, RAM, Storage usage
- `ConnectionInfo.tsx` - Domain, SSH access
- `VMActions.tsx` - Action buttons

---

#### Task 2A.3: Refactor ObservabilityPage (2h)

**Current Issues:**
- Monolithic component
- No tab state management
- Direct API calls

**Target Structure:**
```tsx
// hooks/useObservability.ts
export function useObservability(vmId: string) {
  const [activeTab, setActiveTab] = useState<'metrics' | 'logs' | 'alerts'>('metrics');
  const [timeRange, setTimeRange] = useState('24h');
  
  const { data: metrics } = useQuery({ queryKey: ['metrics', vmId, timeRange], ... });
  const { data: logs } = useQuery({ queryKey: ['logs', vmId], ... });
  const { data: alerts } = useQuery({ queryKey: ['alerts', vmId], ... });
  
  return { activeTab, setActiveTab, timeRange, setTimeRange, metrics, logs, alerts };
}

// pages/ObservabilityPage.tsx
export default function ObservabilityPage() {
  const { vmId } = useParams({ from: '/dashboard/vms/$id' });
  const { activeTab, setActiveTab, metrics, logs, alerts } = useObservability(vmId);
  
  return (
    <DashboardLayout>
      <PageHeader title="Observability" vmId={vmId} />
      <TabNav activeTab={activeTab} onTabChange={setActiveTab} />
      {activeTab === 'metrics' && <MetricsDashboard metrics={metrics} />}
      {activeTab === 'logs' && <LogsViewer logs={logs} />}
      {activeTab === 'alerts' && <AlertsList alerts={alerts} />}
    </DashboardLayout>
  );
}
```

**Components to Create:**
- `PageHeader.tsx` - Page title, breadcrumbs
- `TabNav.tsx` - Tab navigation
- `MetricsDashboard.tsx` - Metrics charts/cards
- `LogsViewer.tsx` - Log viewer with filters
- `AlertsList.tsx` - Alert history

---

#### Task 2A.4: Refactor Admin Pages (2h)

**AdminPage.tsx:**
```tsx
// pages/AdminPage.tsx
export default function AdminPage() {
  const { data: stats } = useAdminStats();
  
  return (
    <DashboardLayout>
      <AdminNav />
      <StatsGrid stats={stats} />
      <QuickActions />
    </DashboardLayout>
  );
}
```

**AdminUsersPage.tsx:**
```tsx
// hooks/useAdminUsers.ts
export function useAdminUsers() {
  const [roleFilter, setRoleFilter] = useState('all');
  const { data: users, ...query } = useQuery({
    queryKey: ['admin-users', roleFilter],
    queryFn: () => api.get(`/admin/users?role=${roleFilter}`),
  });
  
  const mutations = useAdminUserMutations(); // ban, unban, changeRole
  
  return { users, roleFilter, setRoleFilter, ...query, ...mutations };
}

// pages/AdminUsersPage.tsx
export default function AdminUsersPage() {
  const { users, roleFilter, setRoleFilter, banUser, unbanUser } = useAdminUsers();
  
  return (
    <DashboardLayout>
      <PageHeader title="User Management" />
      <FilterBar roleFilter={roleFilter} onFilterChange={setRoleFilter} />
      <UserTable users={users} onBan={banUser} onUnban={unbanUser} />
      <Pagination />
    </DashboardLayout>
  );
}
```

**Components to Create:**
- `AdminNav.tsx` - Admin navigation
- `FilterBar.tsx` - Reusable filter bar
- `UserTable.tsx` - User list with actions
- `Pagination.tsx` - Reusable pagination

---

### Phase 2B: UI/UX Polish (P1) - **ESTIMATED: 6 hours**

#### Task 2B.1: Design System Setup (2h)

**Files to Create/Update:**
```ts
// tailwind.config.ts
export default defineConfig({
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
        },
        // ... semantic colors
      },
      typography: {
        DEFAULT: {
          css: {
            h1: { fontSize: '2.25rem', fontWeight: '700' },
            h2: { fontSize: '1.875rem', fontWeight: '600' },
            // ...
          },
        },
      },
    },
  },
});
```

**Components to Create:**
- `components/ui/Button.tsx` - All button variants
- `components/ui/Input.tsx` - Form inputs
- `components/ui/Select.tsx` - Dropdown selects
- `components/ui/Card.tsx` - Card containers
- `components/ui/Badge.tsx` - Status badges
- `components/ui/Skeleton.tsx` - Loading skeletons
- `components/ui/EmptyState.tsx` - Empty state illustrations

---

#### Task 2B.2: Accessibility Improvements (2h)

**Checklist:**
- [ ] Add ARIA labels to all buttons
- [ ] Implement keyboard navigation for all interactive elements
- [ ] Add focus states for all focusable elements
- [ ] Ensure color contrast meets WCAG AA (4.5:1 for text)
- [ ] Add screen reader announcements for dynamic content
- [ ] Implement skip links for keyboard users
- [ ] Add landmark regions (main, nav, aside)

**Example:**
```tsx
// Before
<button onClick={handleStart}>Start VM</button>

// After
<button
  onClick={handleStart}
  aria-label={`Start VM: ${vm.name}`}
  className="focus:ring-2 focus:ring-primary focus:ring-offset-2"
>
  <PlayIcon aria-hidden="true" />
  <span>Start VM</span>
</button>
```

---

#### Task 2B.3: Loading & Error States (2h)

**Skeleton Components:**
```tsx
// components/ui/Skeleton/VMListSkeleton.tsx
export function VMListSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="animate-pulse flex items-center gap-4 p-4 bg-gray-100 rounded-xl">
          <Skeleton className="h-10 w-10 rounded-lg" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-3 w-24" />
          </div>
          <Skeleton className="h-8 w-24" />
        </div>
      ))}
    </div>
  );
}
```

**Error Boundaries:**
```tsx
// lib/ErrorBoundary.tsx (already exists, enhance it)
export class ErrorBoundary extends Component {
  // Add retry logic
  // Add error reporting
  // Add user-friendly messages
}
```

---

### Phase 2C: Product Features (P2) - **ESTIMATED: 4 hours**

#### Task 2C.1: User Onboarding (2h)

**Features:**
- First-time user welcome modal
- Feature tooltips (using `react-tooltip`)
- Progress checklist
- Quick start guide

**Components:**
```tsx
// components/onboarding/WelcomeModal.tsx
// components/onboarding/FeatureTooltip.tsx
// components/onboarding/ProgressChecklist.tsx
```

---

#### Task 2C.2: Analytics Integration (1h)

**Implementation:**
```ts
// lib/analytics.ts
export function trackPageView(path: string) {
  // Send to analytics service
}

export function trackEvent(event: string, properties?: Record<string, any>) {
  // Track user actions
}

// Usage in pages
useEffect(() => {
  trackPageView(location.pathname);
}, []);

const handleCreateVM = () => {
  trackEvent('vm_create', { tier, os });
  // ...
};
```

---

#### Task 2C.3: Performance Optimization (1h)

**Actions:**
- [ ] Enable code splitting in Vite config
- [ ] Lazy load heavy components (charts, logs viewer)
- [ ] Optimize images (WebP, lazy loading)
- [ ] Tree-shake unused dependencies
- [ ] Analyze bundle with `rollup-plugin-visualizer`

**Vite Config:**
```ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          router: ['@tanstack/react-router'],
          query: ['@tanstack/react-query'],
        },
      },
    },
  },
});
```

---

## 📊 Success Metrics

| Metric | Before | Target | How to Measure |
|--------|--------|--------|----------------|
| **Code Quality** | | | |
| Pages with direct API calls | 7 | 0 | Grep for `api.get` in pages/ |
| Reusable components | 3 | 15+ | Count in components/ui/ |
| Custom hooks | 1 | 8+ | Count in hooks/ |
| **UI/UX** | | | |
| Lighthouse Accessibility | 85 | 95+ | Lighthouse CI |
| Lighthouse Performance | 75 | 90+ | Lighthouse CI |
| Bundle size | ~600KB | <500KB | `npm run build` output |
| **Performance** | | | |
| First Contentful Paint | ~2s | <1s | Web Vitals |
| Time to Interactive | ~4s | <2.5s | Web Vitals |
| **Developer Experience** | | | |
| Component reusability | Low | High | Code reuse audit |
| Test coverage | 0% | 60%+ | Vitest coverage report |

---

## 🛠️ Tools & Libraries

### Existing (Keep)
- ✅ TanStack Router
- ✅ TanStack Query
- ✅ Zustand
- ✅ Tailwind CSS v4
- ✅ Lucide React
- ✅ Sonner (toasts)

### New (Add)
- [ ] `@tanstack/react-table` - For advanced tables (AdminUsersPage)
- [ ] `recharts` or `victory` - For metrics charts
- [ ] `react-tooltip` - For onboarding tooltips
- [ ] `@vitejs/plugin-react-swc` - Faster builds
- [ ] `vitest` + `@testing-library/react` - For testing

---

## 📅 Timeline

| Phase | Tasks | Estimated Time | Dependencies |
|-------|-------|----------------|--------------|
| **2A: Core Pages** | 2A.1-2A.4 | 8 hours | None |
| **2B: UI/UX Polish** | 2B.1-2B.3 | 6 hours | Phase 2A |
| **2C: Product Features** | 2C.1-2C.3 | 4 hours | Phase 2B |
| **Testing** | Write tests | 4 hours | All phases |
| **Total** | | **22 hours** | |

---

## 🚀 Execution Order

1. **Start with Phase 2A** (Core Pages) - Highest impact
2. **Then Phase 2B** (UI/UX Polish) - Improves user experience
3. **Finally Phase 2C** (Product Features) - Nice-to-have
4. **End with Testing** - Ensure quality

---

## ✅ Definition of Done

A page is considered **fully refactored** when:

- [ ] No direct API calls (all through hooks)
- [ ] Uses reusable components from `components/`
- [ ] UI state in Zustand (if shared) or local state
- [ ] Proper loading states (skeletons)
- [ ] Proper error states (error boundary + retry)
- [ ] Proper empty states (illustrations + CTA)
- [ ] ARIA labels on all interactive elements
- [ ] Keyboard navigation works
- [ ] Responsive on all breakpoints
- [ ] Touch targets ≥ 44x44px
- [ ] Color contrast meets WCAG AA
- [ ] Analytics tracking implemented

---

**Ready to execute? Start with Task 2A.1: Refactor DashboardPage** 🚀

# Phase 4 Context: Observability

**Phase:** 4 of 5
**Goal:** Users can monitor VM metrics, view logs, and receive alerts
**Requirements:** 6 (MON-01 through MON-02, MON-03 through MON-04, MON-05 through MON-06)
**Status:** Context gathered — all decisions locked, ready for planning

---

## Prior Context (From Phase 1, 2, 3)

### Architecture Decisions (Locked)

| Decision | Value | Rationale |
|----------|-------|-----------|
| Orchestration | k3s | Cloud native, production-ready, 500MB RAM footprint |
| Backend | Go 1.25+ | Excellent k3s ecosystem, type-safe, performant |
| Frontend | React + TanStack Router + Tailwind v4 | Modern DX, type-safe routing |
| Database | PostgreSQL 15 | Battle-tested, JSONB flexibility |
| VM Abstraction | Docker containers with resource limits | Shared resource model, fast startup |
| Ingress | Traefik | Already configured for wildcard subdomain routing |

### Existing Infrastructure (Phase 1-3)

```
podland/
├── apps/
│   ├── backend/          # Go service (OAuth, JWT, sessions, VM CRUD, DNS automation)
│   └── frontend/         # React + TanStack Router (dashboard, VM management, domains)
├── infra/
│   ├── k3s/
│   │   ├── namespace.yaml
│   │   ├── postgres.yaml
│   │   ├── backend.yaml
│   │   ├── frontend.yaml
│   │   ├── traefik-config.yaml
│   │   ├── traefik-https.yaml
│   │   ├── cloudflared.yaml      # Has metrics endpoint on :2000
│   │   └── origin-ca-secret.yaml
│   └── docker-compose/
└── packages/
    └── types/
```

### Reusable Patterns

- **Backend:** Clean architecture (handler → usecase → repository)
- **Frontend:** TanStack Router with file-based routing
- **k3s:** Namespace per user, Deployment per VM, PVC for storage
- **Monitoring:** Cloudflared already exposes metrics on port 2000

---

## Phase 4 Decisions (Implementation Details)

All decisions below are locked. Planning agent uses these to create actionable tasks.

### 1. Metrics Collection Architecture

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Metrics Source** | cAdvisor via kubelet | Built into k3s, provides CPU/memory/disk/network at `/stats/summary`, zero additional deployment |
| **Scraping Frequency** | 30 seconds | Prometheus default, balanced load, sufficient for alerting |
| **Storage Strategy** | 15-day retention in Prometheus TSDB | Standard, simple config, sufficient for debugging window |
| **Metrics Granularity** | Per-VM (pod-level) | Matches user mental model, precise quota tracking, manageable series count |

**Implementation Pattern:**
```yaml
# Prometheus scrape config for cAdvisor
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cadvisor
  namespace: monitoring
spec:
  selector:
    matchLabels:
      k8s-app: cadvisor
  endpoints:
  - port: https-metrics
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
    interval: 30s
    relabelings:
    # Map pod labels to vm_id
    - sourceLabels: [__meta_kubernetes_pod_label_vm_id]
      targetLabel: vm_id
```

**PromQL Queries:**
```promql
# CPU usage per VM
avg_over_time(container_cpu_usage_seconds_total{vm_id="abc"}[5m])

# Memory usage per VM
container_memory_usage_bytes{vm_id="abc"}

# Network receive bytes
rate(container_network_receive_bytes_total{vm_id="abc"}[5m])

# Disk I/O
rate(container_fs_reads_bytes_total{vm_id="abc"}[5m])
```

**Storage Math:**
- 100 VMs × 10 metrics × 30s interval × 15 days = ~38M samples
- At ~80 bytes/sample = **~3GB storage** (trivial on modern SSD)

---

### 2. Log Aggregation Strategy

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Log Collector** | Promtail (DaemonSet) | Native Loki agent, lightweight (10-20MB RAM), auto-discovery, LogQL support |
| **Log Format** | Plain text with Kubernetes labels | Zero app changes required, Loki labels handle filtering |
| **Retention Policy** | 30 days | Standard debugging window, ~20GB storage for 500 users |
| **Query UI** | Hybrid (Pre-built + "Advanced" toggle) | 90% use simple buttons, 10% power users get LogQL editor |

**Implementation Pattern:**
```yaml
# Promtail DaemonSet
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: promtail
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: promtail
  template:
    metadata:
      labels:
        app: promtail
    spec:
      serviceAccountName: promtail
      containers:
      - name: promtail
        image: grafana/promtail:latest
        args:
        - -config.file=/etc/promtail/config.yml
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
        - name: promtail-config
          mountPath: /etc/promtail
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: promtail-config
        configMap:
          name: promtail-config
```

**Promtail Config:**
```yaml
# config.yml
server:
  http_listen_port: 9080
positions:
  filename: /tmp/positions.yaml
clients:
  - url: http://loki.monitoring.svc:3100/loki/api/v1/push
scrape_configs:
  - job_name: kubernetes-pods
    kubernetes_sd_configs:
      - role: pod
    pipeline_stages:
      - kubernetes_sd:
          pod: true
      - labels:
          vm_id:
          level:
```

**LogQL Examples:**
```logql
# Last 100 lines for VM
{vm_id="abc"} | limit 100

# Errors only
{vm_id="abc"} |= "error"

# Time range
{vm_id="abc"} |~ "2026-03-29"
```

---

### 3. Alerting System Design

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Alert Engine** | Prometheus Alertmanager | Industry standard, Grafana integration, routing/silencing/inhibition included |
| **Threshold Configuration** | Hard-coded system defaults | CPU > 90%, RAM > 85% for all VMs — simple, no UI complexity |
| **Notification Channels** | In-app notifications only | Database table + bell icon, no SMTP/webhook infra required |
| **Alert Evaluation** | Sustained (5-minute average) | Filters transient spikes, industry standard, simple PromQL |

**Implementation Pattern:**
```yaml
# Prometheus alert rules
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: vm-alerts
  namespace: monitoring
spec:
  groups:
  - name: vm-alerts
    rules:
    - alert: VMHighCPU
      expr: avg_over_time(container_cpu_usage_seconds_total{vm_id!=""}[5m]) > 0.9
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "VM {{ $labels.vm_id }} has high CPU usage"
        description: "CPU usage is above 90% for 5 minutes"
    - alert: VMHighMemory
      expr: (container_memory_usage_bytes{vm_id!=""} / container_spec_memory_limit_bytes{vm_id!=""}) > 0.85
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "VM {{ $labels.vm_id }} has high memory usage"
        description: "Memory usage is above 85% for 5 minutes"
```

**Alertmanager Config:**
```yaml
# alertmanager.yml
global:
  resolve_timeout: 5m
route:
  group_by: ['vm_id', 'alertname']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'in-app-webhook'
receivers:
  - name: 'in-app-webhook'
    webhook_configs:
      - url: 'http://podland-backend.podland.svc:8080/api/alerts/webhook'
        send_resolved: true
```

**Database Schema:**
```sql
CREATE TABLE notifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  vm_id UUID REFERENCES vms(id),
  alert_name VARCHAR(100) NOT NULL,
  severity VARCHAR(20) NOT NULL,
  title VARCHAR(255) NOT NULL,
  message TEXT NOT NULL,
  is_read BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMP
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
```

---

### 4. Dashboard UX & Visualization

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Grafana Integration** | Hybrid (Custom summary + Grafana for detail) | Dashboard shows 3 simple charts, "View Full Metrics" opens Grafana — no iframe auth debt |
| **Time Range Selection** | Hybrid (Pre-set + "Custom" button) | 4 buttons cover 90% use cases, "Custom" opens Grafana for date picker |
| **Log Viewer Interaction** | Hybrid (Static + "Live Tail" toggle) | Default shows last 1000 lines, toggle enables WebSocket streaming |
| **Dashboard Layout** | Hybrid (VM detail has summary, dedicated page for full view) | Quick check on VM page, deep dive on observability page |

**Frontend Component Structure:**
```
apps/frontend/src/
├── routes/
│   └── dashboard/
│       ├── -vms/
│       │   └── $id.tsx              # VM detail with metrics summary
│       ├── observability/
│       │   ├── index.tsx            # Dedicated observability page
│       │   ├── metrics.tsx          # Full metrics view
│       │   └── logs.tsx             # Full log viewer
│       └── notifications.tsx        # Bell icon + notification list
├── components/
│   ├── observability/
│   │   ├── MetricsSummary.tsx       # 3 charts (CPU, RAM, Network)
│   │   ├── MetricsDetail.tsx        # Full Grafana link
│   │   ├── LogViewer.tsx            # Static + Live Tail toggle
│   │   ├── LogQLInput.tsx           # Advanced mode toggle
│   │   └── AlertBadge.tsx           # Alert status indicator
│   └── notifications/
│       ├── NotificationBell.tsx
│       └── NotificationList.tsx
```

**VM Detail Page Integration:**
```tsx
// apps/frontend/src/routes/dashboard/-vms/$id.tsx
function VMDetailPage() {
  const vm = useVM(id);
  
  return (
    <div>
      <VMHeader vm={vm} />
      
      {/* Metrics Summary Section */}
      <section>
        <h2>Resource Usage</h2>
        <MetricsSummary vmId={vm.id} timeRange="24h" />
        <Button variant="outline" onClick={() => navigate('/dashboard/observability?vm=' + vm.id)}>
          View Full Metrics →
        </Button>
      </section>
      
      {/* Recent Logs Section */}
      <section>
        <h2>Recent Logs</h2>
        <LogViewer vmId={vm.id} limit={100} showLiveTail={false} />
      </section>
    </div>
  );
}
```

**Observability Page:**
```tsx
// apps/frontend/src/routes/dashboard/observability/index.tsx
function ObservabilityPage() {
  const vmId = useSearch().vm;
  const [tab, setTab] = useState<'metrics' | 'logs' | 'alerts'>('metrics');
  
  return (
    <DashboardLayout>
      <Tabs value={tab} onChange={setTab}>
        <TabsList>
          <TabsTrigger value="metrics">Metrics</TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
          <TabsTrigger value="alerts">Alerts</TabsTrigger>
        </TabsList>
        
        <TabsContent value="metrics">
          <MetricsDetail vmId={vmId} />
        </TabsContent>
        
        <TabsContent value="logs">
          <LogViewer vmId={vmId} showLiveTail={true} />
        </TabsContent>
        
        <TabsContent value="alerts">
          <AlertHistory vmId={vmId} />
        </TabsContent>
      </Tabs>
    </DashboardLayout>
  );
}
```

---

## Requirements (From ROADMAP.md)

| ID | Requirement | Success Criteria |
|----|-------------|------------------|
| MON-01 | System collects CPU, RAM, disk, and network metrics | Prometheus scrapes cAdvisor every 30s |
| MON-02 | User can view Grafana dashboard with metrics | Grafana deployed, dashboards accessible |
| MON-03 | System aggregates logs from all VMs in Loki | Promtail → Loki pipeline working |
| MON-04 | User can view VM logs in dashboard | Log viewer with 1000 lines + Live Tail |
| MON-05 | System generates alerts when VM CPU > 90% | Alert fires after 5-min sustained usage |
| MON-06 | System generates alerts when VM RAM > 85% | Alert fires after 5-min sustained usage |

---

## Technical Milestones

- [ ] Prometheus Operator deployed
- [ ] ServiceMonitor for cAdvisor metrics
- [ ] Grafana deployed with dashboards
- [ ] Loki + Promtail deployed
- [ ] Log query API endpoint
- [ ] Alertmanager configuration
- [ ] Alert notification system (in-app)
- [ ] Frontend: Metrics summary on VM detail page
- [ ] Frontend: Dedicated observability page
- [ ] Frontend: Log viewer with Live Tail
- [ ] Frontend: Notification bell + list

---

## Infrastructure Components

```
monitoring namespace:
├── prometheus-operator      # Operator for CRDs
├── prometheus               # Prometheus instance (TSDB)
├── alertmanager             # Alert routing
├── grafana                  # Metrics visualization
├── loki                     # Log aggregation
├── promtail (DaemonSet)     # Log collection
└── service monitors         # Scrape configs

podland namespace (backend):
└── alert webhook handler    # Receives alerts, stores in DB
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Prometheus storage growth | 15-day retention, persistent volume with size limit |
| Log volume explosion | 30-day retention, Promtail rate limiting |
| Alert fatigue | 5-min evaluation, 4h repeat interval, grouping |
| Grafana auth complexity | Open in new tab, separate session |
| WebSocket connection drops | Auto-reconnect, fallback to polling |

---

## Deferred Ideas (For Phase 5)

| Idea | Deferred To | Reason |
|------|-------------|--------|
| Email notifications | Phase 5 | Requires SMTP setup, deliverability management |
| Custom alert thresholds | Phase 5 | UI complexity, validation logic |
| Webhook integrations (Discord/Slack) | Phase 5 | User setup required, validation complexity |
| Custom date picker | Phase 5 | Grafana handles this adequately |
| Natural language log search | Phase 5 | NLP/LLM integration overkill |
| Thanos/Cortex for long-term retention | Phase 5 | 15 days sufficient for v1 |

---

## Code Context (What Exists)

### Backend (`apps/backend/`)

**Extend these:**
- `internal/handler/vm_handler.go` — Add metrics/logs endpoints
- `internal/handler/notification_handler.go` — New file for notifications
- `internal/entity/vm.go` — Add metrics fields if needed
- `internal/repository/` — Add notification repository
- `cmd/main.go` — Wire up alert webhook handler

**New files to create:**
- `internal/handler/alert_webhook.go` — Receives Alertmanager webhooks
- `internal/repository/notification_repository.go` — Notification CRUD
- `internal/entity/notification.go` — Notification entity

### Frontend (`apps/frontend/`)

**Extend these:**
- `routes/dashboard/-vms/$id.tsx` — Add metrics summary section
- `routes/dashboard/notifications.tsx` — Notification page

**New files to create:**
- `routes/dashboard/observability/index.tsx` — Observability page
- `components/observability/MetricsSummary.tsx` — 3-chart summary
- `components/observability/LogViewer.tsx` — Log viewer with Live Tail
- `components/notifications/NotificationBell.tsx` — Bell icon

### Infrastructure (`infra/k3s/`)

**New files to create:**
- `monitoring/namespace.yaml`
- `monitoring/prometheus-operator.yaml`
- `monitoring/prometheus.yaml`
- `monitoring/alertmanager.yaml`
- `monitoring/grafana.yaml`
- `monitoring/loki.yaml`
- `monitoring/promtail.yaml`
- `monitoring/service-monitors.yaml`
- `monitoring/alert-rules.yaml`

---

## Next Steps

**For Planning Agent:**
1. Break down 6 requirements into implementation tasks
2. Estimate effort per task (hours)
3. Define acceptance criteria per task
4. Identify dependencies (Prometheus before Grafana, Loki before Promtail)

**Planning Output:** `02-PLAN.md` with week-by-week breakdown

---

*Context created: 2026-03-29*
*All 16 decisions locked — ready for planning*

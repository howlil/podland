# Phase 4 Plan: Observability

**Phase:** 4 of 5
**Goal:** Users can monitor VM metrics, view logs, and receive alerts
**Requirements:** 6 (MON-01 through MON-06)
**Duration:** 4 weeks
**Status:** Ready for execution

---

## Week Breakdown

| Week | Focus | Tasks | Estimated Hours |
|------|-------|-------|-----------------|
| 1 | Infrastructure: Prometheus + Alertmanager | 1.1-1.4 | 20h |
| 2 | Infrastructure: Loki + Grafana | 2.1-2.4 | 18h |
| 3 | Backend: Alert Webhook + APIs | 3.1-3.4 | 16h |
| 4 | Frontend: Dashboards + Polish | 4.1-4.4 | 16h |

**Total:** 70 hours (~2 weeks full-time or 4 weeks part-time)

---

## Week 1: Infrastructure — Prometheus + Alertmanager

### Task 1.1: Prometheus Operator Deployment (4h)

**Objective:** Deploy Prometheus Operator to manage Prometheus instances via CRDs.

**Files to Create:**
```
infra/k3s/monitoring/
├── namespace.yaml
├── prometheus-operator.yaml
└── rbac.yaml
```

**Implementation:**
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
  labels:
    pod-security.kubernetes.io/enforce: baseline
---
# prometheus-operator.yaml (simplified Helm chart or manifest)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus-operator
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus-operator
  template:
    metadata:
      labels:
        app: prometheus-operator
    spec:
      containers:
      - name: prometheus-operator
        image: quay.io/prometheus-operator/prometheus-operator:v0.70.0
        args:
        - --kubelet-service=kube-system/kubelet
        - --prometheus-config-reloader=quay.io/prometheus-operator/prometheus-config-reloader:v0.70.0
        ports:
        - containerPort: 8080
          name: http
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
```

**Acceptance Criteria:**
- [ ] `kubectl get pods -n monitoring` shows prometheus-operator running
- [ ] CRDs available: `kubectl get servicemonitor`, `kubectl get prometheusrule`
- [ ] Operator logs show no errors

**Dependencies:** None

---

### Task 1.2: Prometheus Instance with cAdvisor Scrape (6h)

**Objective:** Deploy Prometheus instance configured to scrape cAdvisor metrics.

**Files to Create:**
```
infra/k3s/monitoring/
├── prometheus.yaml              # Prometheus CR
├── prometheus-pvc.yaml          # Persistent storage
└── cadvisor-service-monitor.yaml
```

**Implementation:**
```yaml
# prometheus.yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: podland
  namespace: monitoring
spec:
  replicas: 1
  version: v3.0.0
  serviceAccountName: prometheus
  serviceMonitorSelector: {}
  serviceMonitorNamespaceSelector: {}
  ruleSelector: {}
  resources:
    requests:
      cpu: 100m
      memory: 512Mi
    limits:
      cpu: 500m
      memory: 2Gi
  storage:
    volumeClaimTemplate:
      spec:
        storageClassName: local-lvm
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi
---
# cadvisor-service-monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cadvisor
  namespace: kube-system
  labels:
    app: cadvisor
spec:
  jobLabel: k8s-app
  selector:
    matchLabels:
      k8s-app: cadvisor
  namespaceSelector:
    matchNames:
    - kube-system
  endpoints:
  - port: https-metrics
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
    interval: 30s
    honorLabels: true
    relabelings:
    # Map pod labels to vm_id
    - sourceLabels: [__meta_kubernetes_pod_label_vm_id]
      targetLabel: vm_id
```

**Backend Changes:**
- Add `vm_id` label to VM deployments (already exists from Phase 2)

**Acceptance Criteria:**
- [ ] Prometheus UI accessible via port-forward
- [ ] Query `container_cpu_usage_seconds_total` returns data
- [ ] Query `container_memory_usage_bytes` returns data
- [ ] Metrics include `vm_id` label for VM pods

**Dependencies:** Task 1.1

---

### Task 1.3: Alertmanager Deployment (4h)

**Objective:** Deploy Alertmanager configured to send alerts to backend webhook.

**Files to Create:**
```
infra/k3s/monitoring/
├── alertmanager.yaml
├── alertmanager-config.yaml
└── alertmanager-service.yaml
```

**Implementation:**
```yaml
# alertmanager.yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: podland
  namespace: monitoring
spec:
  replicas: 1
  version: v0.27.0
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 100m
      memory: 128Mi
---
# alertmanager-config.yaml (Secret)
apiVersion: v1
kind: Secret
metadata:
  name: alertmanager-podland-config
  namespace: monitoring
type: Opaque
stringData:
  alertmanager.yml: |
    global:
      resolve_timeout: 5m
    route:
      group_by: ['vm_id', 'alertname']
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 4h
      receiver: 'backend-webhook'
    receivers:
      - name: 'backend-webhook'
        webhook_configs:
          - url: 'http://podland-backend.podland.svc:8080/api/alerts/webhook'
            send_resolved: true
```

**Acceptance Criteria:**
- [ ] Alertmanager pod running
- [ ] Config loaded without errors
- [ ] Test alert successfully sent to webhook

**Dependencies:** Task 1.1

---

### Task 1.4: Alert Rules Configuration (6h)

**Objective:** Define Prometheus alert rules for CPU and memory thresholds.

**Files to Create:**
```
infra/k3s/monitoring/
└── alert-rules.yaml
```

**Implementation:**
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: vm-alerts
  namespace: monitoring
  labels:
    prometheus: podland
spec:
  groups:
  - name: vm-alerts
    interval: 30s
    rules:
    - alert: VMHighCPU
      expr: |
        avg_over_time(
          container_cpu_usage_seconds_total{vm_id!="", image!=""}[5m]
        ) > 0.9
      for: 5m
      labels:
        severity: warning
        alert_type: cpu
      annotations:
        summary: "VM {{ $labels.vm_id }} has high CPU usage"
        description: "CPU usage is above 90% for 5 minutes (current: {{ $value | printf \"%.0f\" }}%)"
    - alert: VMHighMemory
      expr: |
        (
          container_memory_usage_bytes{vm_id!="", image!=""}
          /
          container_spec_memory_limit_bytes{vm_id!="", image!=""}
        ) > 0.85
      for: 5m
      labels:
        severity: warning
        alert_type: memory
      annotations:
        summary: "VM {{ $labels.vm_id }} has high memory usage"
        description: "Memory usage is above 85% (current: {{ $value | mul 100 | printf \"%.0f\" }}%)"
```

**Acceptance Criteria:**
- [ ] Rules loaded in Prometheus (check `/rules` endpoint)
- [ ] Alert fires when threshold breached (test with load)
- [ ] Alert includes vm_id label for routing
- [ ] Alert resolves when metric returns to normal

**Dependencies:** Task 1.2

---

## Week 2: Infrastructure — Loki + Grafana

### Task 2.1: Loki Deployment (4h)

**Objective:** Deploy Loki for log aggregation with 30-day retention.

**Files to Create:**
```
infra/k3s/monitoring/
├── loki.yaml
├── loki-config.yaml
└── loki-service.yaml
```

**Implementation:**
```yaml
# loki.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: loki
  namespace: monitoring
  labels:
    app: loki
spec:
  replicas: 1
  selector:
    matchLabels:
      app: loki
  template:
    metadata:
      labels:
        app: loki
    spec:
      containers:
      - name: loki
        image: grafana/loki:3.0.0
        args:
        - -config.file=/etc/loki/loki.yaml
        ports:
        - containerPort: 3100
          name: http
        resources:
          requests:
            cpu: 100m
            memory: 512Mi
          limits:
            cpu: 500m
            memory: 2Gi
        volumeMounts:
        - name: loki-config
          mountPath: /etc/loki
        - name: loki-storage
          mountPath: /loki
      volumes:
      - name: loki-config
        configMap:
          name: loki-config
      - name: loki-storage
        persistentVolumeClaim:
          claimName: loki-storage
---
# loki-config.yaml (ConfigMap)
apiVersion: v1
kind: ConfigMap
metadata:
  name: loki-config
  namespace: monitoring
data:
  loki.yaml: |
    auth_enabled: false
    server:
      http_listen_port: 3100
    common:
      path_prefix: /loki
      storage:
        filesystem:
          chunks_directory: /loki/chunks
          rules_directory: /loki/rules
    schema_config:
      configs:
      - from: 2024-01-01
        store: tsdb
        object_store: filesystem
        schema: v13
        index:
          prefix: index_
          period: 24h
    limits_config:
      retention_period: 30d
      enforce_metric_name: false
      reject_old_samples: true
      reject_old_samples_max_age: 168h
    chunk_store_config:
      max_look_back_period: 30d
    table_manager:
      retention_deletes_enabled: true
      retention_period: 30d
```

**Acceptance Criteria:**
- [ ] Loki pod running and healthy
- [ ] Query `http://loki.monitoring.svc:3100/ready` returns 200
- [ ] PVC created with 20Gi storage

**Dependencies:** None

---

### Task 2.2: Promtail DaemonSet (4h)

**Objective:** Deploy Promtail to collect logs from all VMs.

**Files to Create:**
```
infra/k3s/monitoring/
├── promtail.yaml
├── promtail-config.yaml
└── promtail-rbac.yaml
```

**Implementation:**
```yaml
# promtail.yaml (DaemonSet)
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: promtail
  namespace: monitoring
  labels:
    app: promtail
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
        image: grafana/promtail:3.0.0
        args:
        - -config.file=/etc/promtail/promtail.yaml
        ports:
        - containerPort: 9080
          name: http
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 100m
            memory: 128Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: containers
          mountPath: /var/lib/docker/containers
          readOnly: true
        - name: promtail-config
          mountPath: /etc/promtail
        env:
        - name: HOSTNAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: containers
        hostPath:
          path: /var/lib/docker/containers
      - name: promtail-config
        configMap:
          name: promtail-config
```

**Acceptance Criteria:**
- [ ] Promtail running on all nodes (DaemonSet)
- [ ] Logs visible in Loki (query via Grafana)
- [ ] Logs include vm_id label

**Dependencies:** Task 2.1

---

### Task 2.3: Grafana Deployment (6h)

**Objective:** Deploy Grafana with pre-configured dashboards.

**Files to Create:**
```
infra/k3s/monitoring/
├── grafana.yaml
├── grafana-config.yaml
├── grafana-datasources.yaml
├── grafana-dashboards.yaml
└── grafana-service.yaml
```

**Implementation:**
```yaml
# grafana.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: monitoring
  labels:
    app: grafana
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:11.0.0
        ports:
        - containerPort: 3000
          name: http
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 512Mi
        volumeMounts:
        - name: grafana-storage
          mountPath: /var/lib/grafana
        - name: grafana-datasources
          mountPath: /etc/grafana/provisioning/datasources
        - name: grafana-dashboards
          mountPath: /etc/grafana/provisioning/dashboards
      volumes:
      - name: grafana-storage
        persistentVolumeClaim:
          claimName: grafana-storage
      - name: grafana-datasources
        configMap:
          name: grafana-datasources
      - name: grafana-dashboards
        configMap:
          name: grafana-dashboards
```

**Acceptance Criteria:**
- [ ] Grafana accessible via port-forward
- [ ] Prometheus data source configured and working
- [ ] Loki data source configured and working
- [ ] VM Metrics dashboard imported and showing data
- [ ] VM Logs dashboard imported and showing data

**Dependencies:** Tasks 1.2, 2.1

---

### Task 2.4: Grafana Dashboards (4h)

**Objective:** Create pre-built Grafana dashboards for VM metrics and logs.

**Files to Create:**
```
infra/k3s/monitoring/
└── dashboards/
    ├── vm-metrics.json
    └── vm-logs.json
```

**Dashboard Panels:**

**VM Metrics Dashboard:**
1. CPU Usage (time series, 24h default)
2. Memory Usage (time series, 24h default)
3. Network I/O (time series, receive/transmit)
4. Disk I/O (time series, read/write)
5. Resource Quota vs Usage (gauge)

**VM Logs Dashboard:**
1. Log Volume Over Time (time series)
2. Log Level Distribution (pie chart)
3. Recent Logs (table with search)
4. Error Logs (filtered table)

**Acceptance Criteria:**
- [ ] Dashboards accessible in Grafana
- [ ] Variables configured (vm_id, time range)
- [ ] Panels show correct data
- [ ] Dashboards exported as JSON for version control

**Dependencies:** Task 2.3

---

## Week 3: Backend — Alert Webhook + APIs

### Task 3.1: Alert Webhook Handler (4h)

**Objective:** Create endpoint to receive Alertmanager webhooks and store notifications.

**Files to Create:**
```
apps/backend/internal/handler/alert_webhook.go
apps/backend/internal/repository/notification_repository.go
apps/backend/internal/entity/notification.go
```

**Implementation:**
```go
// internal/entity/notification.go
type Notification struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    VMID       uuid.UUID
    AlertName  string
    Severity   string
    Title      string
    Message    string
    IsRead     bool
    CreatedAt  time.Time
    ResolvedAt *time.Time
}

// internal/handler/alert_webhook.go
func (h *alertWebhookHandler) HandleAlert(w http.ResponseWriter, r *http.Request) {
    // Verify internal service token
    token := r.Header.Get("X-Service-Token")
    if token != h.config.ServiceToken {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var payload AlertmanagerPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    // Process each alert
    for _, alert := range payload.Alerts {
        // Extract vm_id from alert labels
        vmID := alert.Labels["vm_id"]
        
        // Get VM owner from database
        vm, err := h.vmRepo.GetVMByID(vmID)
        if err != nil {
            continue
        }

        // Create notification
        notification := &entity.Notification{
            UserID:    vm.UserID,
            VMID:      vm.ID,
            AlertName: alert.Labels["alertname"],
            Severity:  alert.Labels["severity"],
            Title:     alert.Annotations["summary"],
            Message:   alert.Annotations["description"],
        }

        if alert.Status == "resolved" {
            now := time.Now()
            notification.ResolvedAt = &now
        }

        if err := h.notificationRepo.Create(notification); err != nil {
            // Log error, continue
        }
    }

    w.WriteHeader(http.StatusOK)
}
```

**Acceptance Criteria:**
- [ ] Webhook endpoint receives alerts
- [ ] Notifications stored in database
- [ ] Resolved alerts update notification status
- [ ] Invalid tokens rejected

**Dependencies:** Task 1.3

---

### Task 3.2: Metrics API Endpoints (4h)

**Objective:** Create API endpoints for fetching VM metrics.

**Files to Create:**
```
apps/backend/internal/handler/metrics_handler.go
```

**Implementation:**
```go
// internal/handler/metrics_handler.go
type MetricsHandler struct {
    prometheusURL string
    authToken     string
}

func (h *MetricsHandler) GetVMMetrics(w http.ResponseWriter, r *http.Request) {
    vmID := chi.URLParam(r, "id")
    
    // Parse query params
    rangeStr := r.URL.Query().Get("range") // 24h, 7d, 30d
    if rangeStr == "" {
        rangeStr = "24h"
    }

    // Query Prometheus
    cpuQuery := fmt.Sprintf(`avg_over_time(container_cpu_usage_seconds_total{vm_id="%s"}[%s])`, vmID, rangeStr)
    memoryQuery := fmt.Sprintf(`container_memory_usage_bytes{vm_id="%s"}`, vmID)
    
    cpuData, err := h.queryPrometheus(cpuQuery)
    memoryData, err := h.queryPrometheus(memoryQuery)
    
    response := MetricsResponse{
        CPU:    cpuData,
        Memory: memoryData,
    }
    
    pkgresponse.JSON(w, http.StatusOK, response)
}

func (h *MetricsHandler) RedirectToGrafana(w http.ResponseWriter, r *http.Request) {
    vmID := chi.URLParam(r, "id")
    grafanaURL := fmt.Sprintf("http://grafana.monitoring.svc:3000/d/vm-metrics?var-vm_id=%s", vmID)
    http.Redirect(w, r, grafanaURL, http.StatusSeeOther)
}
```

**Acceptance Criteria:**
- [ ] GET /api/vms/:id/metrics returns metrics data
- [ ] GET /api/vms/:id/metrics/detail redirects to Grafana
- [ ] Query params (range, step) work correctly
- [ ] Error handling for Prometheus unavailable

**Dependencies:** Task 1.2

---

### Task 3.3: Logs API Endpoints (4h)

**Objective:** Create API endpoints for fetching VM logs.

**Files to Create:**
```
apps/backend/internal/handler/logs_handler.go
```

**Implementation:**
```go
// internal/handler/logs_handler.go
type LogsHandler struct {
    lokiURL string
}

func (h *LogsHandler) GetVMLogs(w http.ResponseWriter, r *http.Request) {
    vmID := chi.URLParam(r, "id")
    limit := r.URL.Query().Get("limit")
    level := r.URL.Query().Get("level")
    
    // Build LogQL query
    query := fmt.Sprintf(`{vm_id="%s"}`, vmID)
    if level != "" {
        query += fmt.Sprintf(` |= "%s"`, level)
    }
    query += fmt.Sprintf(` | limit %s`, limit)
    
    // Query Loki
    logs, err := h.queryLoki(query)
    
    pkgresponse.JSON(w, http.StatusOK, logs)
}

func (h *LogsHandler) StreamVMLogs(w http.ResponseWriter, r *http.Request) {
    vmID := chi.URLParam(r, "id")
    
    // Upgrade to WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    // Stream logs from Loki
    query := fmt.Sprintf(`{vm_id="%s"}`, vmID)
    stream, err := h.tailLoki(query)
    
    for log := range stream {
        conn.WriteJSON(log)
    }
}
```

**Acceptance Criteria:**
- [ ] GET /api/vms/:id/logs returns last N lines
- [ ] GET /api/vms/:id/logs/stream works via WebSocket
- [ ] Query params (limit, level) filter correctly
- [ ] WebSocket auto-reconnects on disconnect

**Dependencies:** Task 2.2

---

### Task 3.4: Notifications API (4h)

**Objective:** Create API endpoints for managing user notifications.

**Files to Create:**
```
apps/backend/internal/handler/notification_handler.go
apps/backend/internal/repository/notification_repository.go
```

**Implementation:**
```go
// internal/handler/notification_handler.go
func (h *notificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(uuid.UUID)
    
    notifications, err := h.notificationRepo.GetByUserID(userID)
    
    pkgresponse.JSON(w, http.StatusOK, notifications)
}

func (h *notificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
    notificationID := chi.URLParam(r, "id")
    userID := r.Context().Value("user_id").(uuid.UUID)
    
    err := h.notificationRepo.MarkAsRead(notificationID, userID)
    
    pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *notificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(uuid.UUID)
    
    count, err := h.notificationRepo.GetUnreadCount(userID)
    
    pkgresponse.JSON(w, http.StatusOK, map[string]int{"count": count})
}
```

**Acceptance Criteria:**
- [ ] GET /api/notifications returns user's notifications
- [ ] POST /api/notifications/:id/read marks as read
- [ ] GET /api/notifications/unread-count returns count
- [ ] Auth required for all endpoints

**Dependencies:** Task 3.1

---

## Week 4: Frontend — Dashboards + Polish

### Task 4.1: Metrics Summary Component (4h)

**Objective:** Create reusable metrics summary component for VM detail page.

**Files to Create:**
```
apps/frontend/src/components/observability/
├── MetricsSummary.tsx
├── CPUGauge.tsx
├── MemoryGauge.tsx
└── NetworkChart.tsx
```

**Implementation:**
```tsx
// components/observability/MetricsSummary.tsx
export function MetricsSummary({ vmId, timeRange = '24h' }) {
  const { data: metrics } = useQuery({
    queryKey: ['metrics', vmId, timeRange],
    queryFn: () => api.get(`/api/vms/${vmId}/metrics?range=${timeRange}`),
  });

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <CPUGauge value={metrics?.cpu.current} limit={metrics?.cpu.limit} />
      <MemoryGauge value={metrics?.memory.current} limit={metrics?.memory.limit} />
      <NetworkChart data={metrics?.network} />
    </div>
  );
}
```

**Acceptance Criteria:**
- [ ] Shows CPU, Memory, Network in 3 cards
- [ ] Gauges show percentage of quota
- [ ] Responsive layout (mobile → desktop)
- [ ] Loading states and error handling

**Dependencies:** Task 3.2

---

### Task 4.2: VM Detail Page Integration (4h)

**Objective:** Add metrics summary section to existing VM detail page.

**Files to Modify:**
```
apps/frontend/src/routes/dashboard/-vms/$id.tsx
```

**Implementation:**
```tsx
// routes/dashboard/-vms/$id.tsx
import { MetricsSummary } from '~/components/observability/MetricsSummary';
import { LogViewer } from '~/components/observability/LogViewer';

export default function VMDetailPage() {
  const { vm } = useVM();

  return (
    <DashboardLayout>
      <VMHeader vm={vm} />
      
      {/* Metrics Summary Section */}
      <section className="mt-6">
        <div className="flex justify-between items-center mb-4">
          <h2>Resource Usage</h2>
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/dashboard/observability?vm=${vm.id}`)}
          >
            View Full Metrics →
          </Button>
        </div>
        <MetricsSummary vmId={vm.id} timeRange="24h" />
      </section>

      {/* Recent Logs Section */}
      <section className="mt-6">
        <div className="flex justify-between items-center mb-4">
          <h2>Recent Logs</h2>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate(`/dashboard/observability/logs?vm=${vm.id}`)}
          >
            View All Logs →
          </Button>
        </div>
        <LogViewer vmId={vm.id} limit={100} showLiveTail={false} />
      </section>
    </DashboardLayout>
  );
}
```

**Acceptance Criteria:**
- [ ] Metrics summary visible on VM detail page
- [ ] "View Full Metrics" button navigates to observability page
- [ ] Recent logs section shows last 100 lines
- [ ] Layout matches design system

**Dependencies:** Task 4.1

---

### Task 4.3: Observability Page (4h)

**Objective:** Create dedicated observability page with tabs for Metrics/Logs/Alerts.

**Files to Create:**
```
apps/frontend/src/routes/dashboard/observability/
├── index.tsx
├── metrics.tsx
└── logs.tsx
```

**Implementation:**
```tsx
// routes/dashboard/observability/index.tsx
import { Tabs } from '~/components/ui/tabs';

export default function ObservabilityPage() {
  const search = useSearch();
  const vmId = search.vm;

  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1>Observability</h1>
        {vmId && <VMNameBadge vmId={vmId} />}
      </div>

      <Tabs defaultValue="metrics" className="w-full">
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

**Acceptance Criteria:**
- [ ] Three tabs: Metrics, Logs, Alerts
- [ ] vm_id query param filters data
- [ ] Each tab shows full-featured view
- [ ] Navigation from VM detail page works

**Dependencies:** Tasks 4.1, 4.2

---

### Task 4.4: Notification Bell + Polish (4h)

**Objective:** Create notification bell component and final polish.

**Files to Create:**
```
apps/frontend/src/components/notifications/
├── NotificationBell.tsx
└── NotificationList.tsx
apps/frontend/src/routes/dashboard/notifications.tsx
```

**Implementation:**
```tsx
// components/notifications/NotificationBell.tsx
export function NotificationBell() {
  const { data: unreadCount } = useQuery({
    queryKey: ['notifications-unread'],
    queryFn: () => api.get('/api/notifications/unread-count'),
    refetchInterval: 30000, // Poll every 30s
  });

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <BellIcon className="h-5 w-5" />
          {unreadCount?.count > 0 && (
            <span className="absolute -top-1 -right-1 h-4 w-4 bg-red-500 rounded-full text-xs text-white flex items-center justify-center">
              {unreadCount.count}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80">
        <NotificationList />
      </PopoverContent>
    </Popover>
  );
}
```

**Acceptance Criteria:**
- [ ] Bell icon in dashboard header
- [ ] Red badge shows unread count
- [ ] Click opens notification list
- [ ] Mark as read works
- [ ] Auto-refresh every 30s

**Dependencies:** Task 3.4

---

## Success Criteria Verification

| Requirement | Success Criteria | Verification Method |
|-------------|------------------|---------------------|
| MON-01 | System collects CPU, RAM, disk, network metrics | Query Prometheus: `container_cpu_usage_seconds_total`, `container_memory_usage_bytes` |
| MON-02 | User can view Grafana dashboard with metrics | Open Grafana URL, see VM metrics dashboard with data |
| MON-03 | System aggregates logs from all VMs in Loki | Query Loki: `{vm_id="..."}` returns logs |
| MON-04 | User can view VM logs in dashboard | Open VM detail page, see last 100 lines + Live Tail toggle |
| MON-05 | System generates alerts when VM CPU > 90% | Load test VM, alert fires after 5min, notification appears |
| MON-06 | System generates alerts when VM RAM > 85% | Load test VM memory, alert fires after 5min, notification appears |

---

## Testing Checklist

### Infrastructure Tests
```bash
# Prometheus scraping
kubectl port-forward -n monitoring svc/podland 9090:9090
# Query: container_cpu_usage_seconds_total{vm_id!=""}

# Alertmanager
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
# Check: http://localhost:9093/#/alerts

# Loki
kubectl port-forward -n monitoring svc/loki 3100:3100
# Query: http://localhost:3100/loki/api/v1/query_range?query={vm_id="..."}

# Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000
# Login: admin/admin, check dashboards
```

### Backend Tests
```bash
# Alert webhook
curl -X POST http://localhost:8080/api/alerts/webhook \
  -H "X-Service-Token: test-token" \
  -H "Content-Type: application/json" \
  -d '{"alerts": [...]}'

# Metrics API
curl http://localhost:8080/api/vms/<vm-id>/metrics?range=24h

# Logs API
curl http://localhost:8080/api/vms/<vm-id>/logs?limit=100
```

### Frontend Tests
```bash
# Run dev server
npm run dev

# Manual testing checklist
# 1. Open VM detail page → see metrics summary
# 2. Click "View Full Metrics" → Grafana opens
# 3. Open Observability page → tabs work
# 4. Click notification bell → list appears
# 5. Mark notification as read → badge disappears
```

---

## Deployment Checklist

### Prerequisites
- [ ] k3s cluster running
- [ ] Backend deployed and healthy
- [ ] Frontend deployed and healthy
- [ ] PostgreSQL database running

### Infrastructure Deployment
```bash
# 1. Deploy monitoring namespace
kubectl apply -f infra/k3s/monitoring/namespace.yaml

# 2. Deploy Prometheus Operator
kubectl apply -f infra/k3s/monitoring/prometheus-operator.yaml

# 3. Deploy Prometheus
kubectl apply -f infra/k3s/monitoring/prometheus.yaml
kubectl apply -f infra/k3s/monitoring/cadvisor-service-monitor.yaml

# 4. Deploy Alertmanager
kubectl apply -f infra/k3s/monitoring/alertmanager.yaml
kubectl apply -f infra/k3s/monitoring/alert-rules.yaml

# 5. Deploy Loki
kubectl apply -f infra/k3s/monitoring/loki.yaml

# 6. Deploy Promtail
kubectl apply -f infra/k3s/monitoring/promtail.yaml

# 7. Deploy Grafana
kubectl apply -f infra/k3s/monitoring/grafana.yaml
kubectl apply -f infra/k3s/monitoring/grafana-datasources.yaml
kubectl apply -f infra/k3s/monitoring/grafana-dashboards.yaml
```

### Backend Deployment
```bash
# Set environment variables
export PROMETHEUS_URL=http://prometheus.monitoring.svc:9090
export LOKI_URL=http://loki.monitoring.svc:3100
export ALERTMANAGER_WEBHOOK_SECRET=<generate-secret>

# Deploy backend with new config
kubectl apply -f infra/k3s/backend.yaml
```

### Frontend Deployment
```bash
# Build and deploy
npm run build
kubectl apply -f infra/k3s/frontend.yaml
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Prometheus storage growth | 10Gi PVC limit, 15-day retention |
| Log volume explosion | 30-day retention, Promtail rate limiting |
| Alert fatigue | 5-min evaluation, 4h repeat interval |
| Grafana auth complexity | Open in new tab, no auth integration |
| WebSocket drops | Auto-reconnect, polling fallback |

---

*Plan created: 2026-03-29*
*Ready for execution — follow week-by-week*

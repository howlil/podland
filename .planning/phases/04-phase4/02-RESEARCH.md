# Phase 4 Research: Observability Stack

**Phase:** 4 of 5
**Status:** Research complete — ready for planning

---

## Research Questions Answered

### 1. k3s Metrics Collection

**Question:** What's the best way to collect container metrics in k3s?

**Answer:** Use **cAdvisor** (built into k3s) scraped by Prometheus.

**Findings:**
- cAdvisor is **already running** in k3s at `/stats/summary` endpoint
- Provides: CPU, memory, disk I/O, network I/O at container level
- Metrics Server only provides CPU/memory — not enough for Phase 4
- Node Exporter adds host-level metrics but duplicates cAdvisor effort

**Implementation Pattern:**
```bash
# Access cAdvisor metrics directly
kubectl port-forward -n kube-system daemonset/kube-proxy 10250:10250
curl -k https://localhost:10250/stats/summary
```

**Prometheus Scrape Config:**
```yaml
scrape_configs:
  - job_name: 'cadvisor'
    scheme: https
    tls_config:
      insecure_skip_verify: true
    kubernetes_sd_configs:
      - role: node
    relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
    metrics_path: /metrics/cadvisor
```

**Source:** [k3s metrics docs](https://docs.k3s.io/reference/metrics), [cAdvisor guide](https://cast.ai/blog/cadvisor/)

---

### 2. Log Collection: Promtail vs Fluent Bit

**Question:** Which log collector is best for Loki on k3s?

**Answer:** **Promtail** — native Loki agent, lightweight, simple.

**Comparison:**

| Feature | Promtail | Fluent Bit |
|---------|----------|------------|
| RAM Usage | 10-20MB | 30-50MB |
| Loki Integration | Native (Grafana Labs) | 3rd party |
| Kubernetes Discovery | Auto-discovery + labels | Requires config |
| LogQL Support | Native | Translation layer |
| Setup Complexity | Simple (DaemonSet) | Medium |

**Why Promtail for Podland:**
- Single-server k3s = single Promtail instance = ~20MB RAM
- Native LogQL support — no translation layer
- Auto-discovers pods, adds Kubernetes labels automatically
- Grafana Labs maintains both Loki + Promtail — guaranteed compatibility

**Note:** The "Fluent Bit + Promtail" pattern is for enterprise EKS with node rotation. Overkill for single-server k3s.

**Source:** [Promtail vs Fluent Bit comparison](https://yoo.be/zero-downtime-log-aggregation-eks-fluent-bit-promtail/)

---

### 3. Alerting: Alertmanager vs Custom

**Question:** Should we use Prometheus Alertmanager or build custom alerting?

**Answer:** **Prometheus Alertmanager** — industry standard, free features.

**Why Alertmanager:**
- **Built-in features:**
  - Alert grouping (don't spam 100 alerts for 1 VM)
  - Silencing (maintenance windows)
  - Inhibition (if VM down, don't alert on CPU)
  - Routing (different channels for different alerts)
- **Native Prometheus integration** — alert rules in same YAML
- **Battle-tested** — handles edge cases you won't think of
- **YAML config is fine** — 50 lines per alert rule, version-controlled

**Custom Go Service Would Require:**
- Polling Prometheus API every 30s
- Deduplication logic
- Rate limiting
- Routing logic
- Silencing logic
- = ~2,000 LOC of alerting logic

**Alertmanager Config (50 lines):**
```yaml
global:
  resolve_timeout: 5m
route:
  group_by: ['vm_id', 'alertname']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'webhook'
receivers:
  - name: 'webhook'
    webhook_configs:
      - url: 'http://backend:8080/api/alerts/webhook'
        send_resolved: true
```

**Source:** [Alertmanager best practices](https://blog.elest.io/prometheus-alertmanager-build-production-ready-monitoring-with-custom-alert-rules/)

---

### 4. Grafana Integration: iframe vs Custom

**Question:** How to show Grafana dashboards without auth complexity?

**Answer:** **Hybrid approach** — custom summary charts + "View Full Metrics" opens Grafana in new tab.

**Options Analyzed:**

| Approach | Pros | Cons | LOC Estimate |
|----------|------|------|--------------|
| Embed iframe | Full Grafana features | Auth proxy, CORS, session sync | ~1,000 LOC glue code |
| Custom charts (Recharts) | Perfect UX match | Rebuild all Grafana features | ~2,000 LOC charts |
| Grafana HTTP API | Query data, custom render | Complex integration, auth | ~1,500 LOC |
| **Hybrid (open in new tab)** | Simple, no auth debt | Two tabs | ~200 LOC |

**Hybrid Pattern:**
```tsx
// Dashboard summary (custom charts)
<MetricsSummary vmId={vm.id} />

// Link to full Grafana dashboard
<Button
  onClick={() => window.open(
    'http://grafana.podland.app/d/vm-metrics?var-vm_id=' + vm.id,
    '_blank'
  )}
>
  View Full Metrics →
</Button>
```

**Why This Works:**
- Students get quick glance (custom charts) + deep dive (Grafana)
- No auth proxy, no CORS, no session sync
- Grafana loads independently with its own session
- Matches Vercel/Netlify pattern (summary → full analytics)

**Source:** [Grafana embedding discussion](https://community.grafana.com/t/embed-grafana-without-iframe-in-a-react-plugin/27095/)

---

## Component Versions (2026)

| Component | Version | Notes |
|-----------|---------|-------|
| Prometheus | 3.0+ | Latest stable, native Helm chart |
| Alertmanager | 0.27+ | Compatible with Prometheus 3.0 |
| Grafana | 11.0+ | Latest LTS, improved performance |
| Loki | 3.0+ | Compatible with Promtail 3.0 |
| Promtail | 3.0+ | Native Loki agent |
| Prometheus Operator | 0.70+ | CRD-based management |

---

## Resource Requirements

### Monitoring Stack (Single Server)

| Component | CPU Request | CPU Limit | RAM Request | RAM Limit | Storage |
|-----------|-------------|-----------|-------------|-----------|---------|
| Prometheus | 100m | 500m | 512Mi | 2Gi | 10Gi (15-day retention) |
| Alertmanager | 50m | 100m | 64Mi | 128Mi | 1Gi |
| Grafana | 100m | 200m | 128Mi | 512Mi | 5Gi (dashboards) |
| Loki | 100m | 500m | 512Mi | 2Gi | 20Gi (30-day logs) |
| Promtail (per node) | 50m | 100m | 64Mi | 128Mi | None |
| **Total** | **400m** | **1.4** | **1.3Gi** | **5Gi** | **36Gi** |

**For Podland (500 users, 1 node):**
- Total monitoring overhead: ~500MB RAM average
- Storage: 36Gi (SSD recommended for Prometheus TSDB)
- CPU: Negligible (idle most of the time)

---

## Deployment Order

1. **Prometheus Operator** — Provides CRDs (ServiceMonitor, PrometheusRule)
2. **Prometheus** — TSDB instance with cAdvisor scrape config
3. **Alertmanager** — Webhook receiver for backend
4. **Loki** — Log aggregation storage
5. **Promtail** — Log collection DaemonSet
6. **Grafana** — Dashboards + data sources
7. **Backend alert webhook** — Receives alerts, stores in DB
8. **Frontend components** — Metrics summary, log viewer, notifications

---

## API Design

### Metrics Endpoints (Backend)

```
GET /api/vms/:id/metrics          # Get VM metrics summary
  Query: ?range=24h&step=1h
  Response: { cpu: [], memory: [], network: [] }

GET /api/vms/:id/metrics/detail   # Redirect to Grafana
  Response: 302 → Grafana URL
```

### Logs Endpoints (Backend)

```
GET /api/vms/:id/logs             # Get last N lines
  Query: ?limit=100&level=error
  Response: { lines: [{ ts, msg, level }] }

GET /api/vms/:id/logs/stream      # WebSocket for live tail
  WS messages: { ts, msg, level }
```

### Notifications Endpoints (Backend)

```
POST /api/alerts/webhook          # Alertmanager webhook
  Body: { alerts: [{ labels, annotations, status }] }
  Auth: Internal service token

GET /api/notifications            # List user's notifications
  Auth: JWT required

POST /api/notifications/:id/read  # Mark as read
  Auth: JWT required
```

---

## Database Schema

```sql
-- Notifications table
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
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

-- Alert history (optional, for audit)
CREATE TABLE alert_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  vm_id UUID REFERENCES vms(id),
  alert_name VARCHAR(100) NOT NULL,
  fired_at TIMESTAMP NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMP,
  severity VARCHAR(20) NOT NULL,
  metric_value DECIMAL(10,2) NOT NULL,
  threshold DECIMAL(10,2) NOT NULL
);

CREATE INDEX idx_alert_history_vm_id ON alert_history(vm_id);
CREATE INDEX idx_alert_history_fired_at ON alert_history(fired_at);
```

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Prometheus storage growth | Medium | Medium | 15-day retention, PV size limit |
| Log volume explosion | Low | Medium | 30-day retention, Promtail rate limiting |
| Alert fatigue | Medium | Low | 5-min evaluation, 4h repeat interval |
| Grafana auth complexity | Low | High | Open in new tab (no auth integration) |
| WebSocket connection drops | Medium | Low | Auto-reconnect, fallback to polling |
| cAdvisor performance impact | Low | Low | 30s scrape interval, not 15s |

---

## Next Steps

**Ready for planning:** All research questions answered.

**Planning should produce:**
1. Week-by-week task breakdown
2. YAML manifests for each component
3. Backend endpoint specifications
4. Frontend component specifications
5. Testing checklist

---

*Research completed: 2026-03-29*
*All technical questions answered — ready for planning*

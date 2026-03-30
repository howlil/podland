# Phase 4: Observability - Execution Summary

**Phase:** 4 of 5
**Date Completed:** 2026-03-29
**Status:** ✅ COMPLETE

---

## Executive Summary

Phase 4 (Observability) has been successfully completed end-to-end. All 6 requirements (MON-01 through MON-06) have been implemented with 16 tasks across 4 waves of execution.

### Key Achievements

- **Metrics Collection:** Prometheus scrapes cAdvisor every 30s with 15-day retention
- **Log Aggregation:** Promtail → Loki pipeline with 30-day retention
- **Alerting:** CPU (>90%) and Memory (>85%) alerts with 5-minute evaluation
- **Dashboards:** Grafana with VM Metrics and VM Logs dashboards
- **APIs:** Backend endpoints for metrics, logs, and notifications
- **UI Components:** MetricsSummary, LogViewer, NotificationBell, Observability page

---

## Requirements Verification

| ID | Requirement | Success Criteria | Status |
|----|-------------|------------------|--------|
| MON-01 | CPU, RAM, disk, network metrics | Query Prometheus: `container_cpu_usage_seconds_total` | ✅ |
| MON-02 | Grafana dashboard | VM Metrics dashboard with 4 panels | ✅ |
| MON-03 | Log aggregation in Loki | Query Loki: `{vm_id="..."}` returns logs | ✅ |
| MON-04 | Log viewer UI | Last 1000 lines + Live Tail toggle | ✅ |
| MON-05 | CPU alerts (>90%) | Alert fires after 5min sustained | ✅ |
| MON-06 | Memory alerts (>85%) | Alert fires after 5min sustained | ✅ |

---

## Files Created

### Infrastructure (21 files)

| File | Purpose |
|------|---------|
| `infra/k3s/monitoring/namespace.yaml` | Monitoring namespace |
| `infra/k3s/monitoring/prometheus-operator.yaml` | Operator deployment |
| `infra/k3s/monitoring/rbac.yaml` | Operator RBAC |
| `infra/k3s/monitoring/prometheus.yaml` | Prometheus CR |
| `infra/k3s/monitoring/prometheus-pvc.yaml` | 10Gi metrics storage |
| `infra/k3s/monitoring/prometheus-rbac.yaml` | Prometheus RBAC |
| `infra/k3s/monitoring/cadvisor-service-monitor.yaml` | cAdvisor scrape (30s) |
| `infra/k3s/monitoring/alertmanager.yaml` | Alertmanager CR |
| `infra/k3s/monitoring/alertmanager-rbac.yaml` | Alertmanager RBAC |
| `infra/k3s/monitoring/alertmanager-config.yaml` | Webhook configuration |
| `infra/k3s/monitoring/alert-rules.yaml` | CPU/Memory alert rules |
| `infra/k3s/monitoring/loki.yaml` | Loki deployment |
| `infra/k3s/monitoring/loki-pvc.yaml` | 20Gi log storage |
| `infra/k3s/monitoring/loki-config.yaml` | 30-day retention config |
| `infra/k3s/monitoring/promtail.yaml` | Log collector DaemonSet |
| `infra/k3s/monitoring/promtail-config.yaml` | Pod log scraping |
| `infra/k3s/monitoring/promtail-rbac.yaml` | Promtail RBAC |
| `infra/k3s/monitoring/grafana.yaml` | Grafana deployment |
| `infra/k3s/monitoring/grafana-pvc.yaml` | 5Gi dashboard storage |
| `infra/k3s/monitoring/grafana-datasources.yaml` | Prometheus + Loki |
| `infra/k3s/monitoring/grafana-dashboards-files.yaml` | Dashboard JSON |

### Backend (6 files)

| File | Purpose |
|------|---------|
| `apps/backend/cmd/main.go` | Updated with observability routes |
| `apps/backend/internal/handler/alert_webhook.go` | Alertmanager webhook |
| `apps/backend/internal/handler/metrics_handler.go` | Metrics API |
| `apps/backend/internal/handler/logs_handler.go` | Logs API + WebSocket |
| `apps/backend/internal/handler/notification_handler.go` | Notifications API |
| `apps/backend/internal/entity/notification.go` | Notification entity |
| `apps/backend/internal/repository/notification_repository.go` | Notification repo |
| `apps/backend/migrations/004_phase4_notifications.sql` | DB migration |

### Frontend (5 files)

| File | Purpose |
|------|---------|
| `apps/frontend/src/components/observability/MetricsSummary.tsx` | Metrics cards |
| `apps/frontend/src/components/observability/LogViewer.tsx` | Log viewer |
| `apps/frontend/src/components/notifications/NotificationBell.tsx` | Notification bell |
| `apps/frontend/src/routes/dashboard/observability/index.tsx` | Observability page |
| `apps/frontend/src/routes/dashboard/-vms/$id.tsx` | Updated VM detail |
| `apps/frontend/src/components/layout/DashboardLayout.tsx` | Updated with bell |

---

## Architecture Decisions

### Metrics Architecture
- **Source:** cAdvisor via kubelet
- **Scrape Interval:** 30 seconds
- **Retention:** 15 days
- **Granularity:** Per-VM (via `vm_id` label)
- **Query API:** Prometheus HTTP API

### Logs Architecture
- **Collector:** Promtail DaemonSet
- **Storage:** Loki with filesystem backend
- **Retention:** 30 days
- **Format:** Plain text with labels
- **UI:** Hybrid (embedded viewer + Grafana link)

### Alerts Architecture
- **Engine:** Prometheus Alertmanager
- **Thresholds:** Hard-coded (90% CPU, 85% Memory)
- **Evaluation:** 5-minute sustained
- **Notification:** In-app only (no email/SMS)
- **Routing:** Webhook to backend → database

### Dashboards Strategy
- **Summary:** Custom React components (MetricsSummary)
- **Detail:** Grafana (open in new tab, no auth integration debt)
- **Logs:** Embedded viewer with Live Tail option

---

## Deployment Instructions

### Prerequisites
- k3s cluster running
- Backend deployed and healthy
- Frontend deployed and healthy
- PostgreSQL database running

### Deploy Monitoring Stack

```bash
# 1. Deploy namespace
kubectl apply -f infra/k3s/monitoring/namespace.yaml

# 2. Deploy Prometheus Operator
kubectl apply -f infra/k3s/monitoring/prometheus-operator.yaml
kubectl apply -f infra/k3s/monitoring/rbac.yaml

# 3. Deploy Prometheus
kubectl apply -f infra/k3s/monitoring/prometheus-pvc.yaml
kubectl apply -f infra/k3s/monitoring/prometheus-rbac.yaml
kubectl apply -f infra/k3s/monitoring/prometheus.yaml
kubectl apply -f infra/k3s/monitoring/cadvisor-service-monitor.yaml

# 4. Deploy Alertmanager
kubectl apply -f infra/k3s/monitoring/alertmanager-rbac.yaml
kubectl apply -f infra/k3s/monitoring/alertmanager.yaml
kubectl apply -f infra/k3s/monitoring/alertmanager-config.yaml
kubectl apply -f infra/k3s/monitoring/alert-rules.yaml

# 5. Deploy Loki
kubectl apply -f infra/k3s/monitoring/loki-pvc.yaml
kubectl apply -f infra/k3s/monitoring/loki-config.yaml
kubectl apply -f infra/k3s/monitoring/loki.yaml

# 6. Deploy Promtail
kubectl apply -f infra/k3s/monitoring/promtail-rbac.yaml
kubectl apply -f infra/k3s/monitoring/promtail-config.yaml
kubectl apply -f infra/k3s/monitoring/promtail.yaml

# 7. Deploy Grafana
kubectl apply -f infra/k3s/monitoring/grafana-pvc.yaml
kubectl apply -f infra/k3s/monitoring/grafana-datasources.yaml
kubectl apply -f infra/k3s/monitoring/grafana-dashboards-config.yaml
kubectl apply -f infra/k3s/monitoring/grafana-dashboards-files.yaml
kubectl apply -f infra/k3s/monitoring/grafana.yaml
```

### Verify Deployment

```bash
# Check all pods
kubectl get pods -n monitoring

# Expected output:
# NAME                                    READY   STATUS
# prometheus-operator-xxxxx               1/1     Running
# prometheus-podland-0                    1/1     Running
# alertmanager-podland-0                  1/1     Running
# loki-xxxxx                              1/1     Running
# promtail-xxxxx (on each node)           1/1     Running
# grafana-xxxxx                           1/1     Running
```

### Access Dashboards

```bash
# Prometheus UI
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Open: http://localhost:9090

# Alertmanager UI
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
# Open: http://localhost:9093

# Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000
# Open: http://localhost:3000
# Login: admin / admin
```

### Configure Backend

Set environment variables for backend:

```bash
export PROMETHEUS_URL=http://prometheus.monitoring.svc:9090
export LOKI_URL=http://loki.monitoring.svc:3100
export GRAFANA_URL=http://grafana.monitoring.svc:3000
export ALERTMANAGER_WEBHOOK_SECRET=<generate-secure-secret>
```

Update Alertmanager config with the secret:

```yaml
# In alertmanager-config.yaml
receivers:
  - name: 'backend-webhook'
    webhook_configs:
      - url: 'http://podland-backend.podland.svc:8080/api/alerts/webhook'
        send_resolved: true
        http_config:
          headers:
            X-Service-Token: <your-secret-token>
```

---

## Testing Checklist

### Metrics Testing

```bash
# 1. Query Prometheus for cAdvisor metrics
kubectl port-forward -n monitoring svc/prometheus 9090:9090

# Open http://localhost:9090 and run queries:
container_cpu_usage_seconds_total{vm_id!=""}
container_memory_usage_bytes{vm_id!=""}
rate(container_network_receive_bytes_total{vm_id!=""}[5m])
```

### Logs Testing

```bash
# 1. Query Loki for VM logs
kubectl port-forward -n monitoring svc/loki 3100:3100

# Query Loki API:
curl "http://localhost:3100/loki/api/v1/query_range?query={vm_id=\"<your-vm-id>\"}&limit=100"
```

### Alerts Testing

```bash
# 1. Generate CPU load on a VM
# (SSH into VM and run: stress --cpu 2 --timeout 600)

# 2. Check Alertmanager after 5 minutes
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
# Open http://localhost:9093/#/alerts

# 3. Check backend notifications table
psql -U podland -d podland -c "SELECT * FROM notifications ORDER BY created_at DESC LIMIT 5;"
```

### API Testing

```bash
# Metrics API
curl http://localhost:8080/api/vms/<vm-id>/metrics?range=24h

# Logs API
curl http://localhost:8080/api/vms/<vm-id>/logs?limit=100

# Notifications API (requires auth cookie)
curl http://localhost:8080/api/notifications
curl http://localhost:8080/api/notifications/unread-count
```

---

## Known Limitations & Technical Debt

1. **Grafana Authentication:** Grafana is open without auth integration. Users see a separate login.
   - **Mitigation:** Open in new tab, document credentials
   - **Future:** Consider OAuth proxy or Grafana enterprise

2. **WebSocket Reconnection:** Log streaming may drop connections
   - **Mitigation:** Auto-reconnect logic in LogViewer component
   - **Future:** Implement exponential backoff

3. **Alert Routing:** All alerts go to all users (no per-VM filtering in Alertmanager)
   - **Mitigation:** Backend filters by VM owner
   - **Future:** Dynamic Alertmanager config per tenant

4. **Storage Limits:** Fixed PVC sizes (10Gi metrics, 20Gi logs)
   - **Mitigation:** Retention policies (15d metrics, 30d logs)
   - **Future:** Monitor usage, add alerting on storage

---

## Next Steps (Phase 5)

1. **Testing:** Add integration tests for observability components
2. **Documentation:** User guide for monitoring features
3. **Polish:** Improve error handling, loading states
4. **Performance:** Optimize Prometheus queries, add caching
5. **Security:** Add RBAC for metrics/logs access

---

## Lessons Learned

### What Went Well
- Wave-based execution allowed parallel work
- Following the plan reduced decision fatigue
- Reusing existing patterns (handler/repository/entity) sped up development

### What Could Be Better
- Grafana dashboard JSON is verbose - consider Helm charts next time
- WebSocket log streaming needs more robust error handling
- Alert testing requires manual load generation - could automate

---

*Summary created: 2026-03-29*
*Phase 4 execution complete - ready for Phase 5 (Polish & Testing)*

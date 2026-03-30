# Launch Checklist: Podland v1.0

**Phase:** 5 of 5 - Admin + Polish
**Target Launch Date:** [DATE]
**Status:** In Progress

---

## Pre-Launch (T-7 days)

### Infrastructure

- [ ] k3s cluster healthy
  ```bash
  kubectl get nodes
  kubectl get pods --all-namespaces
  ```
- [ ] PostgreSQL backup running (verify latest backup)
  ```bash
  kubectl get jobs --namespace podland
  kubectl logs job/pg-backup -n podland
  ```
- [ ] Monitoring stack healthy (Prometheus, Loki, Grafana)
  ```bash
  kubectl get pods -n monitoring
  ```
- [ ] Cloudflare Tunnel active
  ```bash
  kubectl get pods -n cloudflare
  ```

### Backend

- [ ] All migrations applied
  ```bash
  cd apps/backend
  # Verify migration 005_phase5_admin.sql is applied
  ```
- [ ] Environment variables set
  - [ ] `SENDGRID_API_KEY` configured
  - [ ] `SENDGRID_FROM_EMAIL` configured
  - [ ] `PROMETHEUS_URL` configured
- [ ] Admin superadmin user created
  ```sql
  UPDATE users SET role = 'superadmin' WHERE email = 'admin@podland.app';
  ```

### Frontend

- [ ] Build successful
  ```bash
  cd apps/frontend
  npm run build
  ```
- [ ] All routes accessible
  - [ ] `/admin` - Admin dashboard
  - [ ] `/admin/users` - User management
  - [ ] `/admin/health` - System health
  - [ ] `/admin/audit-log` - Audit log
- [ ] Mobile responsive tested

### Security

- [ ] Security audit passed (no critical/high vulnerabilities)
  ```bash
  cd apps/backend
  go mod verify
  
  cd apps/frontend
  npm audit
  ```
- [ ] Dependencies updated
- [ ] CSRF protection enabled
- [ ] Rate limiting configured

---

## Launch Day (T-0)

### Final Checks

- [ ] Load test passed (100 concurrent users)
  ```bash
  k6 run tests/load/critical-paths.js
  ```
- [ ] Database backup taken (manual pre-launch backup)
  ```bash
  ./scripts/backup-db.sh
  ```
- [ ] Monitoring alerts configured
- [ ] On-call rotation established

### Deployment

- [ ] Deploy backend
  ```bash
  kubectl apply -f infra/k3s/backend.yaml
  kubectl rollout restart deployment/backend -n podland
  ```
- [ ] Deploy frontend
  ```bash
  kubectl apply -f infra/k3s/frontend.yaml
  kubectl rollout restart deployment/frontend -n podland
  ```
- [ ] Verify health endpoints
  - [ ] `GET /api/health` returns 200
  - [ ] `GET /api/admin/health` returns metrics
- [ ] Test critical user flows
  - [ ] Login flow
  - [ ] VM creation
  - [ ] VM start/stop
  - [ ] Metrics viewing

### Post-Deployment Verification

- [ ] Login flow works
- [ ] VM creation works
- [ ] Metrics visible
- [ ] Admin panel accessible
- [ ] Email notifications sent (test idle warning)
- [ ] Pin VM feature works
- [ ] Audit logging works

---

## Post-Launch (T+7 days)

### Monitoring

- [ ] Review error logs daily
  ```bash
  kubectl logs -l app=backend -n podland --since=24h
  ```
- [ ] Monitor resource usage
  - [ ] CPU < 80%
  - [ ] Memory < 80%
  - [ ] Storage < 80%
- [ ] Check idle detection working
  ```bash
  kubectl logs -l app=backend -n podland | grep "idle VM detection"
  ```
- [ ] Verify backup restoration (dry run)
  ```bash
  gunzip -c backup.sql.gz | psql -h localhost -U podland -d podland_test
  ```

### User Feedback

- [ ] Collect user feedback
- [ ] Document issues encountered
- [ ] Plan Phase 5.1 improvements

---

## Rollback Plan

If issues occur during launch:

1. **Immediate Rollback**
   ```bash
   kubectl rollout undo deployment/backend -n podland
   kubectl rollout undo deployment/frontend -n podland
   ```

2. **Database Rollback**
   ```bash
   # Restore from pre-launch backup
   gunzip -c podland-YYYYMMDD-HHMMSS.sql.gz | psql -h localhost -U podland -d podland
   ```

3. **Communication**
   - Notify users via status page
   - Document incident
   - Schedule re-launch

---

## Success Metrics

| Metric | Target | Actual |
|--------|--------|--------|
| API p95 latency | < 500ms | - |
| Error rate | < 0.1% | - |
| Uptime | > 99.9% | - |
| User signups (Week 1) | 50+ | - |
| VMs created (Week 1) | 100+ | - |

---

## Sign-Off

| Role | Name | Date | Status |
|------|------|------|--------|
| Engineering | | | [ ] |
| Operations | | | [ ] |
| Security | | | [ ] |
| Product | | | [ ] |

---

*Last updated: 2026-03-30*

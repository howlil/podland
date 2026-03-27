# Phase 2: Core VM - Documentation

**Phase:** 2 of 5
**Status:** Complete
**Goal:** Users can create and manage VMs with resource quotas enforced

---

## Table of Contents

1. [Overview](#overview)
2. [VM Lifecycle](#vm-lifecycle)
3. [API Endpoints](#api-endpoints)
4. [Quota System](#quota-system)
5. [Troubleshooting](#troubleshooting)
6. [Example Requests/Responses](#example-requestsresponses)
7. [Common Errors](#common-errors)

---

## Overview

Phase 2 implements the core VM management functionality for Podland. Users can create, start, stop, restart, and delete virtual machines with enforced resource quotas.

### Key Features

- **VM Creation**: Create VMs with name, OS (Ubuntu 22.04, Debian 12), and tier (nano → xlarge)
- **VM Lifecycle**: Start, stop, restart, and delete VMs
- **Quota Enforcement**: Per-user resource limits (CPU, RAM, storage, VM count)
- **Role-Based Tiers**: External users limited to nano/micro tiers; internal users can access all tiers
- **SSH Key Generation**: Ed25519 keypair generated per VM
- **Kubernetes Backend**: VMs run as containers in k3s with resource limits and security contexts

### Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Frontend  │────▶│  API Server  │────▶│   k3s       │
│  (React)    │     │   (Go)       │     │  (K8s)      │
└─────────────┘     └──────────────┘     └─────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │  PostgreSQL  │
                    │  (Database)  │
                    └──────────────┘
```

### Database Schema

**Tables:**
- `vms` - VM instances
- `user_quotas` - User resource limits
- `user_quota_usage` - Current resource usage
- `tiers` - VM tier definitions

See `.planning/phases/2-PLAN.md` for full schema.

---

## VM Lifecycle

### States

| State | Description |
|-------|-------------|
| `pending` | VM is being created (k8s resources provisioning) |
| `running` | VM is active and accessible |
| `stopped` | VM is stopped (scaled to 0 replicas) |
| `error` | VM creation failed or encountered an error |
| `deleting` | VM is being deleted |

### State Transitions

```
                    ┌──────────┐
                    │ pending  │
                    └────┬─────┘
                         │
              ┌──────────┼──────────┐
              │          │          │
              ▼          │          ▼
         ┌────────┐      │     ┌─────────┐
         │ running│◀─────┴────▶│ stopped │
         └───┬────┘            └────┬────┘
             │                      │
             │                      │
             ▼                      ▼
         ┌─────────┐           ┌─────────┐
         │ deleting│           │ deleting│
         └────┬────┘           └────┬────┘
              │                     │
              ▼                     ▼
         ┌─────────┐           ┌─────────┐
         │ deleted │           │ deleted │
         └─────────┘           └─────────┘
```

### Lifecycle Operations

#### Create VM

1. User submits VM creation request (name, OS, tier)
2. API validates JWT and user role
3. API checks tier availability (external vs internal)
4. API checks quota (SELECT FOR UPDATE to prevent race conditions)
5. Generate Ed25519 SSH keypair
6. Create VM record in database (status: `pending`)
7. Update quota usage
8. Async: Create k8s resources (namespace, PVC, Deployment, Service, Ingress)
9. Async: Update VM status to `running` on success or `error` on failure
10. Return 202 Accepted with VM ID and SSH private key (shown once!)

#### Start VM

1. User submits start request
2. API validates VM ownership and current status (must be `stopped`)
3. Scale k8s Deployment to 1 replica
4. Update VM status to `pending`
5. Return 202 Accepted
6. Async: Monitor pod status, update to `running` when ready

#### Stop VM

1. User submits stop request
2. API validates VM ownership and current status (must be `running`)
3. Scale k8s Deployment to 0 replicas
4. Update VM status to `stopped`
5. Return 202 Accepted

#### Restart VM

1. User submits restart request
2. API validates VM ownership and current status (must be `running`)
3. Scale k8s Deployment to 0 replicas (stop)
4. Wait for pods to terminate
5. Scale k8s Deployment to 1 replica (start)
6. Update VM status to `pending`
7. Return 202 Accepted
8. Async: Monitor pod status, update to `running` when ready

#### Delete VM

1. User submits delete request
2. API validates VM ownership
3. Create snapshot (7-day retention)
4. Delete k8s resources (Deployment, Service, Ingress, PVC, namespace if empty)
5. Update quota usage (decrement)
6. Delete VM record from database
7. Return 200 OK

---

## API Endpoints

### Base URL

```
http://localhost:8080/api/vms
```

### Authentication

All endpoints require JWT authentication via `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

### Endpoints

#### POST /api/vms

Create a new VM.

**Request:**
```json
{
  "name": "my-vm",
  "os": "ubuntu-2204",
  "tier": "nano"
}
```

**Response (202 Accepted):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
  "message": "VM is being created. SSH key shown only once - download now!"
}
```

**Errors:**
- `400 Bad Request` - Invalid request body or tier
- `401 Unauthorized` - Missing or invalid JWT
- `403 Forbidden` - Quota exceeded or tier not available for role

---

#### GET /api/vms

List all VMs for the authenticated user.

**Response (200 OK):**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-vm",
    "os": "ubuntu-2204",
    "tier": "nano",
    "cpu": 0.25,
    "ram": 536870912,
    "storage": 5368709120,
    "status": "running",
    "domain": "my-vm.user-123.podland.app",
    "created_at": "2026-03-27T10:00:00Z"
  }
]
```

---

#### GET /api/vms/{id}

Get details for a specific VM.

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-vm",
  "os": "ubuntu-2204",
  "tier": "nano",
  "cpu": 0.25,
  "ram": 536870912,
  "storage": 5368709120,
  "status": "running",
  "domain": "my-vm.user-123.podland.app",
  "ssh_public_key": "ssh-ed25519 AAAA...",
  "created_at": "2026-03-27T10:00:00Z",
  "updated_at": "2026-03-27T10:00:30Z",
  "started_at": "2026-03-27T10:00:30Z"
}
```

**Errors:**
- `404 Not Found` - VM not found or user doesn't own it

---

#### POST /api/vms/{id}/start

Start a stopped VM.

**Response (202 Accepted):**
```json
{
  "status": "pending"
}
```

**Errors:**
- `400 Bad Request` - VM must be stopped to start
- `404 Not Found` - VM not found or user doesn't own it

---

#### POST /api/vms/{id}/stop

Stop a running VM.

**Response (202 Accepted):**
```json
{
  "status": "stopped"
}
```

**Errors:**
- `400 Bad Request` - VM must be running to stop
- `404 Not Found` - VM not found or user doesn't own it

---

#### POST /api/vms/{id}/restart

Restart a running VM.

**Response (202 Accepted):**
```json
{
  "status": "pending"
}
```

**Errors:**
- `400 Bad Request` - VM must be running to restart
- `404 Not Found` - VM not found or user doesn't own it

---

#### DELETE /api/vms/{id}

Delete a VM.

**Response (200 OK):**
```json
{
  "message": "VM deleted successfully"
}
```

**Errors:**
- `404 Not Found` - VM not found or user doesn't own it

---

## Quota System

### Default Quotas

| User Type | CPU | RAM | Storage | VM Count |
|-----------|-----|-----|---------|----------|
| External (default) | 0.5 | 1 GB | 10 GB | 2 |
| Internal (NIM 1152+) | 4.0 | 8 GB | 100 GB | 5 |

### VM Tiers

| Tier | CPU | RAM | Storage | Min Role |
|------|-----|-----|---------|----------|
| nano | 0.25 | 512 MB | 5 GB | external |
| micro | 0.5 | 1 GB | 10 GB | external |
| small | 1.0 | 2 GB | 20 GB | internal |
| medium | 2.0 | 4 GB | 40 GB | internal |
| large | 4.0 | 8 GB | 80 GB | internal |
| xlarge | 4.0 | 8 GB | 100 GB | internal |

### Quota Enforcement

Quota checks use database transactions with `SELECT FOR UPDATE` to prevent race conditions:

```sql
BEGIN;

-- Lock quota row for update
SELECT cpu_limit, ram_limit, storage_limit, vm_count_limit
FROM user_quotas
WHERE user_id = $1
FOR UPDATE;

-- Get current usage
SELECT cpu_used, ram_used, storage_used, vm_count
FROM user_quota_usage
WHERE user_id = $1;

-- Check if new VM fits
-- (application logic)

-- Update usage
UPDATE user_quota_usage
SET cpu_used = cpu_used + $1,
    ram_used = ram_used + $2,
    storage_used = storage_used + $3,
    vm_count = vm_count + $4
WHERE user_id = $5;

COMMIT;
```

### Quota Dashboard

The frontend displays quota usage as a progress bar:

```
CPU:    [████████░░] 0.4 / 0.5 (80%)
RAM:    [████░░░░░░] 400 MB / 1 GB (40%)
Storage:[██░░░░░░░░] 2 GB / 10 GB (20%)
VMs:    [██████████] 2 / 2 (100%)
```

---

## Troubleshooting

### VM Stuck in "pending" Status

**Symptoms:** VM remains in `pending` status for more than 2 minutes.

**Causes:**
1. k3s cluster is unavailable or overloaded
2. Insufficient cluster resources (CPU, memory)
3. PVC provisioning failure
4. NetworkPolicy creation failure

**Resolution:**
```bash
# Check k3s cluster status
kubectl get nodes
kubectl get pods -A

# Check VM namespace
kubectl get namespace user-<user-id>

# Check for pending PVCs
kubectl get pvc -n user-<user-id>

# Check Deployment status
kubectl get deployment vm-<vm-id> -n user-<user-id>
kubectl describe deployment vm-<vm-id> -n user-<user-id>

# Check pod events
kubectl get pods -n user-<user-id>
kubectl describe pod <pod-name> -n user-<user-id>
```

---

### "Quota Exceeded" Error

**Symptoms:** VM creation fails with 403 Forbidden and "Quota exceeded" message.

**Causes:**
1. User has reached CPU, RAM, storage, or VM count limit
2. Concurrent VM creation exhausted quota

**Resolution:**
```sql
-- Check user quota
SELECT * FROM user_quotas WHERE user_id = '<user-id>';

-- Check current usage
SELECT * FROM user_quota_usage WHERE user_id = '<user-id>';

-- For superadmin: increase quota
UPDATE user_quotas
SET cpu_limit = 1.0, ram_limit = 2147483648
WHERE user_id = '<user-id>';

-- Or delete unused VMs to free quota
DELETE FROM vms WHERE user_id = '<user-id>' AND status = 'stopped';
```

---

### "Tier Not Available" Error

**Symptoms:** VM creation fails with 403 Forbidden and "Tier not available for your role" message.

**Causes:**
1. External user attempting to create internal tier (small, medium, large, xlarge)

**Resolution:**
- Use nano or micro tier for external users
- Upgrade user role to internal (requires NIM 1152+ or admin action)

---

### SSH Connection Failed

**Symptoms:** Cannot SSH to VM despite having private key.

**Causes:**
1. VM not yet in `running` status
2. SSH service not started in container
3. Network ingress not configured
4. Wrong SSH key (private key shown only once at creation)

**Resolution:**
```bash
# Verify VM status
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/vms/<vm-id>

# Check domain resolution
nslookup <vm-domain>

# Test SSH connection (verbose)
ssh -i <private-key> -o StrictHostKeyChecking=no -v user@<vm-domain>

# Check k8s Service
kubectl get service vm-<vm-id>-ssh -n user-<user-id>

# Check pod logs
kubectl logs -n user-<user-id> -l app=vm-<vm-id>
```

---

### VM Creation Timeout

**Symptoms:** VM creation request times out after 30 seconds.

**Causes:**
1. k3s API server overloaded
2. Database connection pool exhausted
3. Network latency

**Resolution:**
```bash
# Check API server health
curl http://localhost:8080/health

# Check database connections
psql -c "SELECT count(*) FROM pg_stat_activity;"

# Retry with exponential backoff
# (implement in client code)
```

---

## Example Requests/Responses

### cURL Examples

#### Create VM

```bash
curl -X POST http://localhost:8080/api/vms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "name": "my-first-vm",
    "os": "ubuntu-2204",
    "tier": "nano"
  }'
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZWQyNTUxOQAAACB...\n-----END OPENSSH PRIVATE KEY-----",
  "message": "VM is being created. SSH key shown only once - download now!"
}
```

---

#### List VMs

```bash
curl http://localhost:8080/api/vms \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-first-vm",
    "os": "ubuntu-2204",
    "tier": "nano",
    "cpu": 0.25,
    "ram": 536870912,
    "storage": 5368709120,
    "status": "running",
    "domain": "my-first-vm.user-123.podland.app",
    "created_at": "2026-03-27T10:00:00Z"
  }
]
```

---

#### Stop VM

```bash
curl -X POST http://localhost:8080/api/vms/550e8400-e29b-41d4-a716-446655440000/stop \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "status": "stopped"
}
```

---

#### Start VM

```bash
curl -X POST http://localhost:8080/api/vms/550e8400-e29b-41d4-a716-446655440000/start \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "status": "pending"
}
```

---

#### Delete VM

```bash
curl -X DELETE http://localhost:8080/api/vms/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "message": "VM deleted successfully"
}
```

---

### JavaScript/TypeScript Examples

```typescript
const API_BASE = 'http://localhost:8080/api';
const JWT_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';

async function createVM(name: string, os: string, tier: string) {
  const response = await fetch(`${API_BASE}/vms`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
    body: JSON.stringify({ name, os, tier }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message);
  }

  return response.json();
}

async function listVMs() {
  const response = await fetch(`${API_BASE}/vms`, {
    headers: {
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to list VMs');
  }

  return response.json();
}

async function stopVM(vmId: string) {
  const response = await fetch(`${API_BASE}/vms/${vmId}/stop`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to stop VM');
  }

  return response.json();
}

async function startVM(vmId: string) {
  const response = await fetch(`${API_BASE}/vms/${vmId}/start`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to start VM');
  }

  return response.json();
}

async function deleteVM(vmId: string) {
  const response = await fetch(`${API_BASE}/vms/${vmId}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to delete VM');
  }

  return response.json();
}

// Usage
async function main() {
  try {
    // Create VM
    const vm = await createVM('my-vm', 'ubuntu-2204', 'nano');
    console.log('VM created:', vm.id);
    console.log('SSH key (save this!):', vm.ssh_key);

    // Wait for VM to be running
    await new Promise(resolve => setTimeout(resolve, 30000));

    // List VMs
    const vms = await listVMs();
    console.log('VMs:', vms);

    // Stop VM
    await stopVM(vm.id);
    console.log('VM stopped');

    // Start VM
    await startVM(vm.id);
    console.log('VM started');

    // Delete VM
    await deleteVM(vm.id);
    console.log('VM deleted');
  } catch (error) {
    console.error('Error:', error);
  }
}
```

---

## Common Errors

### 400 Bad Request

| Error | Cause | Resolution |
|-------|-------|------------|
| `Invalid request body` | Malformed JSON or missing fields | Validate request body format |
| `Invalid tier` | Tier name not in [nano, micro, small, medium, large, xlarge] | Use valid tier name |
| `VM must be stopped to start` | Attempting to start a running/pending VM | Stop VM first or check status |
| `VM must be running to stop` | Attempting to stop a stopped/pending VM | Start VM first or check status |

---

### 401 Unauthorized

| Error | Cause | Resolution |
|-------|-------|------------|
| `Missing authentication` | No Authorization header | Add `Authorization: Bearer <token>` |
| `Invalid token` | Expired or malformed JWT | Refresh token or re-authenticate |

---

### 403 Forbidden

| Error | Cause | Resolution |
|-------|-------|------------|
| `Quota exceeded` | User has reached resource limit | Delete unused VMs or request quota increase |
| `Tier not available for your role` | External user attempting internal tier | Use nano/micro tier or upgrade role |

---

### 404 Not Found

| Error | Cause | Resolution |
|-------|-------|------------|
| `VM not found` | VM ID doesn't exist or user doesn't own it | Check VM ID and ownership |

---

### 409 Conflict

| Error | Cause | Resolution |
|-------|-------|------------|
| `VM name already exists` | Duplicate VM name for user | Use unique VM name |

---

### 500 Internal Server Error

| Error | Cause | Resolution |
|-------|-------|------------|
| `Failed to create VM` | k8s resource creation failed | Check k3s cluster health and logs |
| `Failed to update quota` | Database error | Check database connectivity |

---

## Testing

### Load Testing

Run the load test script to verify system performance under concurrent load:

```bash
./scripts/load-test-vm-creation.sh \
  --jwt-token "eyJhbGc..." \
  --concurrency 100 \
  --timeout 30 \
  --verbose
```

**Metrics:**
- Success rate (% of successful VM creations)
- Average creation time (ms)
- P50, P95, P99 latencies
- Quota race condition detection
- Error breakdown by type

### Integration Testing

Run the VM lifecycle integration tests:

```bash
./scripts/test-vm-lifecycle.sh \
  --jwt-token "eyJhbGc..." \
  --verbose
```

**Tests:**
1. VM Lifecycle: Create → Running → Stop → Start → Restart → Delete
2. Quota Enforcement: Create VMs until quota exceeded
3. SSH Key: Verify SSH key format and availability
4. Tier Restrictions: External users cannot create internal tiers

---

## Security Considerations

### SSH Keys

- Ed25519 keys generated per VM
- Private key shown **only once** at creation time
- Store private key securely immediately
- Public key stored in database for reference

### Container Security

- VMs run as UID 1000 (non-root)
- Pod Security Standards: `restricted`
- Capabilities dropped: `ALL`
- Privilege escalation: disabled
- NetworkPolicy denies inter-namespace traffic

### Authentication

- JWT required for all API endpoints
- Token expiration: 24 hours (configurable)
- HTTP-only cookies for web clients
- CORS enabled for frontend domain only

---

## Related Documents

- `.planning/phases/2-PLAN.md` - Phase 2 implementation plan
- `.planning/STATE.md` - Project state tracking
- `QUICKSTART.md` - Quick start guide
- `README.md` - Project overview

---

*Last updated: 2026-03-27*
*Phase 2 Status: Complete*

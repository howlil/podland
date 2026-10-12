# 04 — Data Model & API Contract (Draft)

## Logical entities

| Entity | Fields utama | Catatan |
|---|---|---|
| `installation` | id, setup_state, owner_user_id | tepat satu V1 |
| `users`, `sessions` | id, github_user_id, expires_at | owner-only |
| `integrations` | id, provider, credential_ref, status | credentials terenkripsi |
| `projects` | id, name, created_at | kelompok apps/services |
| `applications` | id, project_id, slug, desired_state, active_deployment_id | slug/domain stable |
| `app_revisions` | id, app_id, source_snapshot, config_snapshot, created_at | immutable |
| `deployments` | id, app_id, revision_id, status, image_digest, timestamps, error_code | deployment immutable history |
| `operations` | id, resource_id, kind, state, idempotency_key, retry_count | persistent job queue |
| `runtime_instances` | id, deployment_id, runtime_id, state | Podman external ref |
| `domains` | id, app_id, hostname, route_status | unique hostname |
| `environment_variables` | id, app_id, key, encrypted_value, revision | no plaintext response |
| `volumes` | id, owner_ref, name, mount_path, deletion_state | durable data |
| `services` | id, project_id, type, image_version, state | managed DB |
| `webhook_deliveries` | delivery_id, provider, received_at, processed_at | replay dedup |

**Data invariants:** ULID/UUID keys, foreign keys ON, timestamps UTC, migrations versioned, SQLite WAL on local filesystem. Index `deployments(app_id, created_at)`, unique `domains(hostname)` dan `webhook_deliveries(delivery_id)`.

## API conventions

- REST prefix `/api/v1`, JSON body, `OpenAPI 3.1` sebagai kontrak source-of-truth; frontend client typed dari kontrak bila generator ringan/matang.
- Error shape: `{ "error": {"code": "CONFLICT", "message": "...", "request_id": "..."} }`.
- Mutations yang async merespons `202 Accepted` dengan operation/deployment ID; fetch status terpisah.
- `Idempotency-Key` untuk create deployment dan operasi rawan retry. Mutasi app bersamaan → `409 Conflict`.
- Session cookie aman, CSRF protected, secrets write-only/redacted. Logs via SSE dengan bounded replay.

## Initial endpoint inventory

| Method | Path | Purpose |
|---|---|---|
| GET | `/healthz` | liveness, tidak boleh expose secrets |
| GET | `/api/v1/setup/state` | current bootstrap/setup state |
| POST | `/api/v1/setup/bootstrap` | claim installation dengan one-time token |
| GET | `/api/v1/auth/github/start` | start owner OAuth |
| GET | `/api/v1/auth/github/callback` | callback session |
| GET | `/api/v1/me` | owner info |
| GET/POST | `/api/v1/projects` | list/create |
| GET/PATCH/DELETE | `/api/v1/projects/{id}` | project detail/update/delete |
| GET/POST | `/api/v1/projects/{id}/apps` | list/create apps |
| GET/PATCH/DELETE | `/api/v1/apps/{id}` | app management |
| POST | `/api/v1/apps/{id}/deployments` | trigger deploy |
| GET | `/api/v1/apps/{id}/deployments` | deployment history |
| GET | `/api/v1/deployments/{id}` | deployment state |
| GET | `/api/v1/operations/{id}` | async operation state |
| POST | `/api/v1/apps/{id}/actions` | start/stop/restart/redeploy/rollback |
| GET | `/api/v1/apps/{id}/logs` | live logs (SSE) |
| GET/PUT | `/api/v1/apps/{id}/env` | config updates, values redacted on GET |
| GET/POST | `/api/v1/apps/{id}/domains` | domain management |
| GET/POST | `/api/v1/projects/{id}/services` | service management |
| POST | `/api/v1/webhooks/github` | signed webhook ingestion |

Cloudflare/GitHub connection endpoints **TBD** setelah auth mechanism dan GitHub App install/callback design divalidasi. Jangan menambahkan endpoint spekulatif ke kontrak final.

### Async deploy example

`POST /api/v1/apps/{id}/deployments`

```json
{"revision_id":"rev_123"}
```

`202 Accepted`

```json
{"id":"dep_123", "operation_id":"op_456", "status":"queued"}
```

Setiap slice implementation memperluas `contracts/openapi.yaml` beserta contract tests. Endpoint inventory di atas merupakan draft, bukan kontrak terimplementasi.

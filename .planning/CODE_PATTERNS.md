# Code Patterns & Constraints - Podland Backend

## Overview

This document defines the coding standards, patterns, and constraints for the Podland backend codebase. All contributors MUST follow these patterns to maintain code quality and consistency.

---

## Architecture Principles

### 1. Clean Architecture Layers

```
┌─────────────────────────────────────────┐
│  Handler Layer (handler/)               │
│  - HTTP request/response handling       │
│  - Input validation                     │
│  - Authentication middleware            │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  Usecase Layer (internal/usecase/)      │
│  - Business logic                       │
│  - Transaction management               │
│  - Error wrapping                       │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  Repository Layer (internal/repository/)│
│  - Data access                          │
│  - SQL queries                          │
│  - Context propagation                  │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  Entity Layer (internal/entity/)        │
│  - Domain models (no tags)              │
│  - Business logic methods               │
└─────────────────────────────────────────┘
```

### 2. Dependency Injection

**Rule:** All dependencies MUST be injected via constructor functions.

**✅ CORRECT:**
```go
type AuthHandler struct {
    userRepo    repository.UserRepository
    sessionRepo repository.SessionRepository
}

func NewAuthHandler(
    userRepo repository.UserRepository,
    sessionRepo repository.SessionRepository,
) *AuthHandler {
    return &AuthHandler{userRepo, sessionRepo}
}
```

**❌ WRONG:**
```go
var db *sql.DB  // NO global variables!

func HandleLogin(w http.ResponseWriter, r *http.Request) {
    dbWrapper := database.NewDB(db)  // Accessing global
}
```

---

## Coding Standards

### 1. Context Propagation

**Rule:** ALL database operations MUST accept `context.Context` as the first parameter.

**✅ CORRECT:**
```go
// Repository interface
type UserRepository interface {
    GetUserByID(ctx context.Context, id string) (*entity.User, error)
}

// Implementation
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
    err := r.db.QueryRowContext(ctx, query, id).Scan(...)
}

// Handler usage
user, err := h.userRepo.GetUserByID(r.Context(), userID)
```

**❌ WRONG:**
```go
func GetUserByID(id string) (*entity.User, error) {
    err := r.db.QueryRow(query, id).Scan(...)  // No context!
}
```

### 2. Error Handling

**Rule:** Use sentinel errors and wrap with context.

**Define errors in `pkg/errors/errors.go`:**
```go
var (
    ErrUserNotFound  = errors.New("user not found")
    ErrQuotaExceeded = errors.New("quota exceeded")
)
```

**Wrap errors in usecase layer:**
```go
user, err := uc.userRepo.GetUserByID(ctx, userID)
if err != nil {
    return nil, pkgerrors.Wrap(err, "failed to get user")
}
```

**Check errors in handler layer:**
```go
if err == repository.ErrUserNotFound {
    response.NotFound(w, "User not found")
    return
}
```

### 3. Repository Pattern

**Rule:** ALL data access MUST go through repository interfaces.

**Interface definition:**
```go
type UserRepository interface {
    CreateUser(ctx context.Context, input UserCreateInput) (*entity.User, error)
    GetUserByID(ctx context.Context, id string) (*entity.User, error)
    // ...
}
```

**Implementation:**
```go
type userRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
    return &userRepository{db: db}
}
```

**NO direct database access from handlers or usecases!**

### 4. Entity Models

**Rule:** Entities are pure domain objects with NO database or JSON tags.

**✅ CORRECT:**
```go
// internal/entity/user.go
type User struct {
    ID          string
    GithubID    string
    Email       string
    DisplayName string
    NIM         string
    Role        string
    CreatedAt   time.Time
}

// Business logic methods
func (u *User) IsInternal() bool {
    return u.Role == "internal"
}
```

**Database types (for persistence):**
```go
// internal/database/types.go
type User struct {
    ID          string     `json:"id"`
    GithubID    string     `json:"github_id"`
    Email       string     `json:"email"`
    DisplayName *string    `json:"display_name,omitempty"`
    // ...
}
```

### 5. HTTP Response Standardization

**Rule:** ALWAYS use `pkg/response` for consistent API responses.

**✅ CORRECT:**
```go
import "github.com/podland/backend/pkg/response"

func HandleGetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.userRepo.GetUserByID(r.Context(), userID)
    if err != nil {
        response.NotFound(w, "User not found")
        return
    }
    response.Success(w, http.StatusOK, user)
}
```

**❌ WRONG:**
```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(user)  // Inconsistent!
```

**Available response functions:**
```go
response.Success(w, status, data)
response.Error(w, status, message)
response.Created(w, data)
response.Accepted(w, data)
response.NoContent(w)
response.BadRequest(w, message)
response.Unauthorized(w, message)
response.Forbidden(w, message)
response.NotFound(w, message)
response.InternalError(w, message)
```

---

## Project Constraints

### 1. Directory Structure

**Middleware Organization:**

There are TWO middleware-related packages with DIFFERENT purposes:

1. **`middleware/`** - HTTP middleware (chi middleware pattern)
   - `CORSMiddleware` - CORS headers
   - `CSRFMiddleware` - CSRF token validation
   - `AuthMiddleware` - JWT validation & context injection

2. **`handler/auth_helper.go`** - Auth helper utilities (NOT middleware!)
   - `AuthHelper` - Extract user info from context
   - `GetAuthUserID()` - Get user ID from request
   - `GetAuthUserEmail()` - Get email from request
   
**Why separate?**
- `middleware/` = chi router middleware (wraps http.Handler)
- `handler/auth_helper.go` = helper functions for handlers

**DO NOT** create `handler/middleware/` directory - it's confusing!

```
apps/backend/
├── cmd/
│   └── main.go                  # Application entry point
├── handler/                     # HTTP handlers
│   ├── auth_handler.go          # Auth handler
│   ├── auth_helper.go           # Auth helper utilities
│   ├── vm_handler.go            # VM handler
│   └── health.go                # Health check
├── middleware/                  # HTTP middleware (CORS, CSRF, Auth)
│   └── middleware.go
├── internal/
│   ├── auth/                    # Auth logic (JWT, OAuth, sessions)
│   ├── config/                  # Configuration management
│   ├── database/                # DB connection & types (NO queries!)
│   ├── entity/                  # Domain models
│   ├── repository/              # Data access layer
│   ├── usecase/                 # Business logic
│   └── ssh/                     # SSH key generation
├── pkg/
│   ├── errors/                  # Application errors
│   └── response/                # HTTP response helpers
├── migrations/                  # SQL migrations
└── uploads/                     # Static files
```

### 2. Forbidden Patterns

**❌ NEVER DO THESE:**

1. **No global variables:**
   ```go
   var db *sql.DB  // FORBIDDEN
   ```

2. **No direct database access from handlers:**
   ```go
   func HandleUser(w http.ResponseWriter, r *http.Request) {
       db.Query(...)  // FORBIDDEN - use repository
   }
   ```

3. **No duplicate handler directories:**
   - Only `handler/` is allowed
   - No `handlers/`, `http/`, `api/` etc.

4. **No context-less database operations:**
   ```go
   db.QueryRow(query, id)  // FORBIDDEN
   db.QueryRowContext(ctx, query, id)  // REQUIRED
   ```

5. **No mixed response patterns:**
   ```go
   json.NewEncoder(w).Encode(data)  // FORBIDDEN
   response.Success(w, http.StatusOK, data)  // REQUIRED
   ```

### 3. YAGNI Principle

**Rule:** Do not add features "just in case".

**Current examples:**
- ❌ Removed: `internal/k8s/` - not currently used
- ❌ Removed: `ReconcileUsage` functions - not called anywhere
- ✅ Keep: k8s fields in database types (already in schema, might use later)

**Decision criteria:**
- Will we use this in the next 3 months? → If NO, don't add it
- Is this required for current features? → If NO, defer it
- Does this add complexity without immediate value? → If YES, skip it

---

## Testing Standards

### 1. Test File Organization

```
internal/
  usecase/
    vm_usecase.go
    vm_usecase_test.go  # Unit tests
  repository/
    user_repository.go
    # Integration tests optional
handler/
  auth_handler.go
  # Handler tests with mocked usecases
```

### 2. Mock Repositories for Testing

```go
type MockUserRepository struct {
    GetUserByIDFn func(ctx context.Context, id string) (*entity.User, error)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
    if m.GetUserByIDFn != nil {
        return m.GetUserByIDFn(ctx, id)
    }
    return nil, errors.New("GetUserByID not implemented")
}

// Implement ALL interface methods (even if no-op)
func (m *MockUserRepository) CreateActivityLog(ctx context.Context, userID string, action string, metadata map[string]interface{}) error {
    return nil // No-op for tests
}
```

---

## Build & Verification

### Required Checks Before Commit

```bash
# 1. Build must succeed
go build ./...

# 2. No vet warnings
go vet ./...

# 3. All tests must pass
go test ./...

# 4. (Optional) Run linter
golangci-lint run
```

### CI/CD Requirements

All PRs must pass:
- ✅ `go build ./...`
- ✅ `go vet ./...`
- ✅ `go test ./...`

---

## Database Conventions

### 1. Migration Files

- Location: `migrations/*.sql`
- Naming: `NNN_description.sql` (e.g., `002_phase2_vm_quota.sql`)
- Must be idempotent (use `CREATE TABLE IF NOT EXISTS`)

### 2. Soft Delete Pattern

All user-created resources MUST use soft delete:

```sql
deleted_at TIMESTAMP
```

```go
// Query always filters deleted records
WHERE id = $1 AND deleted_at IS NULL
```

### 3. Transaction Isolation

For quota operations and other critical updates:

```go
tx, err := db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
})
```

Use `SELECT FOR UPDATE` to prevent race conditions.

---

## Security Patterns

### 1. Authentication

- JWT access tokens (15 min expiry)
- Opaque refresh tokens (7 days, HTTP-only cookies)
- Token rotation on every refresh
- Token reuse detection → revoke all sessions

### 2. CSRF Protection

- Double-submit cookie pattern
- `XSRF-TOKEN` cookie + `X-XSRF-TOKEN` header
- Skip for auth endpoints (they have OAuth state parameter)

### 3. Input Validation

**ALWAYS validate in handler layer:**
```go
if req.Name == "" {
    response.BadRequest(w, "VM name is required")
    return
}
```

### 4. SQL Injection Prevention

**ALWAYS use parameterized queries:**
```go
// ✅ CORRECT
db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", userID)

// ❌ WRONG - NEVER DO THIS
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID)
db.QueryRow(query)
```

---

## Code Review Checklist

### Before Submitting PR:

- [ ] No global variables added
- [ ] All database operations use `context.Context`
- [ ] All errors are wrapped with context
- [ ] Response uses `pkg/response` package
- [ ] No duplicate code (DRY)
- [ ] No unused code (YAGNI)
- [ ] Simple solution (KISS)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes

---

## Decision Log

### ADR-001: Repository Pattern Only
**Date:** 2026-03-28  
**Status:** Accepted  
**Decision:** Use repository pattern exclusively. Remove direct database access from handlers.

### ADR-002: Context Propagation
**Date:** 2026-03-28  
**Status:** Accepted  
**Decision:** All repository methods MUST accept `context.Context` as first parameter.

### ADR-003: Single Handler Directory
**Date:** 2026-03-28  
**Status:** Accepted  
**Decision:** Consolidate to `handler/` only. No `handlers/` or other variants.

### ADR-004: Remove K8s Integration (YAGNI)
**Date:** 2026-03-28  
**Status:** Accepted  
**Decision:** Remove `internal/k8s/` directory. Add back when actually needed.

---

## References

- [Clean Code by Robert C. Martin](https://www.amazon.com/Clean-Code-Handbook-Software-Craftsmanship/dp/0132350882)
- [Clean Architecture by Robert C. Martin](https://www.amazon.com/Clean-Architecture-Craftsmans-Software-Structure/dp/0134494164)
- [Go Best Practices](https://github.com/golang-standards/project-layout)
- [Uber Go Style Guide](https://github.com/uber-go/guide)

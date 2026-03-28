# Current Architecture State - Podland Backend

**Last Updated:** 2026-03-28  
**Status:** ✅ REFACTORED - Clean Architecture Applied

---

## Current Folder Structure

```
apps/backend/
├── cmd/
│   └── main.go                      # Application entry point (131 LOC)
│
├── internal/                        # Private application code
│   ├── handler/                     # HTTP handlers (moved from root)
│   │   ├── auth_handler.go          # Auth handlers (407 LOC)
│   │   ├── auth_helper.go           # Auth utilities (101 LOC)
│   │   ├── vm_handler.go            # VM handlers (247 LOC)
│   │   └── health.go                # Health check (17 LOC)
│   │
│   ├── middleware/                  # HTTP middleware (moved from root)
│   │   └── middleware.go            # CORS, CSRF, Auth (91 LOC)
│   │
│   ├── auth/                        # Authentication logic
│   │   ├── jwt.go                   # JWT token handling (115 LOC)
│   │   ├── oauth.go                 # GitHub OAuth (167 LOC)
│   │   ├── session.go               # Session management (154 LOC)
│   │   └── ssh.go                   # SSH keygen (30 LOC, moved from ssh/)
│   │
│   ├── config/                      # Configuration
│   │   └── config.go                # Env validation (33 LOC)
│   │
│   ├── database/                    # Database layer
│   │   ├── database.go              # DB connection & migrations (112 LOC)
│   │   └── types.go                 # DB types (77 LOC)
│   │
│   ├── entity/                      # Domain models
│   │   ├── quota.go                 # Quota entity (91 LOC)
│   │   ├── user.go                  # User entity (37 LOC)
│   │   └── vm.go                    # VM entity (53 LOC)
│   │
│   ├── repository/                  # Data access layer
│   │   ├── quota_repository.go      # Quota operations (253 LOC)
│   │   ├── session_repository.go    # Session operations (262 LOC)
│   │   ├── types.go                 # Repository types/errors (60 LOC)
│   │   ├── user_repository.go       # User operations (196 LOC)
│   │   └── vm_repository.go         # VM operations (219 LOC)
│   │
│   ├── usecase/                     # Business logic layer
│   │   ├── quota_usecase.go         # Quota usecase (46 LOC)
│   │   ├── vm_usecase.go            # VM usecase (178 LOC)
│   │   └── vm_usecase_test.go       # VM tests (574 LOC)
│   │
│   └── test/                        # [FUTURE] Integration tests
│
├── pkg/                             # Public libraries
│   ├── errors/
│   │   └── errors.go                # Application errors (36 LOC)
│   └── response/
│       └── response.go              # HTTP response helpers (82 LOC)
│
├── migrations/                      # Database migrations
│   └── 002_phase2_vm_quota.sql      # VM/Quota schema (4.4 KB)
│
└── uploads/                         # Static files (avatars)
```

---

## Architecture Statistics

| Metric | Value |
|--------|-------|
| **Total Go Files** | 24 |
| **Total Lines of Code** | ~3,600 LOC |
| **Test Coverage** | 1 test file (574 LOC) |
| **Packages** | 12 |
| **Public Packages (pkg/)** | 2 |
| **Private Packages (internal/)** | 10 |

---

## Architecture Pattern: Clean Architecture

### Layer Structure

```
┌─────────────────────────────────────────────────────────────┐
│  Entry Point: cmd/main.go                                   │
│  - Dependency Injection                                     │
│  - Router Setup                                             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Transport Layer: internal/handler/                         │
│  - HTTP request/response handling                           │
│  - Input validation                                         │
│  - Response formatting                                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Business Logic: internal/usecase/                          │
│  - Core business rules                                      │
│  - Transaction management                                   │
│  - Error wrapping                                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Data Access: internal/repository/                          │
│  - SQL queries                                              │
│  - Database operations                                      │
│  - Context propagation                                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Domain Models: internal/entity/                            │
│  - Pure domain objects (no DB tags)                         │
│  - Business logic methods                                   │
└─────────────────────────────────────────────────────────────┘
```

### Cross-Cutting Concerns

```
┌─────────────────────────────────────────────────────────────┐
│  Middleware: internal/middleware/                           │
│  - CORS                                                     │
│  - CSRF Protection                                          │
│  - JWT Authentication                                       │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  Utilities: internal/auth/, pkg/                            │
│  - JWT utilities                                            │
│  - OAuth flow                                               │
│  - Session management                                       │
│  - SSH key generation                                       │
│  - Error types (pkg/errors)                                 │
│  - Response helpers (pkg/response)                          │
└─────────────────────────────────────────────────────────────┘
```

---

## Dependency Graph

```
cmd/main.go
    ↓
internal/handler/
    ├── internal/middleware/
    ├── internal/usecase/
    ├── internal/auth/
    ├── pkg/response
    └── internal/entity/
    
internal/usecase/
    ├── internal/repository/
    ├── internal/entity/
    └── pkg/errors/
    
internal/repository/
    ├── internal/database/
    ├── internal/entity/
    └── context.Context
    
internal/auth/
    ├── internal/repository/
    ├── pkg/errors/
    └── external libs (jwt, oauth2)
```

---

## Key Design Decisions

### 1. All Code in `internal/` Except `pkg/`

**Decision:** Move `handler/` and `middleware/` into `internal/`

**Rationale:**
- Consistency - all application code in one place
- Encapsulation - `internal/` prevents external imports
- Standard Go project layout compliance

### 2. Separate `pkg/` for Reusable Utilities

**Decision:** Keep `pkg/errors/` and `pkg/response/` as separate packages

**Rationale:**
- Clear separation of concerns
- Can be imported independently
- Follows Go standard project layout

### 3. Repository Pattern with Interfaces

**Decision:** Each repository has an interface + implementation

**Rationale:**
- Enables mock-based testing
- Clear contract for data access
- Easy to swap implementations if needed

### 4. Context Propagation

**Decision:** All database operations accept `context.Context`

**Rationale:**
- Supports request cancellation
- Prevents goroutine leaks
- Better resource cleanup

### 5. SSH Keygen in `auth/`

**Decision:** Move `ssh/keygen.go` → `auth/ssh.go`

**Rationale:**
- SSH keys are part of authentication flow
- Reduces directory nesting
- Logical grouping (auth utilities)

---

## Package Responsibilities

| Package | Responsibility | Dependencies |
|---------|---------------|--------------|
| `cmd/main.go` | App bootstrap, DI wiring | All internal packages |
| `internal/handler/` | HTTP layer | usecase, middleware, auth, pkg/response |
| `internal/middleware/` | HTTP middleware | auth |
| `internal/auth/` | Auth logic | repository, pkg/errors, external libs |
| `internal/usecase/` | Business logic | repository, entity, pkg/errors |
| `internal/repository/` | Data access | database, entity |
| `internal/entity/` | Domain models | None (pure Go) |
| `internal/database/` | DB connection | external libs (pq) |
| `internal/config/` | Config validation | external libs (godotenv) |
| `pkg/errors/` | Error types | None (standard lib only) |
| `pkg/response/` | Response helpers | None (standard lib only) |

---

## Import Path Conventions

```go
// Internal packages (cannot be imported outside this module)
import "github.com/podland/backend/internal/handler"
import "github.com/podland/backend/internal/usecase"
import "github.com/podland/backend/internal/repository"

// Public packages (can be imported by other modules)
import "github.com/podland/backend/pkg/errors"
import "github.com/podland/backend/pkg/response"
```

---

## Code Organization Principles

### 1. Separation of Concerns

- **Handlers:** Only HTTP logic (parsing, validation, response formatting)
- **Usecases:** Only business logic (rules, transactions, error wrapping)
- **Repositories:** Only database operations (SQL queries)
- **Entities:** Only domain models (no side effects)

### 2. Dependency Rule

**Dependencies point inward:**
```
handler → usecase → repository → entity
                 → database
```

**No circular dependencies!**

### 3. Interface Placement

**Interfaces are defined WHERE THEY ARE USED, not where they are implemented:**

```go
// internal/usecase/vm_usecase.go
type VMRepository interface {
    CreateVM(...)
    GetVMByID(...)
}

// internal/repository/vm_repository.go
type vmRepository struct { ... }
func (r *vmRepository) CreateVM(...) { ... }
```

### 4. Error Handling

**Standard pattern:**
```go
// Define in pkg/errors/errors.go
var ErrUserNotFound = errors.New("user not found")

// Wrap in usecase layer
return nil, pkgerrors.Wrap(err, "failed to get user")

// Check in handler layer
if err == repository.ErrUserNotFound {
    pkgresponse.NotFound(w, "User not found")
}
```

---

## Testing Strategy

### Current State

- **Unit Tests:** `internal/usecase/vm_usecase_test.go` (574 LOC)
- **Mock Repositories:** Generated manually for testing
- **Coverage:** Focused on business logic (usecase layer)

### Test Organization

```
internal/
└── usecase/
    ├── vm_usecase.go
    └── vm_usecase_test.go  # Unit tests with mocks
```

### Future Improvements

1. **Add Integration Tests:**
   ```
   internal/
   └── test/
       └── integration/
           ├── vm_test.go
           └── auth_test.go
   ```

2. **Add Handler Tests:**
   ```
   internal/
   └── handler/
       ├── vm_handler.go
       └── vm_handler_test.go  # With mocked usecases
   ```

---

## Build & Verification

### Commands

```bash
# Build
go build ./...

# Vet (static analysis)
go vet ./...

# Test
go test ./...

# Tidy dependencies
go mod tidy
```

### Current Status

```
✅ go build ./...    - SUCCESS
✅ go vet ./...      - No warnings
✅ go test ./...     - All tests pass
✅ go mod tidy       - Dependencies clean
```

---

## Recent Changes (2026-03-28 Refactoring)

### Moved to `internal/`

- ✅ `handler/` → `internal/handler/`
- ✅ `middleware/` → `internal/middleware/`
- ✅ `ssh/keygen.go` → `internal/auth/ssh.go`

### Reorganized `pkg/`

- ✅ `pkg/errors.go` → `pkg/errors/errors.go`
- ✅ `pkg/response.go` → `pkg/response/response.go`

### Simplified Structure

- **Before:** 8 top-level directories
- **After:** 4 top-level directories (`cmd/`, `internal/`, `pkg/`, `migrations/`)

### Impact

- **Better encapsulation** - All app code in `internal/`
- **Clearer organization** - Logical grouping
- **Standard compliance** - Follows Go project layout
- **Reduced nesting** - Flatter structure

---

## AI Assistant Guidelines

### Understanding the Pattern

When working on this codebase:

1. **New Handler?** → `internal/handler/`
2. **New Middleware?** → `internal/middleware/`
3. **New Business Logic?** → `internal/usecase/`
4. **New Repository?** → `internal/repository/`
5. **New Entity?** → `internal/entity/`
6. **New Utility?** → `pkg/` (if reusable) or `internal/auth/` (if auth-related)

### Import Rules

- **NEVER** import `internal/` from outside the module
- **ALWAYS** use `pkg/` for shared utilities
- **KEEP** dependencies pointing inward (handler → usecase → repository)

### Code Style

- **Handlers:** HTTP logic only, delegate to usecases
- **Usecases:** Business logic only, delegate to repositories
- **Repositories:** Database only, return entities
- **Entities:** Pure domain models, no side effects

---

## Next Steps

### Short Term

1. ✅ Structure refactoring complete
2. ⏳ Add integration test directory
3. ⏳ Add handler tests

### Medium Term

1. ⏳ Add API documentation (`docs/`)
2. ⏳ Add OpenAPI specs (`api/`)
3. ⏳ Add config files (`configs/`)

### Long Term

1. ⏳ Consider adding worker service (`cmd/worker/`)
2. ⏳ Consider adding gRPC (`api/grpc/`)
3. ⏳ Consider service extraction (microservices)

---

## References

- **CODE_PATTERNS.md** - Coding standards and constraints
- **CLEAN_CODE_REFACTORING.md** - Research and guidelines
- **REFACTORING_SUMMARY.md** - Refactoring changes summary

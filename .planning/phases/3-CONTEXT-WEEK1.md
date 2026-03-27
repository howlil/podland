# Phase 3 Context: Networking — Week 1 Refactor Foundation

**Phase:** 3 of 5
**Week:** 1 of 4 (Refactor Foundation)
**Goal:** Apply Clean Architecture pattern (KISS, YAGNI, DRY)
**Status:** Context gathered — ready for planning

---

## Prior Context (From Phase 2)

### What Works Well ✅
- VM CRUD API functional (7 endpoints)
- Quota enforcement with SELECT FOR UPDATE
- SSH key generation (Ed25519)
- k8s VMManager structure in place

### Problems Identified ❌
| Issue | Current State | Impact |
|-------|--------------|--------|
| **Handler Layer** | `handlers/vms.go` (472 lines) | Too much logic, hard to test |
| **Global Variables** | `db` global variable | Hidden dependency, hard to mock |
| **TODO Comments** | k8s integration not implemented | Fake implementation (sleep 2s) |
| **Repetitive Auth** | Auth check in every handler | DRY violation |
| **Inconsistent Errors** | Mixed logging/error handling | Hard to debug |
| **God File** | `internal/database/queries.go` (200+ lines) | Hard to navigate |

---

## Clean Architecture Decisions (Week 1 Refactor)

### **1. Layer Structure (4 Layers)**

```
apps/backend/
├── cmd/
│   └── main.go              # Entry point, DI only
├── handler/                 # HTTP Layer (THIN)
│   ├── vm_handler.go        # Only: parse request, call usecase, return response
│   ├── auth_handler.go
│   └── middleware/
│       └── auth.go
├── internal/
│   ├── entity/              # Core domain objects (NEW!)
│   │   ├── vm.go            # VM entity + business rules
│   │   ├── user.go          # User entity
│   │   └── quota.go         # Quota entity
│   ├── usecase/             # Business Logic Layer (NEW!)
│   │   ├── vm_usecase.go    # VM business logic
│   │   └── quota_usecase.go # Quota business logic
│   ├── repository/          # Data access (renamed from database/)
│   │   ├── vm_repository.go # Per entity (not god file)
│   │   ├── quota_repository.go
│   │   └── user_repository.go
│   └── infrastructure/      # External services
│       ├── k8s/             # Kubernetes client
│       └── ssh/             # SSH key generation
└── pkg/                     # Shared utilities
    ├── errors/              # Common errors (DRY)
    └── response/            # HTTP response helpers (DRY)
```

**Key Principles:**
- **KISS:** Simple 4 layers, no over-engineering
- **YAGNI:** Only create what's needed now
- **DRY:** Shared `pkg/errors`, `pkg/response`

---

### **2. Repository Pattern (Per Entity)**

**Current:** `internal/database/queries.go` (200+ lines, all queries mixed)

**Proposed:**
```
internal/repository/
├── vm_repository.go      # ~100 lines (VM queries only)
├── quota_repository.go   # ~80 lines (Quota queries only)
├── user_repository.go    # ~60 lines (User queries only)
└── types.go              # Shared types
```

**Interface Pattern:**
```go
// repository/vm_repository.go
type VMRepository interface {
    CreateVM(ctx context.Context, input VMCreateInput) (*entity.VM, error)
    GetVMByID(ctx context.Context, id string) (*entity.VM, error)
    GetUserVMs(ctx context.Context, userID string) ([]*entity.VM, error)
    UpdateVMStatus(ctx context.Context, id, status string) error
    DeleteVM(ctx context.Context, id string) error
}

type vmRepository struct {
    db *sql.DB
}

func NewVMRepository(db *sql.DB) VMRepository {
    return &vmRepository{db: db}
}
```

**Benefits:**
- ✅ Easy to navigate (VM queries in `vm_repository.go`)
- ✅ Parallel development (dev A: VM, dev B: User)
- ✅ Clear interface per entity
- ✅ Testable (mock repository interface)

---

### **3. Usecase Layer (Business Logic)**

**Purpose:** Extract business logic from handlers

**Pattern:**
```go
// usecase/vm_usecase.go
type VMUsecase struct {
    vmRepo    repository.VMRepository
    quotaRepo repository.QuotaRepository
}

func NewVMUsecase(vmRepo repository.VMRepository, quotaRepo repository.QuotaRepository) *VMUsecase {
    return &VMUsecase{
        vmRepo:    vmRepo,
        quotaRepo: quotaRepo,
    }
}

func (uc *VMUsecase) CreateVM(ctx context.Context, userID string, input CreateVMInput) (*entity.VM, string, error) {
    // 1. Validation
    // 2. Get tier
    // 3. Check quota
    // 4. Create VM
    // 5. Update quota
    // 6. Return
}
```

**Benefits:**
- ✅ Testable without HTTP
- ✅ Reusable (CLI, gRPC, etc.)
- ✅ Clear separation (HTTP vs Business)

---

### **4. Handler Layer (THIN - HTTP Only)**

**Current (Phase 2):**
```go
// 472 lines: validation + business logic + DB + k8s + response
func HandleCreateVM(w http.ResponseWriter, r *http.Request) {
    // 50+ lines validation
    // Check quota
    // Generate SSH key
    // Create VM in database
    // Update quota
    // Create k8s resources (async)
    // Return response
}
```

**Proposed (Phase 3):**
```go
// ~30 lines: parse request, call usecase, return response
func (h *VMHandler) HandleCreateVM(w http.ResponseWriter, r *http.Request) {
    userID := getAuthUserID(r)
    
    var req CreateVMRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    vm, sshKey, err := h.vmUsecase.CreateVM(r.Context(), userID, CreateVMInput{
        Name: req.Name,
        OS:   req.OS,
        Tier: req.Tier,
    })
    if err != nil {
        if err == usecase.ErrQuotaExceeded {
            response.Error(w, http.StatusForbidden, "Quota exceeded")
            return
        }
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    response.JSON(w, http.StatusAccepted, CreateVMResponse{
        ID:      vm.ID,
        Status:  vm.Status,
        SSHKey:  sshKey,
    })
}
```

**Benefits:**
- ✅ Easy to read (30 lines vs 472)
- ✅ Easy to test (mock usecase)
- ✅ Consistent error handling

---

### **5. Entity Layer (Domain Objects)**

**Purpose:** Core domain objects with business rules

**Pattern:**
```go
// entity/vm.go
package entity

import "time"

type VM struct {
    ID        string
    UserID    string
    Name      string
    OS        string
    Tier      string
    CPU       float64
    RAM       int64
    Storage   int64
    Status    string
    Domain    string
    CreatedAt time.Time
}

// Business rule method
func (v *VM) IsRunning() bool {
    return v.Status == "running"
}

// Business rule: Check if VM can be started
func (v *VM) CanStart() bool {
    return v.Status == "stopped"
}
```

**Why:**
- ✅ No DB tags (handled in repository)
- ✅ No JSON tags (handled in handler)
- ✅ Pure domain logic
- ✅ Business methods on entity

---

### **6. Dependency Injection (cmd/main.go)**

**Pattern:**
```go
// cmd/main.go
func main() {
    // 1. Database
    db := sql.Open("postgres", getEnv("DATABASE_URL"))
    
    // 2. Repositories
    vmRepo := repository.NewVMRepository(db)
    quotaRepo := repository.NewQuotaRepository(db)
    
    // 3. Usecases
    vmUsecase := usecase.NewVMUsecase(vmRepo, quotaRepo)
    
    // 4. Handlers
    vmHandler := handler.NewVMHandler(vmUsecase)
    
    // 5. Router
    r := chi.NewRouter()
    r.Post("/api/vms", vmHandler.HandleCreateVM)
    
    // 6. Start server
    http.ListenAndServe(":8080", r)
}
```

**Benefits:**
- ✅ Single place for DI
- ✅ Clear dependencies
- ✅ Easy to mock for tests

---

### **7. Error Handling Standardization**

**Current:** Inconsistent (some log, some don't)

**Proposed:**
```go
// pkg/errors/errors.go
package errors

import "errors"

// Common errors (DRY - defined once)
var (
    ErrVMNotFound    = errors.New("vm not found")
    ErrQuotaExceeded = errors.New("quota exceeded")
    ErrInvalidTier   = errors.New("invalid tier")
)

// Wrap error with context
func Wrap(err error, message string) error {
    return fmt.Errorf("%s: %w", message, err)
}
```

**Usage:**
```go
// In usecase/repository
if err != nil {
    return nil, errors.Wrap(err, "failed to create VM")
}

// In handler
if err == usecase.ErrQuotaExceeded {
    response.Error(w, http.StatusForbidden, "Quota exceeded")
    return
}
```

---

### **8. Response Helpers (DRY)**

```go
// pkg/response/response.go
package response

func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string) {
    JSON(w, status, map[string]string{"error": message})
}
```

**Usage:**
```go
// In handler (DRY - no repetition)
response.JSON(w, http.StatusAccepted, vm)
response.Error(w, http.StatusBadRequest, "Invalid request")
```

---

## Week 1 Scope (Refactor Foundation)

### **In Scope (Week 1 Only)**
- [ ] Create `entity/` layer (VM, User, Quota)
- [ ] Create `repository/` layer (split from `database/`)
- [ ] Create `usecase/` layer (extract business logic)
- [ ] Refactor `handler/` to be thin
- [ ] Create `pkg/errors/` and `pkg/response/`
- [ ] Update `cmd/main.go` with DI
- [ ] Remove global `db` variable
- [ ] Implement proper error handling
- [ ] Add unit tests for usecases

### **Out of Scope (Week 2+)**
- [ ] Cloudflare integration (Week 2)
- [ ] Domain management (Week 3)
- [ ] SSL certificates (Week 3)
- [ ] Integration tests (Week 4)

---

## Migration Strategy

### **Approach: Incremental Refactor**

**Don't:** Rewrite everything at once (too risky)

**Do:** Refactor per handler as we touch them

**Week 1 Plan:**
1. **Day 1:** Create foundation layers (`entity/`, `repository/`, `pkg/`)
2. **Day 2:** Create `usecase/` layer (VM + Quota)
3. **Day 3:** Refactor VM handler to use usecase
4. **Day 4:** Update `cmd/main.go` with DI
5. **Day 5:** Test + fix issues

### **Backward Compatibility**

**Keep working:**
- All existing API endpoints
- Database schema (no changes)
- k8s integration (will be implemented properly in Week 2)

**Changes:**
- Internal structure only (no API changes)
- Users won't notice difference

---

## Success Criteria (Week 1)

- [ ] All API endpoints still work
- [ ] No global `db` variable
- [ ] Handler layer < 100 lines per file
- [ ] Business logic in `usecase/` layer
- [ ] Repository per entity (not god file)
- [ ] Consistent error handling
- [ ] Unit tests for usecases passing
- [ ] Build successful (`go build ./...`)

---

## Code Quality Metrics

| Metric | Before (Phase 2) | After (Week 1) |
|--------|-----------------|----------------|
| Handler lines | 472 (vms.go) | < 100 per handler |
| Repository files | 1 (queries.go) | 3+ (per entity) |
| Global variables | 1 (`db`) | 0 |
| Error handling | Inconsistent | Standardized |
| Testability | Hard (no interfaces) | Easy (mock interfaces) |
| DRY violations | High (auth check x8) | Low (middleware) |

---

## Next Steps

**After Week 1:**
- Week 2: Cloudflare Service (DNS + Tunnel automation)
- Week 3: Domain Management (subdomain assignment, SSL)
- Week 4: Testing + Documentation

**Command:** `/ez:execute-phase 3 --week 1`

---

*Context gathered: 2026-03-27*
*Clean Architecture pattern based on research (KISS, YAGNI, DRY)*
*Ready for Week 1 planning*

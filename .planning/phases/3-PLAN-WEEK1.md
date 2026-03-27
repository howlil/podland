# Phase 3 Plan: Week 1 — Refactor Foundation

**Phase:** 3 of 5 (Networking)
**Week:** 1 of 4 (Refactor Foundation)
**Goal:** Apply Clean Architecture pattern (KISS, YAGNI, DRY)
**Duration:** 5 days
**Status:** Ready for implementation

---

## Success Criteria

1. ✅ All API endpoints still work (backward compatible)
2. ✅ No global `db` variable
3. ✅ Handler layer < 100 lines per file
4. ✅ Business logic in `usecase/` layer
5. ✅ Repository per entity (not god file)
6. ✅ Consistent error handling
7. ✅ Unit tests for usecases passing
8. ✅ Build successful (`go build ./...`)

---

## Technical Milestones

- [ ] Create `entity/` layer (VM, User, Quota entities)
- [ ] Create `repository/` layer (VM, Quota, User repositories)
- [ ] Create `usecase/` layer (VM, Quota usecases)
- [ ] Create `pkg/errors/` and `pkg/response/`
- [ ] Refactor VM handler to be thin
- [ ] Update `cmd/main.go` with dependency injection
- [ ] Remove global `db` variable
- [ ] Add unit tests for usecases

---

## Implementation Tasks

### Day 1: Foundation Layers

#### Task 1.1: Create Entity Layer
**Estimate:** 2 hours
**Acceptance Criteria:**
- [ ] `internal/entity/vm.go` created
- [ ] `internal/entity/user.go` created
- [ ] `internal/entity/quota.go` created
- [ ] Entities have business rule methods
- [ ] No DB tags, no JSON tags (pure domain objects)

**Implementation:**
```go
// internal/entity/vm.go
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

// Business rule methods
func (v *VM) IsRunning() bool {
    return v.Status == "running"
}

func (v *VM) CanStart() bool {
    return v.Status == "stopped"
}

func (v *VM) CanStop() bool {
    return v.Status == "running"
}
```

---

#### Task 1.2: Create Repository Layer
**Estimate:** 3 hours
**Acceptance Criteria:**
- [ ] `internal/repository/vm_repository.go` created
- [ ] `internal/repository/quota_repository.go` created
- [ ] `internal/repository/user_repository.go` created
- [ ] Each repository has interface + implementation
- [ ] No business logic in repositories (only DB queries)

**Implementation:**
```go
// internal/repository/vm_repository.go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    
    "github.com/podland/backend/internal/entity"
)

// VMRepository defines interface for VM data access
type VMRepository interface {
    CreateVM(ctx context.Context, input VMCreateInput) (*entity.VM, error)
    GetVMByID(ctx context.Context, id string) (*entity.VM, error)
    GetVMByIDAndUser(ctx context.Context, id, userID string) (*entity.VM, error)
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

// CreateVM implements VMRepository
func (r *vmRepository) CreateVM(ctx context.Context, input VMCreateInput) (*entity.VM, error) {
    query := `
        INSERT INTO vms (user_id, name, os, tier, cpu, ram, storage, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')
        RETURNING id, created_at
    `
    
    vm := &entity.VM{}
    err := r.db.QueryRowContext(ctx, query,
        input.UserID, input.Name, input.OS, input.Tier,
        input.CPU, input.RAM, input.Storage,
    ).Scan(&vm.ID, &vm.CreatedAt)
    
    if err != nil {
        return nil, fmt.Errorf("create VM: %w", err)
    }
    
    // Map to entity
    vm.UserID = input.UserID
    vm.Name = input.Name
    vm.OS = input.OS
    vm.Tier = input.Tier
    vm.CPU = input.CPU
    vm.RAM = input.RAM
    vm.Storage = input.Storage
    vm.Status = "pending"
    
    return vm, nil
}

// ... other VM repository methods
```

---

#### Task 1.3: Create Shared Packages
**Estimate:** 1 hour
**Acceptance Criteria:**
- [ ] `pkg/errors/errors.go` created
- [ ] `pkg/response/response.go` created
- [ ] Common errors defined (ErrVMNotFound, ErrQuotaExceeded)
- [ ] Response helpers (JSON, Error)

**Implementation:**
```go
// pkg/errors/errors.go
package errors

import (
    "errors"
    "fmt"
)

// Common errors (DRY)
var (
    ErrVMNotFound    = errors.New("vm not found")
    ErrQuotaExceeded = errors.New("quota exceeded")
    ErrInvalidTier   = errors.New("invalid tier")
    ErrUnauthorized  = errors.New("unauthorized")
)

// Wrap error with context
func Wrap(err error, message string) error {
    return fmt.Errorf("%s: %w", message, err)
}
```

```go
// pkg/response/response.go
package response

import (
    "encoding/json"
    "net/http"
)

// JSON sends JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// Error sends error response
func Error(w http.ResponseWriter, status int, message string) {
    JSON(w, status, map[string]string{"error": message})
}

// Success sends success response
func Success(w http.ResponseWriter, status int, data interface{}) {
    JSON(w, status, map[string]interface{}{
        "success": true,
        "data":    data,
    })
}
```

---

### Day 2: Usecase Layer

#### Task 2.1: Create VM Usecase
**Estimate:** 3 hours
**Acceptance Criteria:**
- [ ] `internal/usecase/vm_usecase.go` created
- [ ] `CreateVM` method with validation + quota check
- [ ] `GetVM` method with ownership check
- [ ] `ListVMs` method
- [ ] `StartVM`, `StopVM`, `RestartVM`, `DeleteVM` methods
- [ ] No HTTP logic, no DB direct access

**Implementation:**
```go
// internal/usecase/vm_usecase.go
package usecase

import (
    "context"
    
    "github.com/podland/backend/internal/entity"
    "github.com/podland/backend/internal/repository"
    "github.com/podland/backend/pkg/errors"
)

// VMUsecase defines VM business logic
type VMUsecase struct {
    vmRepo    repository.VMRepository
    quotaRepo repository.QuotaRepository
}

// NewVMUsecase creates usecase with dependencies
func NewVMUsecase(vmRepo repository.VMRepository, quotaRepo repository.QuotaRepository) *VMUsecase {
    return &VMUsecase{
        vmRepo:    vmRepo,
        quotaRepo: quotaRepo,
    }
}

// CreateVMInput represents the input for creating a VM
type CreateVMInput struct {
    Name string
    OS   string
    Tier string
}

// CreateVM is the business logic for creating a VM
func (uc *VMUsecase) CreateVM(ctx context.Context, userID string, input CreateVMInput) (*entity.VM, error) {
    // 1. Validate input
    if input.Name == "" {
        return nil, errors.New("name is required")
    }
    if input.Tier == "" {
        return nil, errors.New("tier is required")
    }
    
    // 2. Get tier configuration
    tier, err := uc.quotaRepo.GetTier(ctx, input.Tier)
    if err != nil {
        return nil, errors.Wrap(err, "failed to get tier")
    }
    
    // 3. Check quota
    if err := uc.quotaRepo.CheckQuota(ctx, userID, tier.CPU, tier.RAM, tier.Storage); err != nil {
        return nil, errors.ErrQuotaExceeded
    }
    
    // 4. Create VM in database
    vm, err := uc.vmRepo.CreateVM(ctx, repository.VMCreateInput{
        UserID:  userID,
        Name:    input.Name,
        OS:      input.OS,
        Tier:    input.Tier,
        CPU:     tier.CPU,
        RAM:     tier.RAM,
        Storage: tier.Storage,
    })
    if err != nil {
        return nil, errors.Wrap(err, "failed to create VM")
    }
    
    // 5. Update quota usage
    if err := uc.quotaRepo.UpdateUsage(ctx, userID, tier.CPU, tier.RAM, tier.Storage, 1); err != nil {
        return nil, errors.Wrap(err, "failed to update quota")
    }
    
    return vm, nil
}

// GetVMByID gets a VM by ID with ownership check
func (uc *VMUsecase) GetVMByID(ctx context.Context, vmID, userID string) (*entity.VM, error) {
    vm, err := uc.vmRepo.GetVMByIDAndUser(ctx, vmID, userID)
    if err != nil {
        if err == repository.ErrVMNotFound {
            return nil, errors.ErrVMNotFound
        }
        return nil, errors.Wrap(err, "failed to get VM")
    }
    
    return vm, nil
}

// ListVMs lists all VMs for a user
func (uc *VMUsecase) ListVMs(ctx context.Context, userID string) ([]*entity.VM, error) {
    vms, err := uc.vmRepo.GetUserVMs(ctx, userID)
    if err != nil {
        return nil, errors.Wrap(err, "failed to list VMs")
    }
    
    return vms, nil
}

// StartVM starts a stopped VM
func (uc *VMUsecase) StartVM(ctx context.Context, vmID, userID string) error {
    vm, err := uc.vmRepo.GetVMByIDAndUser(ctx, vmID, userID)
    if err != nil {
        return errors.ErrVMNotFound
    }
    
    if !vm.CanStart() {
        return errors.New("VM must be stopped to start")
    }
    
    if err := uc.vmRepo.UpdateVMStatus(ctx, vmID, "pending"); err != nil {
        return errors.Wrap(err, "failed to update VM status")
    }
    
    // k8s integration will be called in handler (async)
    return nil
}

// ... other usecase methods (StopVM, RestartVM, DeleteVM)
```

---

#### Task 2.2: Create Quota Usecase
**Estimate:** 1 hour
**Acceptance Criteria:**
- [ ] `internal/usecase/quota_usecase.go` created
- [ ] `GetQuota` method
- [ ] `GetUsage` method

---

### Day 3: Refactor Handler Layer

#### Task 3.1: Refactor VM Handler
**Estimate:** 3 hours
**Acceptance Criteria:**
- [ ] `handler/vm_handler.go` refactored to < 100 lines
- [ ] No business logic in handler
- [ ] Use `pkg/response` helpers
- [ ] Consistent error handling
- [ ] All endpoints working

**Implementation:**
```go
// handler/vm_handler.go
package handler

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/podland/backend/internal/usecase"
    "github.com/podland/backend/pkg/response"
)

type VMHandler struct {
    vmUsecase *usecase.VMUsecase
}

func NewVMHandler(vmUsecase *usecase.VMUsecase) *VMHandler {
    return &VMHandler{vmUsecase: vmUsecase}
}

// CreateVMRequest represents the request body
type CreateVMRequest struct {
    Name string `json:"name"`
    OS   string `json:"os"`
    Tier string `json:"tier"`
}

// CreateVMResponse represents the response
type CreateVMResponse struct {
    ID      string `json:"id"`
    Status  string `json:"status"`
    Message string `json:"message"`
}

// HandleCreateVM creates a new VM
// POST /api/vms
func (h *VMHandler) HandleCreateVM(w http.ResponseWriter, r *http.Request) {
    userID := getAuthUserID(r)
    
    var req CreateVMRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid request")
        return
    }
    
    vm, err := h.vmUsecase.CreateVM(r.Context(), userID, usecase.CreateVMInput{
        Name: req.Name,
        OS:   req.OS,
        Tier: req.Tier,
    })
    if err != nil {
        if err == usecase.ErrQuotaExceeded {
            response.Error(w, http.StatusForbidden, "Quota exceeded")
            return
        }
        if err.Error() == "name is required" || err.Error() == "tier is required" {
            response.Error(w, http.StatusBadRequest, err.Error())
            return
        }
        response.Error(w, http.StatusInternalServerError, "Failed to create VM")
        return
    }
    
    response.JSON(w, http.StatusAccepted, CreateVMResponse{
        ID:      vm.ID,
        Status:  vm.Status,
        Message: "VM is being created",
    })
}

// HandleListVMs lists all VMs for the user
// GET /api/vms
func (h *VMHandler) HandleListVMs(w http.ResponseWriter, r *http.Request) {
    userID := getAuthUserID(r)
    
    vms, err := h.vmUsecase.ListVMs(r.Context(), userID)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, "Failed to list VMs")
        return
    }
    
    response.Success(w, http.StatusOK, vms)
}

// HandleGetVM gets a specific VM
// GET /api/vms/{id}
func (h *VMHandler) HandleGetVM(w http.ResponseWriter, r *http.Request) {
    userID := getAuthUserID(r)
    vmID := chi.URLParam(r, "id")
    
    vm, err := h.vmUsecase.GetVMByID(r.Context(), vmID, userID)
    if err != nil {
        if err.Error() == "vm not found" {
            response.Error(w, http.StatusNotFound, "VM not found")
            return
        }
        response.Error(w, http.StatusInternalServerError, "Failed to get VM")
        return
    }
    
    response.Success(w, http.StatusOK, vm)
}

// Helper function (should be in middleware)
func getAuthUserID(r *http.Request) string {
    userID := r.Context().Value("user_id")
    if userID == nil {
        return ""
    }
    return userID.(string)
}
```

---

### Day 4: Dependency Injection

#### Task 4.1: Update cmd/main.go
**Estimate:** 2 hours
**Acceptance Criteria:**
- [ ] Remove global `db` variable
- [ ] Create repositories
- [ ] Create usecases
- [ ] Create handlers
- [ ] Wire all dependencies
- [ ] Build successful

**Implementation:**
```go
// cmd/main.go
package main

import (
    "database/sql"
    "log"
    "net/http"
    "os"
    
    "github.com/go-chi/chi/v5"
    _ "github.com/lib/pq"
    
    "github.com/podland/backend/handler"
    "github.com/podland/backend/internal/repository"
    "github.com/podland/backend/internal/usecase"
)

func main() {
    // 1. Database connection
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL is required")
    }
    
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Test connection
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }
    
    // 2. Repositories
    vmRepo := repository.NewVMRepository(db)
    quotaRepo := repository.NewQuotaRepository(db)
    userRepo := repository.NewUserRepository(db)
    
    // 3. Usecases
    vmUsecase := usecase.NewVMUsecase(vmRepo, quotaRepo)
    
    // 4. Handlers
    vmHandler := handler.NewVMHandler(vmUsecase)
    authHandler := handler.NewAuthHandler(/* dependencies */)
    userHandler := handler.NewUserHandler(userRepo)
    
    // 5. Router
    r := chi.NewRouter()
    
    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    // Routes
    r.Post("/api/auth/login", authHandler.HandleLogin)
    r.Post("/api/auth/github/callback", authHandler.HandleCallback)
    
    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.Auth)
        r.Post("/api/vms", vmHandler.HandleCreateVM)
        r.Get("/api/vms", vmHandler.HandleListVMs)
        r.Get("/api/vms/{id}", vmHandler.HandleGetVM)
        r.Post("/api/vms/{id}/start", vmHandler.HandleStartVM)
        r.Post("/api/vms/{id}/stop", vmHandler.HandleStopVM)
        r.Post("/api/vms/{id}/restart", vmHandler.HandleRestartVM)
        r.Delete("/api/vms/{id}", vmHandler.HandleDeleteVM)
    })
    
    // 6. Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Starting server on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}
```

---

### Day 5: Testing + Documentation

#### Task 5.1: Unit Tests for Usecases
**Estimate:** 3 hours
**Acceptance Criteria:**
- [ ] Test `CreateVM` success case
- [ ] Test `CreateVM` quota exceeded
- [ ] Test `GetVMByID` success case
- [ ] Test `GetVMByID` not found
- [ ] Test `ListVMs` success case
- [ ] Mock repositories for tests

**Implementation:**
```go
// internal/usecase/vm_usecase_test.go
package usecase

import (
    "context"
    "testing"
    
    "github.com/podland/backend/internal/repository"
    "github.com/stretchr/testify/mock"
)

// MockVMRepository is a mock implementation
type MockVMRepository struct {
    mock.Mock
}

func (m *MockVMRepository) CreateVM(ctx context.Context, input repository.VMCreateInput) (*entity.VM, error) {
    args := m.Called(ctx, input)
    return args.Get(0).(*entity.VM), args.Error(1)
}

// ... other mock methods

func TestVMUsecase_CreateVM_Success(t *testing.T) {
    // Arrange
    vmRepo := new(MockVMRepository)
    quotaRepo := new(MockQuotaRepository)
    uc := NewVMUsecase(vmRepo, quotaRepo)
    
    input := CreateVMInput{
        Name: "test-vm",
        OS:   "ubuntu-2204",
        Tier: "micro",
    }
    
    // Mock quota check (success)
    quotaRepo.On("GetTier", mock.Anything, "micro").Return(&Tier{
        CPU: 0.5, RAM: 1073741824, Storage: 10737418240,
    }, nil)
    quotaRepo.On("CheckQuota", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
    
    // Mock VM creation
    vmRepo.On("CreateVM", mock.Anything, mock.Anything).Return(&entity.VM{
        ID:     "vm-123",
        Name:   "test-vm",
        Status: "pending",
    }, nil)
    vmRepo.On("UpdateUsage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 1).Return(nil)
    
    // Act
    vm, err := uc.CreateVM(context.Background(), "user-123", input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test-vm", vm.Name)
    assert.Equal(t, "pending", vm.Status)
    
    // Verify mocks
    quotaRepo.AssertExpectations(t)
    vmRepo.AssertExpectations(t)
}

func TestVMUsecase_CreateVM_QuotaExceeded(t *testing.T) {
    // Arrange
    vmRepo := new(MockVMRepository)
    quotaRepo := new(MockQuotaRepository)
    uc := NewVMUsecase(vmRepo, quotaRepo)
    
    input := CreateVMInput{
        Name: "test-vm",
        Tier: "micro",
    }
    
    // Mock quota check (exceeded)
    quotaRepo.On("GetTier", mock.Anything, "micro").Return(&Tier{}, nil)
    quotaRepo.On("CheckQuota", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ErrQuotaExceeded)
    
    // Act
    vm, err := uc.CreateVM(context.Background(), "user-123", input)
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, vm)
    assert.Equal(t, ErrQuotaExceeded, err)
}
```

---

#### Task 5.2: Integration Test
**Estimate:** 1 hour
**Acceptance Criteria:**
- [ ] Test all VM endpoints
- [ ] Verify backward compatibility

---

## Files Summary

| File | Action | Purpose |
|------|--------|---------|
| `internal/entity/vm.go` | Create | VM entity |
| `internal/entity/user.go` | Create | User entity |
| `internal/entity/quota.go` | Create | Quota entity |
| `internal/repository/vm_repository.go` | Create | VM data access |
| `internal/repository/quota_repository.go` | Create | Quota data access |
| `internal/repository/user_repository.go` | Create | User data access |
| `internal/usecase/vm_usecase.go` | Create | VM business logic |
| `internal/usecase/quota_usecase.go` | Create | Quota business logic |
| `pkg/errors/errors.go` | Create | Common errors |
| `pkg/response/response.go` | Create | Response helpers |
| `handler/vm_handler.go` | Refactor | Thin handler |
| `cmd/main.go` | Update | Dependency injection |
| `internal/usecase/vm_usecase_test.go` | Create | Unit tests |

---

## Acceptance Criteria

- [ ] All 8 success criteria met
- [ ] Build successful (`go build ./...`)
- [ ] All tests passing
- [ ] No global `db` variable
- [ ] Handler < 100 lines
- [ ] Repository per entity
- [ ] Consistent error handling
- [ ] Documentation updated

---

*Plan created: 2026-03-27*
*Ready for Week 1 implementation*

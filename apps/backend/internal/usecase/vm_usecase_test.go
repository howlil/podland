package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
	pkgerrors "github.com/podland/backend/pkg/errors"
)

// MockVMRepository is a mock implementation of VMRepository
type MockVMRepository struct {
	CreateVMFn        func(ctx context.Context, input repository.VMCreateInput) (*entity.VM, error)
	GetVMByIDFn       func(ctx context.Context, id string) (*entity.VM, error)
	GetVMByIDAndUserFn func(ctx context.Context, id, userID string) (*entity.VM, error)
	GetUserVMsFn      func(ctx context.Context, userID string) ([]*entity.VM, error)
	UpdateVMStatusFn  func(ctx context.Context, id, status string) error
	DeleteVMFn        func(ctx context.Context, id string) error
}

func (m *MockVMRepository) CreateVM(ctx context.Context, input repository.VMCreateInput) (*entity.VM, error) {
	if m.CreateVMFn != nil {
		return m.CreateVMFn(ctx, input)
	}
	return nil, errors.New("CreateVM not implemented")
}

func (m *MockVMRepository) GetVMByID(ctx context.Context, id string) (*entity.VM, error) {
	if m.GetVMByIDFn != nil {
		return m.GetVMByIDFn(ctx, id)
	}
	return nil, errors.New("GetVMByID not implemented")
}

func (m *MockVMRepository) GetVMByIDAndUser(ctx context.Context, id, userID string) (*entity.VM, error) {
	if m.GetVMByIDAndUserFn != nil {
		return m.GetVMByIDAndUserFn(ctx, id, userID)
	}
	return nil, errors.New("GetVMByIDAndUser not implemented")
}

func (m *MockVMRepository) GetUserVMs(ctx context.Context, userID string) ([]*entity.VM, error) {
	if m.GetUserVMsFn != nil {
		return m.GetUserVMsFn(ctx, userID)
	}
	return nil, errors.New("GetUserVMs not implemented")
}

func (m *MockVMRepository) UpdateVMStatus(ctx context.Context, id, status string) error {
	if m.UpdateVMStatusFn != nil {
		return m.UpdateVMStatusFn(ctx, id, status)
	}
	return errors.New("UpdateVMStatus not implemented")
}

func (m *MockVMRepository) UpdateVM(ctx context.Context, id string, input repository.VMUpdateInput) error {
	return nil
}

func (m *MockVMRepository) DeleteVM(ctx context.Context, id string) error {
	if m.DeleteVMFn != nil {
		return m.DeleteVMFn(ctx, id)
	}
	return errors.New("DeleteVM not implemented")
}

// MockQuotaRepository is a mock implementation of QuotaRepository
type MockQuotaRepository struct {
	CheckQuotaFn   func(ctx context.Context, userID string, cpu float64, ram, storage int64) error
	UpdateUsageFn  func(ctx context.Context, userID string, cpu float64, ram, storage int64, vmCountDelta int) error
	GetQuotaFn     func(ctx context.Context, userID string) (*entity.Quota, error)
	GetTierFn      func(ctx context.Context, name string) (*entity.Tier, error)
	GetAllTiersFn  func(ctx context.Context) ([]*entity.Tier, error)
}

func (m *MockQuotaRepository) CheckQuota(ctx context.Context, userID string, cpu float64, ram, storage int64) error {
	if m.CheckQuotaFn != nil {
		return m.CheckQuotaFn(ctx, userID, cpu, ram, storage)
	}
	return errors.New("CheckQuota not implemented")
}

func (m *MockQuotaRepository) UpdateUsage(ctx context.Context, userID string, cpu float64, ram, storage int64, vmCountDelta int) error {
	if m.UpdateUsageFn != nil {
		return m.UpdateUsageFn(ctx, userID, cpu, ram, storage, vmCountDelta)
	}
	return errors.New("UpdateUsage not implemented")
}

func (m *MockQuotaRepository) GetQuota(ctx context.Context, userID string) (*entity.Quota, error) {
	if m.GetQuotaFn != nil {
		return m.GetQuotaFn(ctx, userID)
	}
	return nil, errors.New("GetQuota not implemented")
}

func (m *MockQuotaRepository) GetTier(ctx context.Context, name string) (*entity.Tier, error) {
	if m.GetTierFn != nil {
		return m.GetTierFn(ctx, name)
	}
	return nil, errors.New("GetTier not implemented")
}

func (m *MockQuotaRepository) GetAllTiers(ctx context.Context) ([]*entity.Tier, error) {
	if m.GetAllTiersFn != nil {
		return m.GetAllTiersFn(ctx)
	}
	return nil, errors.New("GetAllTiers not implemented")
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	GetUserByIDFn func(ctx context.Context, id string) (*entity.User, error)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, input repository.UserCreateInput) (*entity.User, error) {
	return nil, errors.New("CreateUser not implemented")
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, id)
	}
	return nil, errors.New("GetUserByID not implemented")
}

func (m *MockUserRepository) GetUserByGitHubID(ctx context.Context, githubID string) (*entity.User, error) {
	return nil, errors.New("GetUserByGitHubID not implemented")
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, errors.New("GetUserByEmail not implemented")
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, id string, input repository.UserUpdateInput) error {
	return errors.New("UpdateUser not implemented")
}

func (m *MockUserRepository) UpdateUserNIM(ctx context.Context, userID, nim string) error {
	return errors.New("UpdateUserNIM not implemented")
}

// TestVMUsecase_CreateVM_Success tests the success case of creating a VM
func TestVMUsecase_CreateVM_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	input := CreateVMInput{
		Name: "test-vm",
		OS:   "ubuntu-2204",
		Tier: "micro",
	}

	// Mock user retrieval
	userRepo.GetUserByIDFn = func(ctx context.Context, id string) (*entity.User, error) {
		return &entity.User{
			ID:   "user-123",
			Role: "internal",
		}, nil
	}

	// Mock tier retrieval
	quotaRepo.GetTierFn = func(ctx context.Context, name string) (*entity.Tier, error) {
		return &entity.Tier{
			Name:    "micro",
			CPU:     0.5,
			RAM:     1073741824,
			Storage: 10737418240,
			MinRole: "external",
		}, nil
	}

	// Mock quota check (success)
	quotaRepo.CheckQuotaFn = func(ctx context.Context, userID string, cpu float64, ram, storage int64) error {
		return nil
	}

	// Mock VM creation
	vmRepo.CreateVMFn = func(ctx context.Context, input repository.VMCreateInput) (*entity.VM, error) {
		return &entity.VM{
			ID:        "vm-123",
			UserID:    "user-123",
			Name:      "test-vm",
			OS:        "ubuntu-2204",
			Tier:      "micro",
			CPU:       0.5,
			RAM:       1073741824,
			Storage:   10737418240,
			Status:    "pending",
			CreatedAt: now,
		}, nil
	}

	// Mock quota usage update
	quotaRepo.UpdateUsageFn = func(ctx context.Context, userID string, cpu float64, ram, storage int64, vmCountDelta int) error {
		return nil
	}

	// Act
	vm, err := uc.CreateVM(context.Background(), "user-123", input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if vm == nil {
		t.Fatal("expected VM to be created, got nil")
	}
	if vm.Name != "test-vm" {
		t.Errorf("expected VM name 'test-vm', got %q", vm.Name)
	}
	if vm.Status != "pending" {
		t.Errorf("expected VM status 'pending', got %q", vm.Status)
	}
}

// TestVMUsecase_CreateVM_QuotaExceeded tests the quota exceeded case
func TestVMUsecase_CreateVM_QuotaExceeded(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	input := CreateVMInput{
		Name: "test-vm",
		OS:   "ubuntu-2204",
		Tier: "micro",
	}

	// Mock user retrieval
	userRepo.GetUserByIDFn = func(ctx context.Context, id string) (*entity.User, error) {
		return &entity.User{
			ID:   "user-123",
			Role: "internal",
		}, nil
	}

	// Mock tier retrieval
	quotaRepo.GetTierFn = func(ctx context.Context, name string) (*entity.Tier, error) {
		return &entity.Tier{
			Name:    "micro",
			CPU:     0.5,
			RAM:     1073741824,
			Storage: 10737418240,
			MinRole: "external",
		}, nil
	}

	// Mock quota check (exceeded)
	quotaRepo.CheckQuotaFn = func(ctx context.Context, userID string, cpu float64, ram, storage int64) error {
		return errors.New("quota exceeded: CPU limit")
	}

	// Act
	vm, err := uc.CreateVM(context.Background(), "user-123", input)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if vm != nil {
		t.Errorf("expected nil VM, got %v", vm)
	}
	if !errors.Is(err, pkgerrors.ErrQuotaExceeded) {
		t.Errorf("expected ErrQuotaExceeded, got %v", err)
	}
}

// TestVMUsecase_CreateVM_InvalidInput tests the invalid input case
func TestVMUsecase_CreateVM_InvalidInput(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	testCases := []struct {
		name  string
		input CreateVMInput
	}{
		{"empty name", CreateVMInput{Name: "", OS: "ubuntu-2204", Tier: "micro"}},
		{"empty tier", CreateVMInput{Name: "test-vm", OS: "ubuntu-2204", Tier: ""}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			vm, err := uc.CreateVM(context.Background(), "user-123", tc.input)

			// Assert
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if vm != nil {
				t.Errorf("expected nil VM, got %v", vm)
			}
			if !errors.Is(err, pkgerrors.ErrInvalidRequest) {
				t.Errorf("expected ErrInvalidRequest, got %v", err)
			}
		})
	}
}

// TestVMUsecase_GetVMByID_Success tests the success case of getting a VM by ID
func TestVMUsecase_GetVMByID_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	expectedVM := &entity.VM{
		ID:        "vm-123",
		UserID:    "user-123",
		Name:      "test-vm",
		OS:        "ubuntu-2204",
		Tier:      "micro",
		CPU:       0.5,
		RAM:       1073741824,
		Storage:   10737418240,
		Status:    "running",
		CreatedAt: now,
	}

	// Mock VM retrieval
	vmRepo.GetVMByIDAndUserFn = func(ctx context.Context, id, userID string) (*entity.VM, error) {
		if id != "vm-123" || userID != "user-123" {
			return nil, repository.ErrVMNotFound
		}
		return expectedVM, nil
	}

	// Act
	vm, err := uc.GetVMByID(context.Background(), "vm-123", "user-123")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if vm == nil {
		t.Fatal("expected VM, got nil")
	}
	if vm.ID != "vm-123" {
		t.Errorf("expected VM ID 'vm-123', got %q", vm.ID)
	}
	if vm.Name != "test-vm" {
		t.Errorf("expected VM name 'test-vm', got %q", vm.Name)
	}
}

// TestVMUsecase_GetVMByID_NotFound tests the not found case
func TestVMUsecase_GetVMByID_NotFound(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	// Mock VM retrieval (not found)
	vmRepo.GetVMByIDAndUserFn = func(ctx context.Context, id, userID string) (*entity.VM, error) {
		return nil, repository.ErrVMNotFound
	}

	// Act
	vm, err := uc.GetVMByID(context.Background(), "vm-nonexistent", "user-123")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if vm != nil {
		t.Errorf("expected nil VM, got %v", vm)
	}
	if !errors.Is(err, pkgerrors.ErrVMNotFound) {
		t.Errorf("expected ErrVMNotFound, got %v", err)
	}
}

// TestVMUsecase_ListVMs_Success tests the success case of listing VMs
func TestVMUsecase_ListVMs_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	expectedVMs := []*entity.VM{
		{
			ID:        "vm-123",
			UserID:    "user-123",
			Name:      "test-vm-1",
			OS:        "ubuntu-2204",
			Tier:      "micro",
			CPU:       0.5,
			RAM:       1073741824,
			Storage:   10737418240,
			Status:    "running",
			CreatedAt: now,
		},
		{
			ID:        "vm-456",
			UserID:    "user-123",
			Name:      "test-vm-2",
			OS:        "debian-12",
			Tier:      "small",
			CPU:       1.0,
			RAM:       2147483648,
			Storage:   21474836480,
			Status:    "stopped",
			CreatedAt: now,
		},
	}

	// Mock VM listing
	vmRepo.GetUserVMsFn = func(ctx context.Context, userID string) ([]*entity.VM, error) {
		if userID != "user-123" {
			return nil, errors.New("unexpected user ID")
		}
		return expectedVMs, nil
	}

	// Act
	vms, err := uc.ListVMs(context.Background(), "user-123")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(vms) != 2 {
		t.Fatalf("expected 2 VMs, got %d", len(vms))
	}
	if vms[0].Name != "test-vm-1" {
		t.Errorf("expected first VM name 'test-vm-1', got %q", vms[0].Name)
	}
	if vms[1].Name != "test-vm-2" {
		t.Errorf("expected second VM name 'test-vm-2', got %q", vms[1].Name)
	}
}

// TestVMUsecase_StartVM_Success tests the success case of starting a VM
func TestVMUsecase_StartVM_Success(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	// Mock VM retrieval (stopped VM)
	vmRepo.GetVMByIDAndUserFn = func(ctx context.Context, id, userID string) (*entity.VM, error) {
		return &entity.VM{
			ID:     "vm-123",
			UserID: "user-123",
			Name:   "test-vm",
			Status: "stopped",
		}, nil
	}

	// Mock status update
	vmRepo.UpdateVMStatusFn = func(ctx context.Context, id, status string) error {
		if status != "pending" {
			t.Errorf("expected status 'pending', got %q", status)
		}
		return nil
	}

	// Act
	err := uc.StartVM(context.Background(), "vm-123", "user-123")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestVMUsecase_StartVM_NotStopped tests the case when VM is not stopped
func TestVMUsecase_StartVM_NotStopped(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	// Mock VM retrieval (running VM)
	vmRepo.GetVMByIDAndUserFn = func(ctx context.Context, id, userID string) (*entity.VM, error) {
		return &entity.VM{
			ID:     "vm-123",
			UserID: "user-123",
			Name:   "test-vm",
			Status: "running",
		}, nil
	}

	// Act
	err := uc.StartVM(context.Background(), "vm-123", "user-123")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pkgerrors.ErrVMNotStopped) {
		t.Errorf("expected ErrVMNotStopped, got %v", err)
	}
}

// TestVMUsecase_StopVM_Success tests the success case of stopping a VM
func TestVMUsecase_StopVM_Success(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	// Mock VM retrieval (running VM)
	vmRepo.GetVMByIDAndUserFn = func(ctx context.Context, id, userID string) (*entity.VM, error) {
		return &entity.VM{
			ID:     "vm-123",
			UserID: "user-123",
			Name:   "test-vm",
			Status: "running",
		}, nil
	}

	// Mock status update
	vmRepo.UpdateVMStatusFn = func(ctx context.Context, id, status string) error {
		return nil
	}

	// Act
	err := uc.StopVM(context.Background(), "vm-123", "user-123")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestVMUsecase_StopVM_NotRunning tests the case when VM is not running
func TestVMUsecase_StopVM_NotRunning(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	// Mock VM retrieval (stopped VM)
	vmRepo.GetVMByIDAndUserFn = func(ctx context.Context, id, userID string) (*entity.VM, error) {
		return &entity.VM{
			ID:     "vm-123",
			UserID: "user-123",
			Name:   "test-vm",
			Status: "stopped",
		}, nil
	}

	// Act
	err := uc.StopVM(context.Background(), "vm-123", "user-123")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pkgerrors.ErrVMNotRunning) {
		t.Errorf("expected ErrVMNotRunning, got %v", err)
	}
}

// TestVMUsecase_RestartVM_Success tests the success case of restarting a VM
func TestVMUsecase_RestartVM_Success(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	// Mock VM retrieval (running VM)
	vmRepo.GetVMByIDAndUserFn = func(ctx context.Context, id, userID string) (*entity.VM, error) {
		return &entity.VM{
			ID:     "vm-123",
			UserID: "user-123",
			Name:   "test-vm",
			Status: "running",
		}, nil
	}

	// Mock status update
	vmRepo.UpdateVMStatusFn = func(ctx context.Context, id, status string) error {
		return nil
	}

	// Act
	err := uc.RestartVM(context.Background(), "vm-123", "user-123")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestVMUsecase_DeleteVM_Success tests the success case of deleting a VM
func TestVMUsecase_DeleteVM_Success(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	// Mock VM retrieval
	vmRepo.GetVMByIDAndUserFn = func(ctx context.Context, id, userID string) (*entity.VM, error) {
		return &entity.VM{
			ID:      "vm-123",
			UserID:  "user-123",
			Name:    "test-vm",
			Status:  "stopped",
			CPU:     0.5,
			RAM:     1073741824,
			Storage: 10737418240,
		}, nil
	}

	// Mock quota usage update
	quotaRepo.UpdateUsageFn = func(ctx context.Context, userID string, cpu float64, ram, storage int64, vmCountDelta int) error {
		return nil
	}

	// Mock VM deletion
	vmRepo.DeleteVMFn = func(ctx context.Context, id string) error {
		return nil
	}

	// Act
	err := uc.DeleteVM(context.Background(), "vm-123", "user-123")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestVMUsecase_DeleteVM_NotFound tests the not found case for deletion
func TestVMUsecase_DeleteVM_NotFound(t *testing.T) {
	// Arrange
	vmRepo := &MockVMRepository{}
	quotaRepo := &MockQuotaRepository{}
	userRepo := &MockUserRepository{}

	uc := NewVMUsecase(vmRepo, quotaRepo, userRepo)

	// Mock VM retrieval (not found)
	vmRepo.GetVMByIDAndUserFn = func(ctx context.Context, id, userID string) (*entity.VM, error) {
		return nil, repository.ErrVMNotFound
	}

	// Act
	err := uc.DeleteVM(context.Background(), "vm-nonexistent", "user-123")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pkgerrors.ErrVMNotFound) {
		t.Errorf("expected ErrVMNotFound, got %v", err)
	}
}

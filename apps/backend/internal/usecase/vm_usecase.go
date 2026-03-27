package usecase

import (
	"context"

	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
	pkgerrors "github.com/podland/backend/pkg/errors"
)

// VMUsecase defines VM business logic
type VMUsecase struct {
	vmRepo    repository.VMRepository
	quotaRepo repository.QuotaRepository
	userRepo  repository.UserRepository
}

// NewVMUsecase creates a new VM usecase with dependencies
func NewVMUsecase(vmRepo repository.VMRepository, quotaRepo repository.QuotaRepository, userRepo repository.UserRepository) *VMUsecase {
	return &VMUsecase{
		vmRepo:    vmRepo,
		quotaRepo: quotaRepo,
		userRepo:  userRepo,
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
		return nil, pkgerrors.ErrInvalidRequest
	}
	if input.Tier == "" {
		return nil, pkgerrors.ErrInvalidRequest
	}

	// 2. Get user to check role
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to get user")
	}

	// 3. Get tier configuration
	tier, err := uc.quotaRepo.GetTier(ctx, input.Tier)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to get tier")
	}

	// 4. Check tier availability by role
	if !tier.IsValidForRole(user.Role) {
		return nil, pkgerrors.ErrTierNotAvailable
	}

	// 5. Check quota
	if err := uc.quotaRepo.CheckQuota(ctx, userID, tier.CPU, tier.RAM, tier.Storage); err != nil {
		return nil, pkgerrors.ErrQuotaExceeded
	}

	// 6. Create VM in database
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
		return nil, pkgerrors.Wrap(err, "failed to create VM")
	}

	// 7. Update quota usage
	if err := uc.quotaRepo.UpdateUsage(ctx, userID, tier.CPU, tier.RAM, tier.Storage, 1); err != nil {
		return nil, pkgerrors.Wrap(err, "failed to update quota")
	}

	return vm, nil
}

// GetVMByID gets a VM by ID with ownership check
func (uc *VMUsecase) GetVMByID(ctx context.Context, vmID, userID string) (*entity.VM, error) {
	vm, err := uc.vmRepo.GetVMByIDAndUser(ctx, vmID, userID)
	if err != nil {
		if err == repository.ErrVMNotFound {
			return nil, pkgerrors.ErrVMNotFound
		}
		return nil, pkgerrors.Wrap(err, "failed to get VM")
	}

	return vm, nil
}

// ListVMs lists all VMs for a user
func (uc *VMUsecase) ListVMs(ctx context.Context, userID string) ([]*entity.VM, error) {
	vms, err := uc.vmRepo.GetUserVMs(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to list VMs")
	}

	return vms, nil
}

// StartVM starts a stopped VM
func (uc *VMUsecase) StartVM(ctx context.Context, vmID, userID string) error {
	vm, err := uc.vmRepo.GetVMByIDAndUser(ctx, vmID, userID)
	if err != nil {
		return pkgerrors.ErrVMNotFound
	}

	if !vm.CanStart() {
		return pkgerrors.ErrVMNotStopped
	}

	if err := uc.vmRepo.UpdateVMStatus(ctx, vmID, "pending"); err != nil {
		return pkgerrors.Wrap(err, "failed to update VM status")
	}

	// k8s integration will be called in handler (async)
	return nil
}

// StopVM stops a running VM
func (uc *VMUsecase) StopVM(ctx context.Context, vmID, userID string) error {
	vm, err := uc.vmRepo.GetVMByIDAndUser(ctx, vmID, userID)
	if err != nil {
		return pkgerrors.ErrVMNotFound
	}

	if !vm.CanStop() {
		return pkgerrors.ErrVMNotRunning
	}

	if err := uc.vmRepo.UpdateVMStatus(ctx, vmID, "stopped"); err != nil {
		return pkgerrors.Wrap(err, "failed to update VM status")
	}

	// k8s integration will be called in handler (async)
	return nil
}

// RestartVM restarts a running VM
func (uc *VMUsecase) RestartVM(ctx context.Context, vmID, userID string) error {
	vm, err := uc.vmRepo.GetVMByIDAndUser(ctx, vmID, userID)
	if err != nil {
		return pkgerrors.ErrVMNotFound
	}

	if !vm.CanRestart() {
		return pkgerrors.ErrVMNotRunning
	}

	if err := uc.vmRepo.UpdateVMStatus(ctx, vmID, "pending"); err != nil {
		return pkgerrors.Wrap(err, "failed to update VM status")
	}

	// k8s integration will be called in handler (async)
	return nil
}

// DeleteVM deletes a VM
func (uc *VMUsecase) DeleteVM(ctx context.Context, vmID, userID string) error {
	vm, err := uc.vmRepo.GetVMByIDAndUser(ctx, vmID, userID)
	if err != nil {
		return pkgerrors.ErrVMNotFound
	}

	// Get quota info for updating usage
	// Note: We need to get the tier to know the resources to subtract
	// For now, we use the VM's CPU, RAM, Storage directly

	// Update quota usage (decrease)
	if err := uc.quotaRepo.UpdateUsage(ctx, userID, -vm.CPU, -vm.RAM, -vm.Storage, -1); err != nil {
		// Log error but continue with delete
		// In production, use proper logging
	}

	// Delete VM (soft delete)
	if err := uc.vmRepo.DeleteVM(ctx, vmID); err != nil {
		return pkgerrors.Wrap(err, "failed to delete VM")
	}

	return nil
}

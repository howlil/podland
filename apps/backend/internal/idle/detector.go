package idle

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
)

// Detector detects idle VMs based on last update time
type Detector struct {
	vmRepo           repository.VMRepository
	userRepo         repository.UserRepository
	notificationRepo repository.NotificationRepository
	emailService     EmailService
}

// NewDetector creates a new idle detector
func NewDetector(
	vmRepo repository.VMRepository,
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepository,
	emailService EmailService,
) *Detector {
	return &Detector{
		vmRepo:           vmRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		emailService:     emailService,
	}
}

// Run executes idle detection
func (d *Detector) Run() {
	ctx := context.Background()
	log.Println("Running idle VM detection...")

	// Get VMs idle for 48 hours (no updates)
	idleVMs, err := d.vmRepo.GetIdleVMs(ctx, 48)
	if err != nil {
		log.Printf("Error getting idle VMs: %v", err)
		return
	}

	for _, vm := range idleVMs {
		// Skip pinned VMs
		if vm.IsPinned {
			continue
		}

		// Get user for notification
		user, err := d.userRepo.GetUserByID(ctx, vm.UserID)
		if err != nil {
			log.Printf("Error getting user for VM %s: %v", vm.ID, err)
			continue
		}

		// Check if already warned
		if vm.IdleWarnedAt != nil && time.Since(*vm.IdleWarnedAt) > 24*time.Hour {
			// Delete VM after 24h warning period
			log.Printf("Deleting idle VM %s (user %s) after grace period", vm.ID, user.Email)
			d.deleteVM(vm)
		} else if vm.IdleWarnedAt == nil {
			// Send warning for first-time idle VMs
			log.Printf("Sending idle warning for VM %s (user %s)", vm.ID, user.Email)
			d.sendWarning(vm, user)
		}
	}
}

func (d *Detector) sendWarning(vm *entity.VM, user *entity.User) {
	ctx := context.Background()

	// Send email notification (check if email service is configured by checking if client is non-nil)
	err := d.emailService.SendIdleWarning(user.Email, user.DisplayName, vm.Name, vm.ID)
	if err != nil {
		log.Printf("Error sending idle warning email: %v", err)
	}

	// Create in-app notification
	userID, _ := uuid.Parse(user.ID)
	vmID, _ := uuid.Parse(vm.ID)
	notification := entity.NewNotification(
		userID,
		vmID,
		"idle_warning",
		"warning",
		"VM Idle Warning",
		fmt.Sprintf(
			"Your VM '%s' has been idle for 48 hours. It will be deleted in 24 hours if no activity is detected.",
			vm.Name,
		),
	)
	if err := d.notificationRepo.Create(ctx, notification); err != nil {
		log.Printf("Error creating notification: %v", err)
	}

	// Update VM with warned_at timestamp
	now := time.Now()
	if err := d.vmRepo.SetIdleWarnedAt(ctx, vm.ID, now); err != nil {
		log.Printf("Error setting idle_warned_at for VM %s: %v", vm.ID, err)
	}
}

func (d *Detector) deleteVM(vm *entity.VM) {
	ctx := context.Background()

	// Soft delete the VM
	if err := d.vmRepo.DeleteVM(ctx, vm.ID); err != nil {
		log.Printf("Error deleting VM %s: %v", vm.ID, err)
		return
	}

	// Create notification
	user, _ := d.userRepo.GetUserByID(ctx, vm.UserID)
	if user != nil {
		userID, _ := uuid.Parse(user.ID)
		vmID, _ := uuid.Parse(vm.ID)
		notification := entity.NewNotification(
			userID,
			vmID,
			"idle_deletion",
			"info",
			"VM Deleted",
			fmt.Sprintf("Your VM '%s' has been deleted due to inactivity.", vm.Name),
		)
		d.notificationRepo.Create(ctx, notification)
	}

	log.Printf("VM %s deleted successfully", vm.ID)
}

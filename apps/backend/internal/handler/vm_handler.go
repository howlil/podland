package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/podland/backend/internal/auth"
	"github.com/podland/backend/internal/cloudflare"
	"github.com/podland/backend/internal/domain"
	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
	"github.com/podland/backend/internal/usecase"
	"github.com/podland/backend/pkg/response"
)

// VMHandler handles VM HTTP requests
type VMHandler struct {
	vmUsecase  *usecase.VMUsecase
	vmRepo     repository.VMRepository
	userRepo   repository.UserRepository
	dnsManager *cloudflare.DNSManager
	dnsPoller  *domain.DNSPoller
	authHelper *AuthHelper
}

// NewVMHandler creates a new VM handler with dependencies
func NewVMHandler(vmUsecase *usecase.VMUsecase, vmRepo repository.VMRepository, userRepo repository.UserRepository, dnsManager *cloudflare.DNSManager, dnsPoller *domain.DNSPoller) *VMHandler {
	return &VMHandler{
		vmUsecase:  vmUsecase,
		vmRepo:     vmRepo,
		userRepo:   userRepo,
		dnsManager: dnsManager,
		dnsPoller:  dnsPoller,
		authHelper: NewAuthHelper(),
	}
}

// CreateVMRequest represents the request body for creating a VM
type CreateVMRequest struct {
	Name string `json:"name"`
	OS   string `json:"os"`
	Tier string `json:"tier"`
}

// CreateVMResponse represents the response for VM creation
type CreateVMResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	SSHKey  string `json:"ssh_key,omitempty"`
	Message string `json:"message,omitempty"`
}

// VMResponse represents a VM in responses
type VMResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OS        string    `json:"os"`
	Tier      string    `json:"tier"`
	CPU       float64   `json:"cpu"`
	RAM       int64     `json:"ram"`
	Storage   int64     `json:"storage"`
	Status    string    `json:"status"`
	Domain    string    `json:"domain,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// HandleCreateVM creates a new VM
// POST /api/vms
func (h *VMHandler) HandleCreateVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	var req CreateVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgresponse.BadRequest(w, "Invalid request body")
		return
	}

	// Validate input
	if req.Name == "" {
		pkgresponse.BadRequest(w, "VM name is required")
		return
	}
	if req.OS == "" {
		req.OS = "ubuntu-2204"
	}
	if req.Tier == "" {
		pkgresponse.BadRequest(w, "Tier is required")
		return
	}

	// Validate OS
	if req.OS != "ubuntu-2204" && req.OS != "debian-12" {
		pkgresponse.BadRequest(w, "Invalid OS. Must be 'ubuntu-2204' or 'debian-12'")
		return
	}

	// Generate SSH keypair
	privateKey, publicKey, err := auth.GenerateKeyPair()
	if err != nil {
		pkgresponse.InternalError(w, "Failed to generate SSH key")
		return
	}

	// Call usecase (business logic)
	vm, err := h.vmUsecase.CreateVM(r.Context(), userID, usecase.CreateVMInput{
		Name:         req.Name,
		OS:           req.OS,
		Tier:         req.Tier,
		SSHPublicKey: publicKey,
	})
	if err != nil {
		h.handleCreateVMError(w, err)
		return
	}

	// Create DNS record if DNS manager is configured
	if h.dnsManager != nil && h.dnsPoller != nil {
		// Generate subdomain from VM name
		subdomain := sanitizeSubdomain(req.Name)
		domain := subdomain + ".podland.app"

		// Handle collisions by appending user ID suffix
		if existing, _ := h.vmRepo.GetVMByIDAndUser(r.Context(), vm.ID, userID); existing != nil && existing.Domain != nil {
			// Check if domain already exists
			if _, err := h.dnsManager.GetRecordByName(r.Context(), domain); err == nil {
				// Domain collision - append user ID
				subdomain = sanitizeSubdomain(req.Name) + "-u" + userID[:8]
				domain = subdomain + ".podland.app"
			}
		}

		// Create DNS record
		_, err = h.dnsManager.CreateCNAME(r.Context(), domain, "tunnel.podland.app")
		if err != nil {
			// Log error but don't fail VM creation
			log.Printf("WARNING: Failed to create DNS record for VM %s: %v", vm.ID, err)
		} else {
			// Update VM with domain
			if err := h.vmRepo.UpdateVM(r.Context(), vm.ID, repository.VMUpdateInput{
				Domain:       &domain,
				DomainStatus: stringPtr("pending"),
			}); err != nil {
				log.Printf("WARNING: Failed to update VM with domain: %v", err)
			}

			// Start background DNS poller (async)
			h.dnsPoller.StartDNSPoller(vm.ID, domain)
		}

		// Refresh VM entity with domain
		vm, _ = h.vmRepo.GetVMByID(r.Context(), vm.ID)
	}

	// Return 202 Accepted
	pkgresponse.Accepted(w, CreateVMResponse{
		ID:      vm.ID,
		Status:  vm.Status,
		SSHKey:  privateKey,
		Message: "VM is being created. SSH key shown only once - download now!",
	})
}

func (h *VMHandler) handleCreateVMError(w http.ResponseWriter, err error) {
	switch err {
	case usecase.ErrInvalidRequest:
		pkgresponse.BadRequest(w, "Invalid request")
	case usecase.ErrTierNotAvailable:
		pkgresponse.Forbidden(w, "Tier not available for your role")
	case usecase.ErrQuotaExceeded:
		pkgresponse.Forbidden(w, "Quota exceeded")
	default:
		pkgresponse.InternalError(w, "Failed to create VM")
	}
}

// HandleListVMs lists all VMs for the authenticated user
// GET /api/vms
func (h *VMHandler) HandleListVMs(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	vms, err := h.vmUsecase.ListVMs(r.Context(), userID)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to list VMs")
		return
	}

	pkgresponse.Success(w, http.StatusOK, h.toVMResponse(vms))
}

// HandleGetVM gets details for a specific VM
// GET /api/vms/{id}
func (h *VMHandler) HandleGetVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	vm, err := h.vmUsecase.GetVMByID(r.Context(), vmID, userID)
	if err != nil {
		if err == usecase.ErrVMNotFound {
			pkgresponse.NotFound(w, "VM not found")
			return
		}
		pkgresponse.InternalError(w, "Failed to get VM")
		return
	}

	pkgresponse.Success(w, http.StatusOK, h.toVMResponseSingle(vm))
}

// HandleStartVM starts a stopped VM
// POST /api/vms/{id}/start
func (h *VMHandler) HandleStartVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	if err := h.vmUsecase.StartVM(r.Context(), vmID, userID); err != nil {
		h.handleStartVMError(w, err)
		return
	}

	pkgresponse.Accepted(w, map[string]string{"status": "pending"})
}

// HandleStopVM stops a running VM
// POST /api/vms/{id}/stop
func (h *VMHandler) HandleStopVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	if err := h.vmUsecase.StopVM(r.Context(), vmID, userID); err != nil {
		h.handleStopVMError(w, err)
		return
	}

	pkgresponse.Accepted(w, map[string]string{"status": "stopped"})
}

// HandleRestartVM restarts a running VM
// POST /api/vms/{id}/restart
func (h *VMHandler) HandleRestartVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	if err := h.vmUsecase.RestartVM(r.Context(), vmID, userID); err != nil {
		h.handleRestartVMError(w, err)
		return
	}

	pkgresponse.Accepted(w, map[string]string{"status": "pending"})
}

// HandleDeleteVM deletes a VM
// DELETE /api/vms/{id}
func (h *VMHandler) HandleDeleteVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	// Get VM to retrieve domain (before deletion)
	var domainToDelete string
	if h.dnsManager != nil {
		vm, err := h.vmRepo.GetVMByID(r.Context(), vmID)
		if err == nil && vm != nil && vm.Domain != nil {
			domainToDelete = *vm.Domain
		}
	}

	// Delete DNS record first (if domain exists)
	if h.dnsManager != nil && domainToDelete != "" {
		dnsRecord, err := h.dnsManager.GetRecordByName(r.Context(), domainToDelete)
		if err == nil && dnsRecord != nil {
			if err := h.dnsManager.DeleteRecord(r.Context(), dnsRecord.ID); err != nil {
				log.Printf("WARNING: Failed to delete DNS record for VM %s: %v", vmID, err)
			}
		}
	}

	// Delete VM in usecase (which releases quota and soft-deletes)
	if err := h.vmUsecase.DeleteVM(r.Context(), vmID, userID); err != nil {
		if err == usecase.ErrVMNotFound {
			pkgresponse.NotFound(w, "VM not found")
			return
		}
		pkgresponse.InternalError(w, "Failed to delete VM")
		return
	}

	pkgresponse.NoContent(w)
}

// Error handlers

func (h *VMHandler) handleStartVMError(w http.ResponseWriter, err error) {
	if err == usecase.ErrVMNotFound {
		pkgresponse.NotFound(w, "VM not found")
		return
	}
	if err == usecase.ErrVMNotStopped {
		pkgresponse.BadRequest(w, "VM must be stopped to start")
		return
	}
	pkgresponse.InternalError(w, "Failed to start VM")
}

func (h *VMHandler) handleStopVMError(w http.ResponseWriter, err error) {
	if err == usecase.ErrVMNotFound {
		pkgresponse.NotFound(w, "VM not found")
		return
	}
	if err == usecase.ErrVMNotRunning {
		pkgresponse.BadRequest(w, "VM must be running to stop")
		return
	}
	pkgresponse.InternalError(w, "Failed to stop VM")
}

func (h *VMHandler) handleRestartVMError(w http.ResponseWriter, err error) {
	if err == usecase.ErrVMNotFound {
		pkgresponse.NotFound(w, "VM not found")
		return
	}
	if err == usecase.ErrVMNotRunning {
		pkgresponse.BadRequest(w, "VM must be running to restart")
		return
	}
	pkgresponse.InternalError(w, "Failed to restart VM")
}

func (h *VMHandler) toVMResponse(vms []*entity.VM) []VMResponse {
	response := make([]VMResponse, len(vms))
	for i, vm := range vms {
		response[i] = VMResponse{
			ID:        vm.ID,
			Name:      vm.Name,
			OS:        vm.OS,
			Tier:      vm.Tier,
			CPU:       vm.CPU,
			RAM:       vm.RAM,
			Storage:   vm.Storage,
			Status:    vm.Status,
			Domain:    getString(vm.Domain),
			CreatedAt: vm.CreatedAt,
		}
	}
	return response
}

func (h *VMHandler) toVMResponseSingle(vm *entity.VM) VMResponse {
	return VMResponse{
		ID:        vm.ID,
		Name:      vm.Name,
		OS:        vm.OS,
		Tier:      vm.Tier,
		CPU:       vm.CPU,
		RAM:       vm.RAM,
		Storage:   vm.Storage,
		Status:    vm.Status,
		Domain:    getString(vm.Domain),
		CreatedAt: vm.CreatedAt,
	}
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// sanitizeSubdomain converts a VM name to a valid DNS subdomain
func sanitizeSubdomain(name string) string {
	// Lowercase
	s := strings.ToLower(name)
	// Replace spaces and special chars with hyphens
	s = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(s, "-")
	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")
	// Limit to 63 chars (DNS spec)
	if len(s) > 63 {
		s = s[:63]
	}
	// Ensure not empty
	if s == "" {
		s = "vm"
	}
	return s
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// HandlePinVM pins a VM to prevent auto-deletion
// POST /api/vms/{id}/pin
func (h *VMHandler) HandlePinVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	// Verify VM ownership
	_, err := h.vmRepo.GetVMByIDAndUser(r.Context(), vmID, userID)
	if err != nil {
		pkgresponse.NotFound(w, "VM not found")
		return
	}

	// Get user to check pin limit
	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to get user")
		return
	}

	// Check pin limit based on role
	limit := 1
	if user.Role == "internal" {
		limit = 3
	}

	pinnedCount := h.vmRepo.GetPinnedCount(r.Context(), userID)
	if pinnedCount >= limit {
		pkgresponse.BadRequest(w, fmt.Sprintf("Pin limit exceeded (max %d for your role)", limit))
		return
	}

	if err := h.vmRepo.SetPinned(r.Context(), vmID, true); err != nil {
		pkgresponse.InternalError(w, "Failed to pin VM")
		return
	}

	pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleUnpinVM unpins a VM
// DELETE /api/vms/{id}/pin
func (h *VMHandler) HandleUnpinVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	if err := h.vmRepo.SetPinned(r.Context(), vmID, false); err != nil {
		pkgresponse.InternalError(w, "Failed to unpin VM")
		return
	}

	pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

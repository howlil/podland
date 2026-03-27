package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	authmw "github.com/podland/backend/handler/middleware"
	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/usecase"
	sshkey "github.com/podland/backend/internal/ssh"
	"github.com/podland/backend/pkg/response"
)

// VMHandler handles VM HTTP requests
type VMHandler struct {
	vmUsecase  *usecase.VMUsecase
	authHelper *authmw.AuthHelper
}

// NewVMHandler creates a new VM handler with dependencies
func NewVMHandler(vmUsecase *usecase.VMUsecase) *VMHandler {
	return &VMHandler{
		vmUsecase:  vmUsecase,
		authHelper: authmw.NewAuthHelper(),
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
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var req CreateVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate input
	if req.Name == "" {
		response.BadRequest(w, "VM name is required")
		return
	}
	if req.OS == "" {
		req.OS = "ubuntu-2204"
	}
	if req.Tier == "" {
		response.BadRequest(w, "Tier is required")
		return
	}

	// Validate OS
	if req.OS != "ubuntu-2204" && req.OS != "debian-12" {
		response.BadRequest(w, "Invalid OS. Must be 'ubuntu-2204' or 'debian-12'")
		return
	}

	// Generate SSH keypair
	privateKey, publicKey, err := sshkey.GenerateKeyPair()
	if err != nil {
		response.InternalError(w, "Failed to generate SSH key")
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

	// Return 202 Accepted
	response.Accepted(w, CreateVMResponse{
		ID:      vm.ID,
		Status:  vm.Status,
		SSHKey:  privateKey,
		Message: "VM is being created. SSH key shown only once - download now!",
	})
}

func (h *VMHandler) handleCreateVMError(w http.ResponseWriter, err error) {
	switch err {
	case usecase.ErrInvalidRequest:
		response.BadRequest(w, "Invalid request")
	case usecase.ErrTierNotAvailable:
		response.Forbidden(w, "Tier not available for your role")
	case usecase.ErrQuotaExceeded:
		response.Forbidden(w, "Quota exceeded")
	default:
		response.InternalError(w, "Failed to create VM")
	}
}

// HandleListVMs lists all VMs for the authenticated user
// GET /api/vms
func (h *VMHandler) HandleListVMs(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	vms, err := h.vmUsecase.ListVMs(r.Context(), userID)
	if err != nil {
		response.InternalError(w, "Failed to list VMs")
		return
	}

	response.Success(w, http.StatusOK, h.toVMResponse(vms))
}

// HandleGetVM gets details for a specific VM
// GET /api/vms/{id}
func (h *VMHandler) HandleGetVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		response.BadRequest(w, "VM ID is required")
		return
	}

	vm, err := h.vmUsecase.GetVMByID(r.Context(), vmID, userID)
	if err != nil {
		if err == usecase.ErrVMNotFound {
			response.NotFound(w, "VM not found")
			return
		}
		response.InternalError(w, "Failed to get VM")
		return
	}

	response.Success(w, http.StatusOK, h.toVMResponseSingle(vm))
}

// HandleStartVM starts a stopped VM
// POST /api/vms/{id}/start
func (h *VMHandler) HandleStartVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		response.BadRequest(w, "VM ID is required")
		return
	}

	if err := h.vmUsecase.StartVM(r.Context(), vmID, userID); err != nil {
		h.handleStartVMError(w, err)
		return
	}

	response.Accepted(w, map[string]string{"status": "pending"})
}

// HandleStopVM stops a running VM
// POST /api/vms/{id}/stop
func (h *VMHandler) HandleStopVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		response.BadRequest(w, "VM ID is required")
		return
	}

	if err := h.vmUsecase.StopVM(r.Context(), vmID, userID); err != nil {
		h.handleStopVMError(w, err)
		return
	}

	response.Accepted(w, map[string]string{"status": "stopped"})
}

// HandleRestartVM restarts a running VM
// POST /api/vms/{id}/restart
func (h *VMHandler) HandleRestartVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		response.BadRequest(w, "VM ID is required")
		return
	}

	if err := h.vmUsecase.RestartVM(r.Context(), vmID, userID); err != nil {
		h.handleRestartVMError(w, err)
		return
	}

	response.Accepted(w, map[string]string{"status": "pending"})
}

// HandleDeleteVM deletes a VM
// DELETE /api/vms/{id}
func (h *VMHandler) HandleDeleteVM(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		response.BadRequest(w, "VM ID is required")
		return
	}

	if err := h.vmUsecase.DeleteVM(r.Context(), vmID, userID); err != nil {
		if err == usecase.ErrVMNotFound {
			response.NotFound(w, "VM not found")
			return
		}
		response.InternalError(w, "Failed to delete VM")
		return
	}

	response.NoContent(w)
}

// Error handlers

func (h *VMHandler) handleStartVMError(w http.ResponseWriter, err error) {
	if err == usecase.ErrVMNotFound {
		response.NotFound(w, "VM not found")
		return
	}
	if err == usecase.ErrVMNotStopped {
		response.BadRequest(w, "VM must be stopped to start")
		return
	}
	response.InternalError(w, "Failed to start VM")
}

func (h *VMHandler) handleStopVMError(w http.ResponseWriter, err error) {
	if err == usecase.ErrVMNotFound {
		response.NotFound(w, "VM not found")
		return
	}
	if err == usecase.ErrVMNotRunning {
		response.BadRequest(w, "VM must be running to stop")
		return
	}
	response.InternalError(w, "Failed to stop VM")
}

func (h *VMHandler) handleRestartVMError(w http.ResponseWriter, err error) {
	if err == usecase.ErrVMNotFound {
		response.NotFound(w, "VM not found")
		return
	}
	if err == usecase.ErrVMNotRunning {
		response.BadRequest(w, "VM must be running to restart")
		return
	}
	response.InternalError(w, "Failed to restart VM")
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
			Domain:    vm.Domain,
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
		Domain:    vm.Domain,
		CreatedAt: vm.CreatedAt,
	}
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

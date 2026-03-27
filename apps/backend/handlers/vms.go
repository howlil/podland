package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/podland/backend/internal/database"
	sshkey "github.com/podland/backend/internal/ssh"
)

// CreateVMRequest represents the request body for creating a VM
type CreateVMRequest struct {
	Name string `json:"name"`
	OS   string `json:"os"`   // "ubuntu-2204" or "debian-12"
	Tier string `json:"tier"` // "nano", "micro", "small", "medium", "large", "xlarge"
}

// VMResponse represents the response for VM operations
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

// CreateVMResponse represents the response for VM creation
type CreateVMResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	SSHKey    string `json:"ssh_key,omitempty"`
	Message   string `json:"message,omitempty"`
}

// HandleCreateVM creates a new VM
// POST /api/vms
func HandleCreateVM(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "VM name is required", http.StatusBadRequest)
		return
	}
	if req.OS == "" {
		req.OS = "ubuntu-2204"
	}
	if req.Tier == "" {
		http.Error(w, "Tier is required", http.StatusBadRequest)
		return
	}

	// Validate OS
	if req.OS != "ubuntu-2204" && req.OS != "debian-12" {
		http.Error(w, "Invalid OS. Must be 'ubuntu-2204' or 'debian-12'", http.StatusBadRequest)
		return
	}

	dbWrapper := database.NewDB(db)

	// Validate tier and get tier configuration
	tier, err := dbWrapper.GetTier(r.Context(), req.Tier)
	if err != nil {
		http.Error(w, "Invalid tier", http.StatusBadRequest)
		return
	}

	// Get user role to check tier availability
	user, err := dbWrapper.GetUserByID(userIDStr)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check tier availability by role
	if tier.MinRole == "internal" && user.Role != "internal" {
		http.Error(w, "Tier not available for your role", http.StatusForbidden)
		return
	}

	// Check quota
	err = dbWrapper.CheckQuota(r.Context(), userIDStr, tier.CPU, tier.RAM, tier.Storage)
	if err != nil {
		if err == database.ErrQuotaExceeded {
			http.Error(w, "Quota exceeded", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to check quota", http.StatusInternalServerError)
		return
	}

	// Generate SSH keypair
	privateKey, publicKey, err := sshkey.GenerateKeyPair()
	if err != nil {
		http.Error(w, "Failed to generate SSH key", http.StatusInternalServerError)
		return
	}

	// Create VM in database
	vmInput := database.VMCreateInput{
		UserID:       userIDStr,
		Name:         req.Name,
		OS:           req.OS,
		Tier:         req.Tier,
		CPU:          tier.CPU,
		RAM:          tier.RAM,
		Storage:      tier.Storage,
		SSHPublicKey: publicKey,
	}

	vm, err := dbWrapper.CreateVM(vmInput)
	if err != nil {
		http.Error(w, "Failed to create VM", http.StatusInternalServerError)
		return
	}

	// Update quota usage
	err = dbWrapper.UpdateUsage(r.Context(), userIDStr, tier.CPU, tier.RAM, tier.Storage, 1)
	if err != nil {
		http.Error(w, "Failed to update quota", http.StatusInternalServerError)
		return
	}

	// Create k8s resources asynchronously
	go func() {
		// TODO: Initialize vmManager and call CreateVM
		// For now, just update status to running after a delay
		time.Sleep(2 * time.Second)
		dbWrapper.UpdateVMStatus(vm.ID, "running")
	}()

	// Return 202 Accepted
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(CreateVMResponse{
		ID:      vm.ID,
		Status:  vm.Status,
		SSHKey:  privateKey,
		Message: "VM is being created. SSH key shown only once - download now!",
	})
}

// HandleListVMs lists all VMs for the authenticated user
// GET /api/vms
func HandleListVMs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbWrapper := database.NewDB(db)

	vms, err := dbWrapper.GetUserVMs(userIDStr)
	if err != nil {
		http.Error(w, "Failed to list VMs", http.StatusInternalServerError)
		return
	}

	// Convert to response format
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetVM gets details for a specific VM
// GET /api/vms/{id}
func HandleGetVM(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		http.Error(w, "VM ID is required", http.StatusBadRequest)
		return
	}

	dbWrapper := database.NewDB(db)

	vm, err := dbWrapper.GetVMByIDAndUser(vmID, userIDStr)
	if err != nil {
		http.Error(w, "VM not found", http.StatusNotFound)
		return
	}

	response := VMResponse{
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleStartVM starts a stopped VM
// POST /api/vms/{id}/start
func HandleStartVM(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		http.Error(w, "VM ID is required", http.StatusBadRequest)
		return
	}

	dbWrapper := database.NewDB(db)

	vm, err := dbWrapper.GetVMByIDAndUser(vmID, userIDStr)
	if err != nil {
		http.Error(w, "VM not found", http.StatusNotFound)
		return
	}

	if vm.Status != "stopped" {
		http.Error(w, "VM must be stopped to start", http.StatusBadRequest)
		return
	}

	// Update status to pending (will be updated to running by k8s manager)
	err = dbWrapper.UpdateVMStatus(vmID, "pending")
	if err != nil {
		http.Error(w, "Failed to update VM status", http.StatusInternalServerError)
		return
	}

	// Start VM in k8s asynchronously
	go func() {
		// TODO: Call vmManager.StartVM
		time.Sleep(2 * time.Second)
		dbWrapper.UpdateVMStatus(vmID, "running")
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "pending",
	})
}

// HandleStopVM stops a running VM
// POST /api/vms/{id}/stop
func HandleStopVM(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		http.Error(w, "VM ID is required", http.StatusBadRequest)
		return
	}

	dbWrapper := database.NewDB(db)

	vm, err := dbWrapper.GetVMByIDAndUser(vmID, userIDStr)
	if err != nil {
		http.Error(w, "VM not found", http.StatusNotFound)
		return
	}

	if vm.Status != "running" && vm.Status != "pending" {
		http.Error(w, "VM must be running to stop", http.StatusBadRequest)
		return
	}

	// Update status to stopped
	err = dbWrapper.UpdateVMStatus(vmID, "stopped")
	if err != nil {
		http.Error(w, "Failed to update VM status", http.StatusInternalServerError)
		return
	}

	// Stop VM in k8s asynchronously
	go func() {
		// TODO: Call vmManager.StopVM
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "stopped",
	})
}

// HandleRestartVM restarts a running VM
// POST /api/vms/{id}/restart
func HandleRestartVM(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		http.Error(w, "VM ID is required", http.StatusBadRequest)
		return
	}

	dbWrapper := database.NewDB(db)

	vm, err := dbWrapper.GetVMByIDAndUser(vmID, userIDStr)
	if err != nil {
		http.Error(w, "VM not found", http.StatusNotFound)
		return
	}

	if vm.Status != "running" {
		http.Error(w, "VM must be running to restart", http.StatusBadRequest)
		return
	}

	// Update status to pending (restart in progress)
	err = dbWrapper.UpdateVMStatus(vmID, "pending")
	if err != nil {
		http.Error(w, "Failed to update VM status", http.StatusInternalServerError)
		return
	}

	// Restart VM in k8s asynchronously
	go func() {
		// TODO: Call vmManager.RestartVM
		time.Sleep(2 * time.Second)
		dbWrapper.UpdateVMStatus(vmID, "running")
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "pending",
	})
}

// HandleDeleteVM deletes a VM
// DELETE /api/vms/{id}
func HandleDeleteVM(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vmID := chi.URLParam(r, "id")
	if vmID == "" {
		http.Error(w, "VM ID is required", http.StatusBadRequest)
		return
	}

	dbWrapper := database.NewDB(db)

	vm, err := dbWrapper.GetVMByIDAndUser(vmID, userIDStr)
	if err != nil {
		http.Error(w, "VM not found", http.StatusNotFound)
		return
	}

	// Delete k8s resources asynchronously
	go func() {
		// TODO: Call vmManager.DeleteVM
		// TODO: Create snapshot before delete (7-day retention)
	}()

	// Update quota usage (decrease)
	err = dbWrapper.UpdateUsage(r.Context(), userIDStr, -vm.CPU, -vm.RAM, -vm.Storage, -1)
	if err != nil {
		log.Printf("Failed to update quota after delete: %v", err)
		// Continue with delete anyway
	}

	// Delete from database (soft delete)
	err = dbWrapper.DeleteVM(vmID)
	if err != nil {
		http.Error(w, "Failed to delete VM", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to convert *string to string
func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

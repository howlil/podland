package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
	"github.com/podland/backend/pkg/response"
)

// AdminHandler handles admin HTTP requests
type AdminHandler struct {
	userRepo  repository.UserRepository
	auditRepo repository.AuditRepository
	vmRepo    repository.VMRepository
}

// NewAdminHandler creates a new admin handler with dependencies
func NewAdminHandler(userRepo repository.UserRepository, auditRepo repository.AuditRepository, vmRepo repository.VMRepository) *AdminHandler {
	return &AdminHandler{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		vmRepo:    vmRepo,
	}
}

// ListUsers returns all users with optional role filter
// GET /api/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role") // optional filter

	var users []*entity.User
	var err error
	if role != "" {
		users, err = h.userRepo.GetUsersByRole(r.Context(), role)
	} else {
		users, err = h.userRepo.GetAllUsers(r.Context())
	}

	if err != nil {
		pkgresponse.InternalError(w, "Failed to fetch users")
		return
	}

	pkgresponse.JSON(w, http.StatusOK, users)
}

// ChangeRoleRequest represents the request body for changing a user's role
type ChangeRoleRequest struct {
	Role string `json:"role"`
}

// ChangeRole changes a user's role (superadmin only)
// PATCH /api/admin/users/{id}/role
func (h *AdminHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var req ChangeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgresponse.BadRequest(w, "Invalid request body")
		return
	}

	// Validate role
	validRoles := map[string]bool{
		"internal":   true,
		"external":   true,
		"superadmin": true,
	}

	if !validRoles[req.Role] {
		pkgresponse.BadRequest(w, "Invalid role. Must be 'internal', 'external', or 'superadmin'")
		return
	}

	err := h.userRepo.UpdateUserRole(r.Context(), userID, req.Role)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to update role")
		return
	}

	pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// BanUser bans a user
// POST /api/admin/users/{id}/ban
func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	err := h.userRepo.BanUser(r.Context(), userID)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to ban user")
		return
	}

	pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// UnbanUser unbans a user
// POST /api/admin/users/{id}/unban
func (h *AdminHandler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	err := h.userRepo.UnbanUser(r.Context(), userID)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to unban user")
		return
	}

	pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// SystemHealth returns cluster health metrics
// GET /api/admin/health
func (h *AdminHandler) SystemHealth(w http.ResponseWriter, r *http.Request) {
	// Query Prometheus for cluster metrics (placeholder - will be implemented with actual Prometheus queries)
	health := map[string]interface{}{
		"cluster_cpu":     45.2, // %
		"cluster_memory":  62.5, // %
		"cluster_storage": 38.1, // %
		"total_users":     487,
		"total_vms":       234,
		"active_vms":      156,
	}

	pkgresponse.JSON(w, http.StatusOK, health)
}

// AuditLog returns audit log entries
// GET /api/admin/audit-log
func (h *AdminHandler) AuditLog(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}

	logs, err := h.auditRepo.GetRecent(r.Context(), limit)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to fetch audit logs")
		return
	}

	pkgresponse.JSON(w, http.StatusOK, logs)
}

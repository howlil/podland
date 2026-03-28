package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/podland/backend/internal/domain"
	pkgresponse "github.com/podland/backend/pkg/response"
)

// DomainHandler handles domain HTTP requests
type DomainHandler struct {
	domainService *domain.DomainService
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(domainService *domain.DomainService) *DomainHandler {
	return &DomainHandler{
		domainService: domainService,
	}
}

// GetDomains returns all domains for the authenticated user
// GET /api/domains
func (h *DomainHandler) GetDomains(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	domains, err := h.domainService.GetDomainsByUserID(r.Context(), userID)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to fetch domains")
		return
	}

	pkgresponse.Success(w, http.StatusOK, map[string]interface{}{
		"domains": domains,
	})
}

// DeleteDomain deletes a domain by ID
// DELETE /api/domains/{id}
func (h *DomainHandler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "id")
	domainID := vars
	if domainID == "" {
		pkgresponse.BadRequest(w, "Domain ID is required")
		return
	}

	userID := getUserIDFromContext(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	err := h.domainService.DeleteDomain(r.Context(), domainID, userID)
	if err != nil {
		switch err.Error() {
		case "domain not found", "unauthorized":
			pkgresponse.NotFound(w, "Domain not found")
		case "domain not assigned":
			pkgresponse.BadRequest(w, "Domain not assigned to VM")
		default:
			pkgresponse.InternalError(w, "Failed to delete domain")
		}
		return
	}

	pkgresponse.NoContent(w)
}

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(r *http.Request) string {
	userID := r.Context().Value("user_id")
	if userID == nil {
		return ""
	}
	userIDStr, ok := userID.(string)
	if !ok {
		return ""
	}
	return userIDStr
}

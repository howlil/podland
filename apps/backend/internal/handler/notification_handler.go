package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
	"github.com/podland/backend/pkg/response"
)

// NotificationHandler handles notification API requests
type NotificationHandler struct {
	notificationRepo repository.NotificationRepository
	authHelper       *AuthHelper
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationRepo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		notificationRepo: notificationRepo,
		authHelper:       NewAuthHelper(),
	}
}

// ListNotifications returns all notifications for the authenticated user
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		pkgresponse.BadRequest(w, "Invalid user ID format")
		return
	}

	// Parse limit query param
	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	notifications, err := h.notificationRepo.GetByUserID(r.Context(), userUUID, limit)
	if err != nil {
		log.Printf("Failed to get notifications for user %s: %v", userID, err)
		pkgresponse.InternalError(w, "Failed to get notifications")
		return
	}

	pkgresponse.Success(w, http.StatusOK, h.toNotificationResponse(notifications))
}

// GetUnreadCount returns the count of unread notifications
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		pkgresponse.BadRequest(w, "Invalid user ID format")
		return
	}

	count, err := h.notificationRepo.GetUnreadCount(r.Context(), userUUID)
	if err != nil {
		log.Printf("Failed to get unread count for user %s: %v", userID, err)
		pkgresponse.InternalError(w, "Failed to get unread count")
		return
	}

	pkgresponse.Success(w, http.StatusOK, map[string]int{"count": count})
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		pkgresponse.BadRequest(w, "Invalid user ID format")
		return
	}

	notificationIDStr := chi.URLParam(r, "id")
	if notificationIDStr == "" {
		pkgresponse.BadRequest(w, "Notification ID is required")
		return
	}

	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		pkgresponse.BadRequest(w, "Invalid notification ID format")
		return
	}

	err = h.notificationRepo.MarkAsRead(r.Context(), notificationID, userUUID)
	if err != nil {
		log.Printf("Failed to mark notification as read: %v", err)
		pkgresponse.InternalError(w, "Failed to mark notification as read")
		return
	}

	pkgresponse.Success(w, http.StatusOK, map[string]string{"status": "ok"})
}

// MarkAllAsRead marks all notifications as read for the user
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := h.authHelper.GetAuthUserID(r)
	if userID == "" {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		pkgresponse.BadRequest(w, "Invalid user ID format")
		return
	}

	err = h.notificationRepo.MarkAllAsRead(r.Context(), userUUID)
	if err != nil {
		log.Printf("Failed to mark all notifications as read: %v", err)
		pkgresponse.InternalError(w, "Failed to mark all notifications as read")
		return
	}

	pkgresponse.Success(w, http.StatusOK, map[string]string{"status": "ok"})
}

// NotificationResponse represents a notification in API responses
type NotificationResponse struct {
	ID         string  `json:"id"`
	VMID       string  `json:"vm_id"`
	AlertName  string  `json:"alert_name"`
	Severity   string  `json:"severity"`
	Title      string  `json:"title"`
	Message    string  `json:"message"`
	IsRead     bool    `json:"is_read"`
	CreatedAt  string  `json:"created_at"`
	ResolvedAt *string `json:"resolved_at,omitempty"`
}

func (h *NotificationHandler) toNotificationResponse(notifications []*entity.Notification) []NotificationResponse {
	response := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		resp := NotificationResponse{
			ID:        n.ID.String(),
			VMID:      n.VMID.String(),
			AlertName: n.AlertName,
			Severity:  n.Severity,
			Title:     n.Title,
			Message:   n.Message,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if n.ResolvedAt != nil {
			formatted := n.ResolvedAt.Format("2006-01-02T15:04:05Z")
			resp.ResolvedAt = &formatted
		}
		response[i] = resp
	}
	return response
}

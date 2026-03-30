package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
	"github.com/podland/backend/pkg/response"
)

// AlertWebhookHandler handles Alertmanager webhook callbacks
type AlertWebhookHandler struct {
	serviceToken       string
	vmRepo             repository.VMRepository
	notificationRepo   repository.NotificationRepository
}

// AlertmanagerPayload represents the webhook payload from Alertmanager
type AlertmanagerPayload struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
	Alerts            []Alert           `json:"alerts"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
}

// Alert represents a single alert from Alertmanager
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// NewAlertWebhookHandler creates a new alert webhook handler
func NewAlertWebhookHandler(vmRepo repository.VMRepository, notificationRepo repository.NotificationRepository) *AlertWebhookHandler {
	serviceToken := os.Getenv("ALERTMANAGER_WEBHOOK_SECRET")
	if serviceToken == "" {
		// Alertmanager is optional - return nil handler if not configured
		log.Println("ALERTMANAGER_WEBHOOK_SECRET not configured. Alert webhook disabled.")
		return nil
	}

	return &AlertWebhookHandler{
		serviceToken:     serviceToken,
		vmRepo:           vmRepo,
		notificationRepo: notificationRepo,
	}
}

// HandleAlert receives alerts from Alertmanager and creates notifications
func (h *AlertWebhookHandler) HandleAlert(w http.ResponseWriter, r *http.Request) {
	// Verify internal service token
	token := r.Header.Get("X-Service-Token")
	if token != "" && token != h.serviceToken {
		// Allow requests without token for testing, but log warning
		log.Printf("WARNING: Alert webhook received with invalid token: %s", token)
	}

	var payload AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Failed to decode alert payload: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Process each alert
	for _, alert := range payload.Alerts {
		// Extract vm_id from alert labels
		vmIDStr := alert.Labels["vm_id"]
		if vmIDStr == "" {
			log.Printf("Alert missing vm_id label, skipping: %+v", alert.Labels)
			continue
		}

		vmID, err := uuid.Parse(vmIDStr)
		if err != nil {
			log.Printf("Invalid vm_id format: %s, error: %v", vmIDStr, err)
			continue
		}

		// Get VM owner from database
		vm, err := h.vmRepo.GetVMByID(r.Context(), vmID.String())
		if err != nil {
			log.Printf("Failed to get VM %s: %v", vmID, err)
			continue
		}

		if vm == nil {
			log.Printf("VM %s not found, skipping alert", vmID)
			continue
		}

		userID, err := uuid.Parse(vm.UserID)
		if err != nil {
			log.Printf("Invalid user_id format: %s, error: %v", vm.UserID, err)
			continue
		}

		// Create notification
		notification := entity.NewNotification(
			userID,
			vmID,
			alert.Labels["alertname"],
			alert.Labels["severity"],
			alert.Annotations["summary"],
			alert.Annotations["description"],
		)

		// If alert is resolved, mark as resolved
		if alert.Status == "resolved" {
			notification.MarkResolved()
		}

		if err := h.notificationRepo.Create(r.Context(), notification); err != nil {
			log.Printf("Failed to create notification: %v", err)
			// Continue processing other alerts
		} else {
			log.Printf("Created notification for VM %s, alert %s, status %s", vmID, alert.Labels["alertname"], alert.Status)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetVMAlerts retrieves alerts for a specific VM
func (h *AlertWebhookHandler) GetVMAlerts(w http.ResponseWriter, r *http.Request) {
	vmIDStr := chi.URLParam(r, "id")
	if vmIDStr == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	vmID, err := uuid.Parse(vmIDStr)
	if err != nil {
		pkgresponse.BadRequest(w, "Invalid VM ID format")
		return
	}

	limit := 50 // Default limit
	notifications, err := h.notificationRepo.GetByVMID(r.Context(), vmID, limit)
	if err != nil {
		log.Printf("Failed to get notifications for VM %s: %v", vmID, err)
		pkgresponse.InternalError(w, "Failed to get alerts")
		return
	}

	pkgresponse.Success(w, http.StatusOK, h.toAlertResponse(notifications))
}

func (h *AlertWebhookHandler) toAlertResponse(notifications []*entity.Notification) []AlertResponse {
	response := make([]AlertResponse, len(notifications))
	for i, n := range notifications {
		response[i] = AlertResponse{
			ID:         n.ID,
			AlertName:  n.AlertName,
			Severity:   n.Severity,
			Title:      n.Title,
			Message:    n.Message,
			IsRead:     n.IsRead,
			CreatedAt:  n.CreatedAt,
			ResolvedAt: n.ResolvedAt,
		}
	}
	return response
}

// AlertResponse represents an alert in API responses
type AlertResponse struct {
	ID         uuid.UUID  `json:"id"`
	AlertName  string     `json:"alert_name"`
	Severity   string     `json:"severity"`
	Title      string     `json:"title"`
	Message    string     `json:"message"`
	IsRead     bool       `json:"is_read"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

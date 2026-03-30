package entity

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents an in-app notification for alerts
type Notification struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	VMID       uuid.UUID
	AlertName  string
	Severity   string
	Title      string
	Message    string
	IsRead     bool
	CreatedAt  time.Time
	ResolvedAt *time.Time
}

// NewNotification creates a new notification
func NewNotification(userID, vmID uuid.UUID, alertName, severity, title, message string) *Notification {
	return &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		VMID:      vmID,
		AlertName: alertName,
		Severity:  severity,
		Title:     title,
		Message:   message,
		IsRead:    false,
		CreatedAt: time.Now().UTC(),
	}
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	n.IsRead = true
}

// MarkResolved marks the notification as resolved
func (n *Notification) MarkResolved() {
	now := time.Now().UTC()
	n.ResolvedAt = &now
}

// IsResolved returns true if the notification has been resolved
func (n *Notification) IsResolved() bool {
	return n.ResolvedAt != nil
}

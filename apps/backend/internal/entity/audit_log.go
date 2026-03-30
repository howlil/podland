package entity

import "time"

// AuditLog represents an audit log entry for admin actions
type AuditLog struct {
	ID        string
	UserID    string
	Action    string
	IPAddress string
	UserAgent string
	CreatedAt time.Time
}

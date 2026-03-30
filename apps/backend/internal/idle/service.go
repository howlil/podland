package idle

import (
	"github.com/podland/backend/internal/email"
)

// EmailService defines the interface for sending emails
type EmailService interface {
	SendIdleWarning(userEmail, userName, vmName, vmID string) error
}

// Ensure email.EmailService implements idle.EmailService
var _ EmailService = (*email.EmailService)(nil)

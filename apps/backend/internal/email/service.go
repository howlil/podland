package email

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailService handles email sending via SendGrid
type EmailService struct {
	client    *sendgrid.Client
	fromEmail string
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromEmail := os.Getenv("SENDGRID_FROM_EMAIL")

	if apiKey == "" || fromEmail == "" {
		log.Println("WARNING: SendGrid credentials not configured. Email notifications disabled.")
		return &EmailService{}
	}

	return &EmailService{
		client:    sendgrid.NewSendClient(apiKey),
		fromEmail: fromEmail,
	}
}

// SendIdleWarning sends an idle warning email to the user
func (s *EmailService) SendIdleWarning(userEmail, userName, vmName, vmID string) error {
	if s.client == nil {
		return fmt.Errorf("email service not configured")
	}

	deleteAt := time.Now().Add(24 * time.Hour)

	// Create email content
	subject := "Your VM will be deleted in 24 hours"
	htmlBody := s.renderHTMLTemplate(userName, vmName, vmID, deleteAt)
	textBody := s.renderTextTemplate(userName, vmName, vmID, deleteAt)

	// Create email
	from := mail.NewEmail("Podland", s.fromEmail)
	to := mail.NewEmail(userName, userEmail)

	// Build email using SendGrid helper
	message := mail.NewV3MailInit(from, subject, to, mail.NewContent("text/html", htmlBody))
	message.AddContent(mail.NewContent("text/plain", textBody))

	// Send with retry
	return s.sendWithRetry(message, 3)
}

func (s *EmailService) renderHTMLTemplate(userName, vmName, vmID string, deleteAt time.Time) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .warning { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
    .button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px; }
    .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
  </style>
</head>
<body>
  <div class="container">
    <h2>VM Idle Warning</h2>
    <p>Hi %s,</p>

    <div class="warning">
      <strong>⚠️ Action Required:</strong> Your VM <strong>"%s"</strong> (ID: %s) has been idle for 48 hours.
    </div>

    <p>To keep your VM running, please:</p>
    <ol>
      <li>Access your VM dashboard</li>
      <li>Restart or interact with your VM</li>
      <li>Ensure your application is running</li>
    </ol>

    <p><strong>If no activity is detected, your VM will be automatically deleted on:</strong></p>
    <p style="font-size: 18px; font-weight: bold; color: #dc3545;">%s</p>

    <p style="margin-top: 30px;">
      <a href="https://podland.app/dashboard" class="button">Go to Dashboard</a>
    </p>

    <div class="footer">
      <p>This is an automated message from Podland. Please do not reply.</p>
    </div>
  </div>
</body>
</html>
`, userName, vmName, vmID, deleteAt.Format("2006-01-02 15:04:05"))
}

func (s *EmailService) renderTextTemplate(userName, vmName, vmID string, deleteAt time.Time) string {
	return fmt.Sprintf(`
Hi %s,

ACTION REQUIRED: Your VM "%s" (ID: %s) has been idle for 48 hours.

To keep your VM running, please:
1. Access your VM dashboard at https://podland.app/dashboard
2. Restart or interact with your VM
3. Ensure your application is running

If no activity is detected, your VM will be automatically deleted on: %s

Best regards,
The Podland Team

---
This is an automated message. Please do not reply.
`, userName, vmName, vmID, deleteAt.Format("2006-01-02 15:04:05"))
}

func (s *EmailService) sendWithRetry(message *mail.SGMailV3, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		response, err := s.client.Send(message)
		if err == nil && response.StatusCode == 202 {
			log.Printf("Email sent successfully")
			return nil
		}

		lastErr = err
		log.Printf("Email send failed (attempt %d/%d): %v", i+1, maxRetries, err)

		// Exponential backoff
		backoff := time.Duration(i*i) * time.Minute
		if backoff > 10*time.Minute {
			backoff = 10 * time.Minute
		}
		time.Sleep(backoff)
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

package middleware

import (
	"fmt"
	"net/http"

	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
)

// AuditLogger automatically logs all admin actions to the audit log
func AuditLogger(auditRepo repository.AuditRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _ := r.Context().Value("user_id").(string)
			action := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ip := r.RemoteAddr
			userAgent := r.UserAgent()

			// Log asynchronously (don't block request)
			go func() {
				auditRepo.Create(r.Context(), &entity.AuditLog{
					UserID:    userID,
					Action:    action,
					IPAddress: ip,
					UserAgent: userAgent,
				})
			}()

			next.ServeHTTP(w, r)
		})
	}
}

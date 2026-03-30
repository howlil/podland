package middleware

import (
	"context"
	"net/http"

	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
)

// AdminOnly restricts access to superadmin users only
func AdminOnly(userRepo repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("user_id").(string)
			if !ok || userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := userRepo.GetUserByID(r.Context(), userID)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if user.Role != "superadmin" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Add user to context for downstream handlers
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuthUser retrieves the authenticated user from context
func GetAuthUser(r *http.Request) *entity.User {
	user, ok := r.Context().Value("user").(*entity.User)
	if !ok {
		return nil
	}
	return user
}

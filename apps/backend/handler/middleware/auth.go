package middleware

import (
	"context"
	"net/http"
)

// AuthHelper provides authentication helper functions
type AuthHelper struct{}

// NewAuthHelper creates a new auth helper
func NewAuthHelper() *AuthHelper {
	return &AuthHelper{}
}

// GetAuthUserID extracts user ID from request context
func (h *AuthHelper) GetAuthUserID(r *http.Request) string {
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

// GetAuthUserEmail extracts email from request context
func (h *AuthHelper) GetAuthUserEmail(r *http.Request) string {
	email := r.Context().Value("email")
	if email == nil {
		return ""
	}
	emailStr, ok := email.(string)
	if !ok {
		return ""
	}
	return emailStr
}

// EnsureAuthUserID validates user ID exists in context and returns it
func EnsureAuthUserID(r *http.Request) (string, bool) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		return "", false
	}
	userIDStr, ok := userID.(string)
	if !ok {
		return "", false
	}
	return userIDStr, true
}

// ContextWithUser returns a context with user info added
func ContextWithUser(ctx context.Context, userID, email string) context.Context {
	ctx = context.WithValue(ctx, "user_id", userID)
	ctx = context.WithValue(ctx, "email", email)
	return ctx
}

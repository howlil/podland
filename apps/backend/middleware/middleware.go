package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/podland/backend/internal/auth"
	"github.com/podland/backend/internal/database"
)

var (
	dbInstance *sql.DB
	dbOnce     sync.Once
)

// InitDB initializes the database connection (call once at startup)
func InitDB() error {
	var err error
	dbOnce.Do(func() {
		dbInstance, err = database.Init()
	})
	return err
}

// getDBConnection returns the database connection
func getDBConnection() *sql.DB {
	return dbInstance
}

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-XSRF-TOKEN")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isOriginAllowed(origin, allowedOrigins string) bool {
	origins := strings.Split(allowedOrigins, ",")
	for _, o := range origins {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

// CSRFMiddleware validates CSRF tokens for state-changing requests
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for safe methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Validate XSRF-TOKEN header
		xsrfHeader := r.Header.Get("X-XSRF-TOKEN")
		xsrfCookie, err := r.Cookie("XSRF-TOKEN")
		if err != nil || xsrfHeader != xsrfCookie.Value {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates JWT tokens and adds user info to context
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get access token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

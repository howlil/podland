package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/podland/backend/internal/auth"
	"github.com/podland/backend/internal/database"
)

var db *sql.DB

// welcomeSessions stores temporary welcome session data (token -> user ID)
var welcomeSessions sync.Map

// SetDB sets the database connection for handlers
func SetDB(database *sql.DB) {
	db = database
}

// HandleLogin initiates the OAuth login flow
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state token for CSRF protection
	state, err := auth.GenerateStateToken()
	if err != nil {
		http.Error(w, "Failed to generate state token", http.StatusInternalServerError)
		return
	}

	// Set state cookie (5 min expiry, HTTP-only)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   300, // 5 minutes
	})

	// Generate OAuth URL and redirect
	url := auth.GetLoginURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback handles the OAuth callback from GitHub
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate state parameter
	state := r.FormValue("state")
	oauthStateCookie, err := r.Cookie("oauth_state")
	if err != nil || state != oauthStateCookie.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	code := r.FormValue("code")
	token, err := auth.ExchangeToken(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create HTTP client with OAuth token
	client := auth.GetHTTPClient(ctx, token)

	// Fetch user info
	githubUser, err := auth.FetchUser(ctx, client)
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	// Fetch emails
	emails, err := auth.FetchEmails(ctx, client)
	if err != nil {
		http.Error(w, "Failed to fetch emails", http.StatusInternalServerError)
		return
	}

	// Find primary email
	primaryEmail := auth.FindPrimaryEmail(emails)
	if primaryEmail == "" {
		http.Error(w, "No verified primary email found", http.StatusBadRequest)
		return
	}

	// Validate student email
	if !auth.IsValidStudentEmail(primaryEmail) {
		// Redirect to rejection page
		http.Redirect(w, r, "/auth/rejected?reason=invalid_email", http.StatusTemporaryRedirect)
		return
	}

	// Extract NIM and assign role
	nim := auth.ExtractNIM(primaryEmail)
	role := auth.AssignRole(nim)

	// Get database connection
	dbWrapper := database.NewDB(db)

	// Check if user exists
	user, err := dbWrapper.GetUserByGitHubID(githubUser.ID)
	if err == sql.ErrNoRows {
		// New user - download avatar only if not already saved
		avatarURL := githubUser.AvatarURL // Fallback to GitHub URL
		
		// Sanitize filename to prevent path traversal
		safeFilename := githubUser.ID
		if idx := strings.LastIndex(safeFilename, "/"); idx != -1 {
			safeFilename = safeFilename[idx+1:]
		}
		filePath := "uploads/avatars/" + safeFilename + ".jpg"
		
		// Only download if file doesn't exist
		if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
			avatarData, fetchErr := auth.FetchAvatar(githubUser.AvatarURL)
			if fetchErr == nil {
				// Ensure directory exists
				if mkdirErr := os.MkdirAll("uploads/avatars", 0755); mkdirErr == nil {
					if writeErr := os.WriteFile(filePath, avatarData, 0644); writeErr == nil {
						avatarURL = "/uploads/avatars/" + safeFilename + ".jpg"
					}
				}
			}
		}

		// Create user
		user, err = dbWrapper.CreateUser(database.NewUserCreateInput(
			githubUser.ID,
			primaryEmail,
			githubUser.Name,
			avatarURL,
			nim,
			role,
		))
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// Log activity
		_ = dbWrapper.CreateActivityLog(user.ID, "account_created", nil)

		// Create temporary welcome session (store user ID in session, not URL)
		welcomeToken := uuid.New().String()
		// Store in context for welcome route to retrieve (via session cookie)
		http.SetCookie(w, &http.Cookie{
			Name:     "welcome_token",
			Value:    welcomeToken,
			Path:     "/auth",
			HttpOnly: true,
			Secure:   os.Getenv("ENV") == "production",
			SameSite: http.SameSiteStrictMode,
			MaxAge:   300, // 5 minutes
		})
		// Store user ID mapped to token (in-memory for simplicity, could use Redis)
		welcomeSessions.Store(welcomeToken, user.ID)

		// Redirect to welcome screen (no user ID in URL)
		http.Redirect(w, r, "/auth/welcome", http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Existing user - create session
	deviceInfo := auth.DeviceInfo{
		UserAgent: r.UserAgent(),
		IP:        getClientIP(r),
	}

	session, err := auth.CreateSession(dbWrapper, user.ID, deviceInfo)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set auth cookies
	setAuthCookies(w, session, user)

	// Log activity
	_ = dbWrapper.CreateActivityLog(user.ID, "signed_in", map[string]interface{}{
		"ip": deviceInfo.IP,
	})

	// Redirect to dashboard
	http.Redirect(w, r, os.Getenv("FRONTEND_URL")+"/dashboard", http.StatusTemporaryRedirect)
}

// HandleRefresh handles refresh token rotation
func HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get refresh token from cookie
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
		return
	}

	// Rotate token
	dbWrapper := database.NewDB(db)
	newSession, err := auth.RotateRefreshToken(dbWrapper, refreshCookie.Value)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := dbWrapper.GetUserByID(newSession.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Set new cookies
	setAuthCookies(w, newSession, user)

	// Return new access token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": generateAccessToken(user),
	})
}

// HandleLogout handles user logout
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get refresh token from cookie and revoke session
	refreshCookie, err := r.Cookie("refresh_token")
	if err == nil && refreshCookie.Value != "" {
		dbWrapper := database.NewDB(db)
		refreshHash := auth.HashToken(refreshCookie.Value)
		session, err := dbWrapper.GetSessionByRefreshToken(refreshHash)
		if err == nil && session.ID != "" {
			_ = dbWrapper.RevokeSession(session.ID, time.Now())
		}
	}

	// Clear cookies
	clearAuthCookies(w)

	// Redirect to login
	http.Redirect(w, r, "/api/auth/login", http.StatusTemporaryRedirect)
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header - only trust first IP to prevent spoofing
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain (client IP)
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}
	return ip
}

func setAuthCookies(w http.ResponseWriter, session *auth.Session, user *database.User) {
	// Generate access token
	accessToken, _ := auth.GenerateAccessToken(user.ID, user.Email)

	// Set refresh token cookie (HTTP-only)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    session.RefreshToken,
		Path:     "/api/auth/refresh",
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		Expires:  session.ExpiresAt,
	})

	// Set XSRF token cookie (JavaScript-readable)
	xsrfToken, _ := auth.GenerateXSRFToken()
	http.SetCookie(w, &http.Cookie{
		Name:     "XSRF-TOKEN",
		Value:    xsrfToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		Expires:  session.ExpiresAt,
	})

	// Return access token in response body (for in-memory storage)
	_ = accessToken
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth/refresh",
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "XSRF-TOKEN",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func generateAccessToken(user *database.User) string {
	token, _ := auth.GenerateAccessToken(user.ID, user.Email)
	return token
}

// HandleGetWelcomeUser returns user data for welcome screen (uses welcome_token cookie)
func HandleGetWelcomeUser(w http.ResponseWriter, r *http.Request) {
	welcomeCookie, err := r.Cookie("welcome_token")
	if err != nil {
		http.Error(w, "No welcome session", http.StatusNotFound)
		return
	}

	userIDInterface, ok := welcomeSessions.Load(welcomeCookie.Value)
	if !ok {
		http.Error(w, "Welcome session expired", http.StatusGone)
		return
	}

	userID := userIDInterface.(string)
	dbWrapper := database.NewDB(db)
	user, err := dbWrapper.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

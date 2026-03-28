package handler

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/podland/backend/internal/auth"
	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
	"github.com/podland/backend/pkg/response"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	userRepo  repository.UserRepository
	sessionRepo repository.SessionRepository
	quotaRepo repository.QuotaRepository
}

// NewAuthHandler creates a new auth handler with dependencies
func NewAuthHandler(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	quotaRepo repository.QuotaRepository,
) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		quotaRepo:   quotaRepo,
	}
}

// HandleLogin initiates the OAuth login flow
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state token for CSRF protection
	state, err := auth.GenerateStateToken()
	if err != nil {
		pkgresponse.InternalError(w, "Failed to generate state token")
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
func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate state parameter
	state := r.FormValue("state")
	oauthStateCookie, err := r.Cookie("oauth_state")
	if err != nil || state != oauthStateCookie.Value {
		pkgresponse.BadRequest(w, "Invalid state parameter")
		return
	}

	// Exchange code for token
	code := r.FormValue("code")
	token, err := auth.ExchangeToken(ctx, code)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to exchange token: "+err.Error())
		return
	}

	// Create HTTP client with OAuth token
	client := auth.GetHTTPClient(ctx, token)

	// Fetch user info
	githubUser, err := auth.FetchUser(ctx, client)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to fetch user info")
		return
	}

	// Fetch emails
	emails, err := auth.FetchEmails(ctx, client)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to fetch emails")
		return
	}

	// Find primary email
	primaryEmail := auth.FindPrimaryEmail(emails)
	if primaryEmail == "" {
		pkgresponse.BadRequest(w, "No verified primary email found")
		return
	}

	// Validate student email
	if !auth.IsValidStudentEmail(primaryEmail) {
		http.Redirect(w, r, "/auth/rejected?reason=invalid_email", http.StatusTemporaryRedirect)
		return
	}

	// Extract NIM and assign role
	nim := auth.ExtractNIM(primaryEmail)
	role := auth.AssignRole(nim)

	// Check if user exists
	user, err := h.userRepo.GetUserByGitHubID(ctx, githubUser.ID.String())
	if err == repository.ErrUserNotFound {
		// New user - download avatar only if not already saved
		avatarURL := githubUser.AvatarURL // Fallback to GitHub URL

		// Sanitize filename to prevent path traversal
		safeFilename := githubUser.ID.String()
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
		user, err = h.userRepo.CreateUser(ctx, repository.UserCreateInput{
			GithubID:    githubUser.ID.String(),
			Email:       primaryEmail,
			DisplayName: githubUser.Name,
			AvatarURL:   avatarURL,
			NIM:         nim,
			Role:        role,
		})
		if err != nil {
			pkgresponse.InternalError(w, "Failed to create user")
			return
		}

		// Log activity
		_ = h.userRepo.CreateActivityLog(ctx, user.ID, "account_created", nil)

		// Create session for new user
		deviceInfo := auth.DeviceInfo{
			UserAgent: r.UserAgent(),
			IP:        getClientIP(r),
		}
		session, err := auth.CreateSession(ctx, h.sessionRepo, user.ID, deviceInfo)
		if err != nil {
			pkgresponse.InternalError(w, "Failed to create session")
			return
		}
		setAuthCookies(w, session, user)
		_ = h.userRepo.CreateActivityLog(ctx, user.ID, "signed_in", map[string]interface{}{
			"ip": deviceInfo.IP,
		})

		// Redirect to dashboard
		http.Redirect(w, r, os.Getenv("FRONTEND_URL")+"/dashboard", http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		pkgresponse.InternalError(w, "Database error")
		return
	}

	// Existing user - create session
	deviceInfo := auth.DeviceInfo{
		UserAgent: r.UserAgent(),
		IP:        getClientIP(r),
	}

	session, err := auth.CreateSession(ctx, h.sessionRepo, user.ID, deviceInfo)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to create session")
		return
	}

	// Set auth cookies
	setAuthCookies(w, session, user)

	// Log activity
	_ = h.userRepo.CreateActivityLog(ctx, user.ID, "signed_in", map[string]interface{}{
		"ip": deviceInfo.IP,
	})

	// Redirect to dashboard
	http.Redirect(w, r, os.Getenv("FRONTEND_URL")+"/dashboard", http.StatusTemporaryRedirect)
}

// HandleRefresh handles refresh token rotation
func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		pkgresponse.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get refresh token from cookie
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		pkgresponse.Error(w, http.StatusUnauthorized, "Missing refresh token")
		return
	}

	// Rotate token
	newSession, err := auth.RotateRefreshToken(r.Context(), h.sessionRepo, refreshCookie.Value)
	if err != nil {
		pkgresponse.Error(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Get user
	user, err := h.userRepo.GetUserByID(r.Context(), newSession.UserID)
	if err != nil {
		pkgresponse.InternalError(w, "User not found")
		return
	}

	// Set new cookies
	setAuthCookies(w, newSession, user)

	// Return new access token
	pkgresponse.Success(w, http.StatusOK, map[string]string{
		"access_token": generateAccessToken(user),
	})
}

// HandleLogout handles user logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		pkgresponse.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get refresh token from cookie and revoke session
	refreshCookie, err := r.Cookie("refresh_token")
	if err == nil && refreshCookie.Value != "" {
		refreshHash := auth.HashToken(refreshCookie.Value)
		session, err := h.sessionRepo.GetSessionByRefreshToken(r.Context(), refreshHash)
		if err == nil && session.ID != "" {
			_ = h.sessionRepo.RevokeSession(r.Context(), session.ID, time.Now())
		}
	}

	// Clear cookies
	clearAuthCookies(w)

	// Redirect to login
	http.Redirect(w, r, "/api/auth/login", http.StatusTemporaryRedirect)
}

// HandleGetMe returns the current authenticated user
func (h *AuthHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value("user_id")
	if userIDRaw == nil {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		pkgresponse.NotFound(w, "User not found")
		return
	}

	pkgresponse.Success(w, http.StatusOK, user)
}

// HandleGetUser returns a user by ID (requester must own the account)
func (h *AuthHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	requesterIDRaw := r.Context().Value("user_id")
	if requesterIDRaw == nil {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}
	requesterID, ok := requesterIDRaw.(string)
	if !ok {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	// Get user ID from URL path (need to parse from chi router)
	userID := r.PathValue("id")
	if userID == "" {
		pkgresponse.BadRequest(w, "User ID is required")
		return
	}

	if userID != requesterID {
		pkgresponse.Forbidden(w, "Forbidden")
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		pkgresponse.NotFound(w, "User not found")
		return
	}

	pkgresponse.Success(w, http.StatusOK, user)
}

// HandleConfirmNIM confirms or updates the user's NIM
func (h *AuthHandler) HandleConfirmNIM(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value("user_id")
	if userIDRaw == nil {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	var req struct {
		NIM string `json:"nim"`
	}

	if err := r.ParseForm(); err != nil {
		pkgresponse.BadRequest(w, "Invalid request body")
		return
	}

	// Decode JSON body
	decoder := r.Context().Value("decoder")
	if decoder != nil {
		// Use decoder if available
	}

	if req.NIM == "" {
		pkgresponse.BadRequest(w, "NIM is required")
		return
	}

	// Update NIM (this also recalculates role based on NIM)
	if err := h.userRepo.UpdateUserNIM(r.Context(), userID, req.NIM); err != nil {
		pkgresponse.InternalError(w, "Failed to update NIM")
		return
	}

	// Log activity
	_ = h.userRepo.CreateActivityLog(r.Context(), userID, "nim_confirmed", map[string]interface{}{
		"nim": req.NIM,
	})

	pkgresponse.Success(w, http.StatusOK, map[string]string{
		"status": "success",
	})
}

// HandleGetActivity returns the user's activity log
func (h *AuthHandler) HandleGetActivity(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value("user_id")
	if userIDRaw == nil {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		pkgresponse.Unauthorized(w, "Unauthorized")
		return
	}

	activities, err := h.userRepo.GetUserActivity(r.Context(), userID, 50)
	if err != nil {
		pkgresponse.InternalError(w, "Failed to fetch activity")
		return
	}

	pkgresponse.Success(w, http.StatusOK, activities)
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

func setAuthCookies(w http.ResponseWriter, session *auth.Session, user *entity.User) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    session.RefreshToken,
		Path:     "/api/auth/refresh",
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		Expires:  session.ExpiresAt,
	})

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

func generateAccessToken(user *entity.User) string {
	token, _ := auth.GenerateAccessToken(user.ID, user.Email)
	return token
}

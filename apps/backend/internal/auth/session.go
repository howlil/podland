package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/podland/backend/internal/repository"
)

var (
	ErrSessionLimit     = errors.New("session limit reached")
	ErrSessionRevoked   = errors.New("session revoked")
	ErrSessionExpired   = errors.New("session expired")
	ErrSessionNotFound  = errors.New("session not found")
	ErrTokenReuse       = errors.New("token reuse detected")
)

const MaxSessionsPerUser = 3

// Session represents an authenticated session
type Session struct {
	ID           string
	UserID       string
	RefreshToken string
	JTI          string
	DeviceInfo   DeviceInfo
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// DeviceInfo contains information about the client device
type DeviceInfo struct {
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
}

// CreateSession creates a new session for a user
func CreateSession(ctx context.Context, sessionRepo repository.SessionRepository, userID string, deviceInfo DeviceInfo) (*Session, error) {
	// Check concurrent session limit
	count, err := sessionRepo.GetActiveSessionCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	if count >= MaxSessionsPerUser {
		// Revoke oldest session
		if err := sessionRepo.RevokeOldestSession(ctx, userID); err != nil {
			return nil, err
		}
	}

	// Generate tokens
	refreshToken, jti, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		RefreshToken: refreshToken,
		JTI:          jti,
		DeviceInfo:   deviceInfo,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	// Hash the refresh token for storage
	refreshTokenHash := HashToken(session.RefreshToken)

	// Store in database - convert DeviceInfo to JSON
	deviceInfoJSON, err := json.Marshal(deviceInfo)
	if err != nil {
		return nil, err
	}

	dbSession, err := sessionRepo.CreateSession(ctx, repository.SessionCreateInput{
		UserID:           session.UserID,
		RefreshTokenHash: refreshTokenHash,
		JTI:              session.JTI,
		DeviceInfo:       deviceInfoJSON,
		ExpiresAt:        session.ExpiresAt,
	})

	if err != nil {
		return nil, err
	}

	// Convert repository session to auth session
	session.ID = dbSession.ID
	return session, nil
}

// RotateRefreshToken rotates a refresh token (revoke old, create new) using atomic transaction
func RotateRefreshToken(ctx context.Context, sessionRepo repository.SessionRepository, oldToken string) (*Session, error) {
	oldHash := HashToken(oldToken)

	// Find and validate old token first (for error messages)
	oldSession, err := sessionRepo.GetSessionByRefreshToken(ctx, oldHash)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if oldSession.ExpiresAt.Before(time.Now()) {
		return nil, ErrSessionExpired
	}

	if oldSession.RevokedAt != nil {
		// Token reuse detected - security alert, revoke all sessions
		_ = sessionRepo.RevokeAllUserSessions(ctx, oldSession.UserID)
		return nil, ErrTokenReuse
	}

	// Use atomic transaction-based rotation
	dbSession, err := sessionRepo.RotateSession(ctx, oldHash, oldSession.UserID, oldSession.DeviceInfo)
	if err != nil {
		return nil, err
	}

	// Convert repository session to auth session
	session := &Session{
		ID:           dbSession.ID,
		UserID:       dbSession.UserID,
		RefreshToken: dbSession.RefreshTokenHash, // Raw token returned for cookie
		JTI:          dbSession.JTI,
		CreatedAt:    dbSession.CreatedAt,
		ExpiresAt:    dbSession.ExpiresAt,
	}

	// Parse device info from JSON
	if len(dbSession.DeviceInfo) > 0 {
		var deviceInfo DeviceInfo
		if err := json.Unmarshal(dbSession.DeviceInfo, &deviceInfo); err == nil {
			session.DeviceInfo = deviceInfo
		}
	}

	return session, nil
}

// GenerateXSRFToken generates a CSRF token
func GenerateXSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

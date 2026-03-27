package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/podland/backend/internal/database"
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
func CreateSession(db database.DB, userID string, deviceInfo DeviceInfo) (*Session, error) {
	// Check concurrent session limit
	count, err := db.GetActiveSessionCount(userID)
	if err != nil {
		return nil, err
	}

	if count >= MaxSessionsPerUser {
		// Revoke oldest session
		if err := db.RevokeOldestSession(userID); err != nil {
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

	// Store in database - convert DeviceInfo to JSON
	deviceInfoJSON, err := json.Marshal(deviceInfo)
	if err != nil {
		return nil, err
	}

	err = db.CreateSession(database.Session{
		ID:               session.ID,
		UserID:           session.UserID,
		RefreshTokenHash: HashToken(session.RefreshToken),
		JTI:              session.JTI,
		DeviceInfo:       deviceInfoJSON,
		CreatedAt:        session.CreatedAt,
		ExpiresAt:        session.ExpiresAt,
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}

// RotateRefreshToken rotates a refresh token (revoke old, create new) using atomic transaction
func RotateRefreshToken(db database.DB, oldToken string) (*Session, error) {
	oldHash := HashToken(oldToken)

	// Find and validate old token first (for error messages)
	oldSession, err := db.GetSessionByRefreshToken(oldHash)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if oldSession.ExpiresAt.Before(time.Now()) {
		return nil, ErrSessionExpired
	}

	if oldSession.RevokedAt != nil {
		// Token reuse detected - security alert, revoke all sessions
		_ = db.RevokeAllUserSessions(oldSession.UserID)
		return nil, ErrTokenReuse
	}

	// Use atomic transaction-based rotation
	dbSession, err := db.RotateSession(oldHash, oldSession.UserID, oldSession.DeviceInfo)
	if err != nil {
		return nil, err
	}

	// Convert database.Session to auth.Session
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

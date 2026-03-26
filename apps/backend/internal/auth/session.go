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
	refreshToken, jti := GenerateRefreshToken()

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
	deviceInfoJSON, _ := json.Marshal(deviceInfo)
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

// RotateRefreshToken rotates a refresh token (revoke old, create new)
func RotateRefreshToken(db database.DB, oldToken string) (*Session, error) {
	oldHash := HashToken(oldToken)

	// Find and validate old token
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

	// Revoke old token
	err = db.RevokeSession(oldSession.ID, time.Now())
	if err != nil {
		return nil, err
	}

	// Parse device info from JSON
	var deviceInfo DeviceInfo
	_ = json.Unmarshal(oldSession.DeviceInfo, &deviceInfo)

	// Create new session
	newSession, err := CreateSession(db, oldSession.UserID, deviceInfo)
	if err != nil {
		return nil, err
	}

	// Link old to new (for audit trail)
	_ = db.LinkSessionReplacement(oldSession.ID, newSession.ID)

	return newSession, nil
}

// GenerateXSRFToken generates a CSRF token
func GenerateXSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	CreateSession(ctx context.Context, input SessionCreateInput) (*Session, error)
	GetSessionByID(ctx context.Context, id string) (*Session, error)
	GetSessionByRefreshToken(ctx context.Context, hash string) (*Session, error)
	GetActiveSessionCount(ctx context.Context, userID string) (int, error)
	RevokeSession(ctx context.Context, id string, revokedAt time.Time) error
	RevokeOldestSession(ctx context.Context, userID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
	RotateSession(ctx context.Context, oldTokenHash string, userID string, deviceInfo json.RawMessage) (*Session, error)
}

// Session represents a session
type Session struct {
	ID               string          `json:"id"`
	UserID           string          `json:"user_id"`
	RefreshTokenHash string          `json:"refresh_token_hash"`
	JTI              string          `json:"jti"`
	DeviceInfo       json.RawMessage `json:"device_info"`
	CreatedAt        time.Time       `json:"created_at"`
	ExpiresAt        time.Time       `json:"expires_at"`
	RevokedAt        *time.Time      `json:"revoked_at,omitempty"`
	ReplacedBy       *string         `json:"replaced_by,omitempty"`
}

// SessionCreateInput represents input for creating a session
type SessionCreateInput struct {
	UserID           string
	RefreshTokenHash string
	JTI              string
	DeviceInfo       json.RawMessage
	ExpiresAt        time.Time
}

// sessionRepository implements SessionRepository
type sessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// CreateSession creates a new session
func (r *sessionRepository) CreateSession(ctx context.Context, input SessionCreateInput) (*Session, error) {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token_hash, jti, device_info, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6)
		RETURNING id, user_id, refresh_token_hash, jti, device_info, created_at, expires_at, revoked_at, replaced_by
	`

	session := &Session{}
	err := r.db.QueryRowContext(ctx, query,
		uuid.New().String(),
		input.UserID,
		input.RefreshTokenHash,
		input.JTI,
		input.DeviceInfo,
		input.ExpiresAt,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.JTI,
		&session.DeviceInfo,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.ReplacedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

// GetSessionByID gets a session by ID
func (r *sessionRepository) GetSessionByID(ctx context.Context, id string) (*Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, jti, device_info, created_at, expires_at, revoked_at, replaced_by
		FROM sessions
		WHERE id = $1
	`

	session := &Session{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.JTI,
		&session.DeviceInfo,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.ReplacedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session by ID: %w", err)
	}

	return session, nil
}

// GetSessionByRefreshToken gets a session by refresh token hash
func (r *sessionRepository) GetSessionByRefreshToken(ctx context.Context, hash string) (*Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, jti, device_info, created_at, expires_at, revoked_at, replaced_by
		FROM sessions
		WHERE refresh_token_hash = $1
	`

	session := &Session{}
	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.JTI,
		&session.DeviceInfo,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.ReplacedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session by refresh token: %w", err)
	}

	return session, nil
}

// GetActiveSessionCount gets the count of active sessions for a user
func (r *sessionRepository) GetActiveSessionCount(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM sessions
		WHERE user_id = $1
		AND revoked_at IS NULL
		AND expires_at > NOW()
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

// RevokeSession revokes a session
func (r *sessionRepository) RevokeSession(ctx context.Context, id string, revokedAt time.Time) error {
	query := `
		UPDATE sessions
		SET revoked_at = $2
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, revokedAt)
	return err
}

// RevokeOldestSession revokes the oldest session for a user
func (r *sessionRepository) RevokeOldestSession(ctx context.Context, userID string) error {
	query := `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE id = (
			SELECT id
			FROM sessions
			WHERE user_id = $1
			AND revoked_at IS NULL
			AND expires_at > NOW()
			ORDER BY created_at ASC
			LIMIT 1
		)
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *sessionRepository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	query := `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE user_id = $1
		AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// RotateSession atomically rotates a session token (revoke old, create new)
func (r *sessionRepository) RotateSession(ctx context.Context, oldTokenHash string, userID string, deviceInfo json.RawMessage) (*Session, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Find the old session
	var oldID string
	query := `SELECT id FROM sessions WHERE refresh_token_hash = $1`
	if err := tx.QueryRowContext(ctx, query, oldTokenHash).Scan(&oldID); err != nil {
		return nil, err
	}

	// Revoke old session
	query = `UPDATE sessions SET revoked_at = NOW() WHERE id = $1`
	if _, err := tx.ExecContext(ctx, query, oldID); err != nil {
		return nil, err
	}

	// Generate new refresh token and JTI
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	newRefreshToken := base64.URLEncoding.EncodeToString(bytes)

	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return nil, err
	}
	newJTI := hex.EncodeToString(jtiBytes)

	// Hash the refresh token for storage
	hash := sha256.Sum256([]byte(newRefreshToken))
	newRefreshTokenHash := hex.EncodeToString(hash[:])

	newSessionID := uuid.New().String()

	query = `
		INSERT INTO sessions (id, user_id, refresh_token_hash, jti, device_info, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW() + INTERVAL '7 days')
	`
	if _, err := tx.ExecContext(ctx, query, newSessionID, userID, newRefreshTokenHash, newJTI, deviceInfo); err != nil {
		return nil, err
	}

	// Link old to new
	query = `UPDATE sessions SET replaced_by = $2 WHERE id = $1`
	if _, err := tx.ExecContext(ctx, query, oldID, newSessionID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return session with raw token for cookie
	return &Session{
		ID:               newSessionID,
		UserID:           userID,
		RefreshTokenHash: newRefreshToken, // Return raw token
		JTI:              newJTI,
		DeviceInfo:       deviceInfo,
		CreatedAt:        time.Now(),
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
	}, nil
}

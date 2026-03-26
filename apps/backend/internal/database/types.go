package database

import (
	"database/sql"
	"encoding/json"
	"time"
)

// DB interface for database operations
type DB interface {
	// User operations
	CreateUser(user UserCreateInput) (*User, error)
	GetUserByID(id string) (*User, error)
	GetUserByGitHubID(githubID string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(id string, user UserUpdateInput) error
	UpdateUserNIM(userID, nim string) error

	// Session operations
	CreateSession(session Session) error
	GetSessionByID(id string) (*Session, error)
	GetSessionByRefreshToken(hash string) (*Session, error)
	GetActiveSessionCount(userID string) (int, error)
	RevokeSession(id string, revokedAt time.Time) error
	RevokeOldestSession(userID string) error
	RevokeAllUserSessions(userID string) error
	LinkSessionReplacement(oldID, newID string) error

	// Activity operations
	CreateActivityLog(userID string, action string, metadata map[string]interface{}) error
	GetUserActivity(userID string, limit int) ([]ActivityLog, error)

	// Close database connection
	Close() error
}

// User represents a user in the database
type User struct {
	ID          string     `json:"id"`
	GithubID    string     `json:"github_id"`
	Email       string     `json:"email"`
	DisplayName *string    `json:"display_name,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	NIM         string     `json:"nim"`
	Role        string     `json:"role"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UserCreateInput represents input for creating a user
type UserCreateInput struct {
	GithubID    string
	Email       string
	DisplayName string
	AvatarURL   string
	NIM         string
	Role        string
}

// UserUpdateInput represents input for updating a user
type UserUpdateInput struct {
	DisplayName *string
	AvatarURL   *string
}

// Session represents a session in the database
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

// ActivityLog represents an activity log entry
type ActivityLog struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Action    string          `json:"action"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// sqlDB implements the DB interface
type sqlDB struct {
	db *sql.DB
}

// NewDB creates a new database wrapper
func NewDB(db *sql.DB) DB {
	return &sqlDB{db: db}
}

func (d *sqlDB) Close() error {
	return d.db.Close()
}

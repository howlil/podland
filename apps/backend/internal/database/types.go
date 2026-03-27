package database

import (
	"context"
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
	// Transaction-based session rotation (atomic operation)
	RotateSession(oldTokenHash string, userID string, deviceInfo json.RawMessage) (*Session, error)

	// Activity operations
	CreateActivityLog(userID string, action string, metadata map[string]interface{}) error
	GetUserActivity(userID string, limit int) ([]ActivityLog, error)

	// VM operations
	CreateVM(vm VMCreateInput) (*VM, error)
	GetVMByID(id string) (*VM, error)
	GetVMByIDAndUser(id, userID string) (*VM, error)
	GetUserVMs(userID string) ([]*VM, error)
	UpdateVM(id string, vm VMUpdateInput) error
	DeleteVM(id string) error
	UpdateVMStatus(id, status string) error

	// Quota operations
	CheckQuota(ctx context.Context, userID string, cpu float64, ram, storage int64) error
	UpdateUsage(ctx context.Context, userID string, cpu float64, ram, storage int64, vmCountDelta int) error
	GetQuota(ctx context.Context, userID string) (*Quota, error)
	GetTier(ctx context.Context, name string) (*Tier, error)
	GetAllTiers(ctx context.Context) ([]*Tier, error)

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

// VM represents a virtual machine instance
type VM struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Name          string     `json:"name"`
	OS            string     `json:"os"`
	Tier          string     `json:"tier"`
	CPU           float64    `json:"cpu"`
	RAM           int64      `json:"ram"`
	Storage       int64      `json:"storage"`
	Status        string     `json:"status"`
	K8sNamespace  *string    `json:"k8s_namespace,omitempty"`
	K8sDeployment *string    `json:"k8s_deployment,omitempty"`
	K8sService    *string    `json:"k8s_service,omitempty"`
	K8sPVC        *string    `json:"k8s_pvc,omitempty"`
	Domain        *string    `json:"domain,omitempty"`
	SSHPublicKey  *string    `json:"ssh_public_key,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	StoppedAt     *time.Time `json:"stopped_at,omitempty"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// VMCreateInput represents input for creating a VM
type VMCreateInput struct {
	UserID       string
	Name         string
	OS           string
	Tier         string
	CPU          float64
	RAM          int64
	Storage      int64
	SSHPublicKey string
}

// VMUpdateInput represents input for updating a VM
type VMUpdateInput struct {
	Status       *string
	K8sNamespace *string
	K8sDeployment *string
	K8sService   *string
	K8sPVC       *string
	Domain       *string
	StartedAt    *time.Time
	StoppedAt    *time.Time
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

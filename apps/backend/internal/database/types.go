package database

import (
	"encoding/json"
	"time"
)

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
	Status        *string
	K8sNamespace  *string
	K8sDeployment *string
	K8sService    *string
	K8sPVC        *string
	Domain        *string
	StartedAt     *string
	StoppedAt     *string
}

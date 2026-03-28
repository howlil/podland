package repository

import (
	"encoding/json"
	"errors"
)

// Common repository errors
var (
	ErrVMNotFound       = errors.New("vm not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrQuotaNotFound    = errors.New("quota not found")
	ErrSessionNotFound  = errors.New("session not found")
)

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
	DomainStatus  *string
	StartedAt     *string
	StoppedAt     *string
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
	NIM         *string
	Role        *string
}

// ActivityLog represents an activity log entry
type ActivityLog struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Action    string          `json:"action"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt string          `json:"created_at"`
}

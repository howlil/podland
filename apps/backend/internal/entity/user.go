package entity

import "time"

// User represents a user in the system (pure domain object - no DB/JSON tags)
type User struct {
	ID          string
	GithubID    string
	Email       string
	DisplayName string
	AvatarURL   string
	NIM         string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsInternal returns true if user has internal role
func (u *User) IsInternal() bool {
	return u.Role == "internal"
}

// IsExternal returns true if user has external role
func (u *User) IsExternal() bool {
	return u.Role == "external"
}

// IsSuperAdmin returns true if user has superadmin role
func (u *User) IsSuperAdmin() bool {
	return u.Role == "superadmin"
}

// HasNIM returns true if user has a NIM (student ID)
func (u *User) HasNIM() bool {
	return u.NIM != ""
}

// IsStudent returns true if NIM contains student pattern (e.g., 1152)
func (u *User) IsStudent() bool {
	// Students have NIM containing "1152"
	return len(u.NIM) >= 4 && u.NIM[:4] == "1152"
}

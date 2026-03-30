package entity

import "time"

// VM represents a virtual machine instance (pure domain object - no DB/JSON tags)
type VM struct {
	ID           string
	UserID       string
	Name         string
	OS           string
	Tier         string
	CPU          float64
	RAM          int64
	Storage      int64
	Status       string
	Domain       *string
	SSHPublicKey *string
	IsPinned     bool
	IdleWarnedAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    *time.Time
	StoppedAt    *time.Time
	DeletedAt    *time.Time
}

// IsRunning returns true if VM is in running state
func (v *VM) IsRunning() bool {
	return v.Status == "running"
}

// IsStopped returns true if VM is in stopped state
func (v *VM) IsStopped() bool {
	return v.Status == "stopped"
}

// IsPending returns true if VM is in pending state
func (v *VM) IsPending() bool {
	return v.Status == "pending"
}

// CanStart returns true if VM can be started (must be stopped)
func (v *VM) CanStart() bool {
	return v.Status == "stopped"
}

// CanStop returns true if VM can be stopped (must be running or pending)
func (v *VM) CanStop() bool {
	return v.Status == "running" || v.Status == "pending"
}

// CanRestart returns true if VM can be restarted (must be running)
func (v *VM) CanRestart() bool {
	return v.Status == "running"
}

// IsActive returns true if VM is not deleted
func (v *VM) IsActive() bool {
	return v.DeletedAt == nil
}

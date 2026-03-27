package entity

import "time"

// Quota represents a user's resource quota (pure domain object - no DB/JSON tags)
type Quota struct {
	UserID       string
	CPULimit     float64
	RAMLimit     int64
	StorageLimit int64
	VMCountLimit int
	CPUUsed      float64
	RAMUsed      int64
	StorageUsed  int64
	VMCount      int
	UpdatedAt    time.Time
}

// AvailableCPU returns the available CPU capacity
func (q *Quota) AvailableCPU() float64 {
	return q.CPULimit - q.CPUUsed
}

// AvailableRAM returns the available RAM capacity in bytes
func (q *Quota) AvailableRAM() int64 {
	return q.RAMLimit - q.RAMUsed
}

// AvailableStorage returns the available storage capacity in bytes
func (q *Quota) AvailableStorage() int64 {
	return q.StorageLimit - q.StorageUsed
}

// AvailableVMCount returns the number of additional VMs that can be created
func (q *Quota) AvailableVMCount() int {
	return q.VMCountLimit - q.VMCount
}

// CanCreateVM returns true if a VM with the given resources can be created
func (q *Quota) CanCreateVM(cpu float64, ram, storage int64) bool {
	return q.AvailableCPU() >= cpu &&
		q.AvailableRAM() >= ram &&
		q.AvailableStorage() >= storage &&
		q.AvailableVMCount() >= 1
}

// UsagePercentCPU returns the CPU usage percentage
func (q *Quota) UsagePercentCPU() float64 {
	if q.CPULimit == 0 {
		return 0
	}
	return (q.CPUUsed / q.CPULimit) * 100
}

// UsagePercentRAM returns the RAM usage percentage
func (q *Quota) UsagePercentRAM() float64 {
	if q.RAMLimit == 0 {
		return 0
	}
	return (float64(q.RAMUsed) / float64(q.RAMLimit)) * 100
}

// UsagePercentStorage returns the storage usage percentage
func (q *Quota) UsagePercentStorage() float64 {
	if q.StorageLimit == 0 {
		return 0
	}
	return (float64(q.StorageUsed) / float64(q.StorageLimit)) * 100
}

// Tier represents a VM tier configuration (pure domain object - no DB/JSON tags)
type Tier struct {
	Name    string
	CPU     float64
	RAM     int64
	Storage int64
	MinRole string
}

// IsValidForRole returns true if the tier is available for the given role
func (t *Tier) IsValidForRole(role string) bool {
	if t.MinRole == "internal" {
		return role == "internal" || role == "superadmin"
	}
	if t.MinRole == "external" {
		return role == "external" || role == "internal" || role == "superadmin"
	}
	return true
}

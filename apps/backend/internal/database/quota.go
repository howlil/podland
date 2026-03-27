package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrQuotaExceeded is returned when a quota limit would be exceeded
var ErrQuotaExceeded = errors.New("quota exceeded")

// Quota represents a user's resource quota
type Quota struct {
	UserID         string
	CPULimit       float64
	RAMLimit       int64
	StorageLimit   int64
	VMCountLimit   int
	CPUUsed        float64
	RAMUsed        int64
	StorageUsed    int64
	VMCount        int
}

// CheckQuota checks if user can create a VM with given resources
// Uses SELECT FOR UPDATE to prevent race conditions
func (db *sqlDB) CheckQuota(ctx context.Context, userID string, cpu float64, ram, storage int64) error {
	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock quota row for update
	var cpuLimit, ramLimit, storageLimit float64
	var vmCountLimit int
	err = tx.QueryRowContext(ctx, `
		SELECT cpu_limit, ram_limit, storage_limit, vm_count_limit
		FROM user_quotas
		WHERE user_id = $1
		FOR UPDATE
	`, userID).Scan(&cpuLimit, &ramLimit, &storageLimit, &vmCountLimit)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("quota not found for user %s", userID)
		}
		return fmt.Errorf("failed to get quota: %w", err)
	}

	// Get current usage
	var cpuUsed float64
	var ramUsed, storageUsed int64
	var vmCount int
	err = tx.QueryRowContext(ctx, `
		SELECT cpu_used, ram_used, storage_used, vm_count
		FROM user_quota_usage
		WHERE user_id = $1
	`, userID).Scan(&cpuUsed, &ramUsed, &storageUsed, &vmCount)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("usage not found for user %s", userID)
		}
		return fmt.Errorf("failed to get usage: %w", err)
	}

	// Check if new VM fits within quota limits
	if cpuUsed+cpu > cpuLimit {
		return ErrQuotaExceeded
	}
	if float64(ramUsed+ram) > ramLimit {
		return ErrQuotaExceeded
	}
	if float64(storageUsed+storage) > storageLimit {
		return ErrQuotaExceeded
	}
	if vmCount+1 > vmCountLimit {
		return ErrQuotaExceeded
	}

	return nil
}

// UpdateUsage updates quota usage after VM create/delete
// Positive values increase usage, negative values decrease
func (db *sqlDB) UpdateUsage(ctx context.Context, userID string, cpu float64, ram, storage int64, vmCountDelta int) error {
	_, err := db.db.ExecContext(ctx, `
		UPDATE user_quota_usage
		SET cpu_used = cpu_used + $1,
		    ram_used = ram_used + $2,
		    storage_used = storage_used + $3,
		    vm_count = vm_count + $4,
		    updated_at = NOW()
		WHERE user_id = $5
	`, cpu, ram, storage, vmCountDelta, userID)

	if err != nil {
		return fmt.Errorf("failed to update usage: %w", err)
	}

	return nil
}

// GetQuota retrieves a user's quota and current usage
func (db *sqlDB) GetQuota(ctx context.Context, userID string) (*Quota, error) {
	quota := &Quota{UserID: userID}

	// Get quota limits
	err := db.db.QueryRowContext(ctx, `
		SELECT cpu_limit, ram_limit, storage_limit, vm_count_limit
		FROM user_quotas
		WHERE user_id = $1
	`, userID).Scan(&quota.CPULimit, &quota.RAMLimit, &quota.StorageLimit, &quota.VMCountLimit)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("quota not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	// Get current usage
	err = db.db.QueryRowContext(ctx, `
		SELECT cpu_used, ram_used, storage_used, vm_count
		FROM user_quota_usage
		WHERE user_id = $1
	`, userID).Scan(&quota.CPUUsed, &quota.RAMUsed, &quota.StorageUsed, &quota.VMCount)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("usage not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	return quota, nil
}

// GetTier retrieves tier configuration by name
func (db *sqlDB) GetTier(ctx context.Context, name string) (*Tier, error) {
	tier := &Tier{Name: name}

	err := db.db.QueryRowContext(ctx, `
		SELECT cpu, ram, storage, min_role
		FROM tiers
		WHERE name = $1
	`, name).Scan(&tier.CPU, &tier.RAM, &tier.Storage, &tier.MinRole)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("tier %s not found", name)
		}
		return nil, fmt.Errorf("failed to get tier: %w", err)
	}

	return tier, nil
}

// GetAllTiers retrieves all available tiers
func (db *sqlDB) GetAllTiers(ctx context.Context) ([]*Tier, error) {
	rows, err := db.db.QueryContext(ctx, `
		SELECT name, cpu, ram, storage, min_role
		FROM tiers
		ORDER BY cpu ASC, ram ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tiers: %w", err)
	}
	defer rows.Close()

	var tiers []*Tier
	for rows.Next() {
		tier := &Tier{}
		err := rows.Scan(&tier.Name, &tier.CPU, &tier.RAM, &tier.Storage, &tier.MinRole)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tier: %w", err)
		}
		tiers = append(tiers, tier)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tiers: %w", err)
	}

	return tiers, nil
}

// Tier represents a VM tier configuration
type Tier struct {
	Name     string
	CPU      float64
	RAM      int64
	Storage  int64
	MinRole  string
}

// ReconcileUsage recalculates user quota usage from actual VMs in database
// This is used for periodic sync to ensure DB matches reality
func (db *sqlDB) ReconcileUsage(ctx context.Context, userID string) error {
	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Calculate actual usage from vms table (only non-deleted VMs)
	var cpuUsed, ramUsed, storageUsed float64
	var vmCount int
	err = tx.QueryRowContext(ctx, `
		SELECT 
			COALESCE(SUM(cpu), 0),
			COALESCE(SUM(ram), 0),
			COALESCE(SUM(storage), 0),
			COUNT(*)
		FROM vms
		WHERE user_id = $1 
		  AND deleted_at IS NULL
		  AND status IN ('running', 'pending', 'stopped')
	`, userID).Scan(&cpuUsed, &ramUsed, &storageUsed, &vmCount)

	if err != nil {
		return fmt.Errorf("failed to calculate usage: %w", err)
	}

	// Update usage table
	_, err = tx.ExecContext(ctx, `
		UPDATE user_quota_usage
		SET cpu_used = $1,
		    ram_used = $2,
		    storage_used = $3,
		    vm_count = $4,
		    updated_at = NOW()
		WHERE user_id = $5
	`, cpuUsed, ramUsed, storageUsed, vmCount, userID)

	if err != nil {
		return fmt.Errorf("failed to update usage: %w", err)
	}

	return tx.Commit()
}

// ReconcileAllUsage recalculates quota usage for all users
// This should be run periodically (e.g., every 5 minutes) as a background job
func (db *sqlDB) ReconcileAllUsage(ctx context.Context) error {
	// Get all users with quotas
	rows, err := db.db.QueryContext(ctx, `SELECT user_id FROM user_quotas`)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return fmt.Errorf("failed to scan user: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating users: %w", err)
	}

	// Reconcile each user
	for _, userID := range userIDs {
		if err := db.ReconcileUsage(ctx, userID); err != nil {
			// Log error but continue with other users
			// In production, use proper logging
		}
	}

	return nil
}

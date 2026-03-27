package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/podland/backend/internal/entity"
)

// QuotaRepository defines the interface for quota data access
type QuotaRepository interface {
	CheckQuota(ctx context.Context, userID string, cpu float64, ram, storage int64) error
	UpdateUsage(ctx context.Context, userID string, cpu float64, ram, storage int64, vmCountDelta int) error
	GetQuota(ctx context.Context, userID string) (*entity.Quota, error)
	GetTier(ctx context.Context, name string) (*entity.Tier, error)
	GetAllTiers(ctx context.Context) ([]*entity.Tier, error)
}

// quotaRepository implements QuotaRepository
type quotaRepository struct {
	db *sql.DB
}

// NewQuotaRepository creates a new quota repository
func NewQuotaRepository(db *sql.DB) QuotaRepository {
	return &quotaRepository{db: db}
}

// CheckQuota checks if user can create a VM with given resources
// Uses SELECT FOR UPDATE to prevent race conditions
func (r *quotaRepository) CheckQuota(ctx context.Context, userID string, cpu float64, ram, storage int64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
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
			return fmt.Errorf("quota not found for user %s: %w", userID, ErrQuotaNotFound)
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
		return errors.New("quota exceeded: CPU limit")
	}
	if float64(ramUsed+ram) > ramLimit {
		return errors.New("quota exceeded: RAM limit")
	}
	if float64(storageUsed+storage) > storageLimit {
		return errors.New("quota exceeded: storage limit")
	}
	if vmCount+1 > vmCountLimit {
		return errors.New("quota exceeded: VM count limit")
	}

	return nil
}

// UpdateUsage updates quota usage after VM create/delete
// Positive values increase usage, negative values decrease
func (r *quotaRepository) UpdateUsage(ctx context.Context, userID string, cpu float64, ram, storage int64, vmCountDelta int) error {
	_, err := r.db.ExecContext(ctx, `
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
func (r *quotaRepository) GetQuota(ctx context.Context, userID string) (*entity.Quota, error) {
	quota := &entity.Quota{UserID: userID}

	// Get quota limits
	err := r.db.QueryRowContext(ctx, `
		SELECT cpu_limit, ram_limit, storage_limit, vm_count_limit
		FROM user_quotas
		WHERE user_id = $1
	`, userID).Scan(&quota.CPULimit, &quota.RAMLimit, &quota.StorageLimit, &quota.VMCountLimit)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("quota not found for user %s: %w", userID, ErrQuotaNotFound)
		}
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	// Get current usage
	err = r.db.QueryRowContext(ctx, `
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
func (r *quotaRepository) GetTier(ctx context.Context, name string) (*entity.Tier, error) {
	tier := &entity.Tier{Name: name}

	err := r.db.QueryRowContext(ctx, `
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
func (r *quotaRepository) GetAllTiers(ctx context.Context) ([]*entity.Tier, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT name, cpu, ram, storage, min_role
		FROM tiers
		ORDER BY cpu ASC, ram ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tiers: %w", err)
	}
	defer rows.Close()

	var tiers []*entity.Tier
	for rows.Next() {
		tier := &entity.Tier{}
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

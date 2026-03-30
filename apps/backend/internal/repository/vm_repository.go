package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/podland/backend/internal/entity"
)

// VMRepository defines the interface for VM data access
type VMRepository interface {
	CreateVM(ctx context.Context, input VMCreateInput) (*entity.VM, error)
	GetVMByID(ctx context.Context, id string) (*entity.VM, error)
	GetVMByIDAndUser(ctx context.Context, id, userID string) (*entity.VM, error)
	GetUserVMs(ctx context.Context, userID string) ([]*entity.VM, error)
	UpdateVMStatus(ctx context.Context, id, status string) error
	UpdateVM(ctx context.Context, id string, input VMUpdateInput) error
	DeleteVM(ctx context.Context, id string) error
	// Pin methods
	SetPinned(ctx context.Context, id string, pinned bool) error
	GetPinnedCount(ctx context.Context, userID string) int
	GetIdleVMs(ctx context.Context, hours int) ([]*entity.VM, error)
	// Idle warning methods
	SetIdleWarnedAt(ctx context.Context, id string, warnedAt time.Time) error
}

// vmRepository implements VMRepository
type vmRepository struct {
	db *sql.DB
}

// NewVMRepository creates a new VM repository
func NewVMRepository(db *sql.DB) VMRepository {
	return &vmRepository{db: db}
}

// CreateVM creates a new VM in the database
func (r *vmRepository) CreateVM(ctx context.Context, input VMCreateInput) (*entity.VM, error) {
	query := `
		INSERT INTO vms (user_id, name, os, tier, cpu, ram, storage, status, ssh_public_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, NOW(), NOW())
		RETURNING id, user_id, name, os, tier, cpu, ram, storage, status, domain, ssh_public_key, created_at, updated_at
	`

	vm := &entity.VM{}
	err := r.db.QueryRowContext(ctx, query,
		input.UserID,
		input.Name,
		input.OS,
		input.Tier,
		input.CPU,
		input.RAM,
		input.Storage,
		input.SSHPublicKey,
	).Scan(
		&vm.ID,
		&vm.UserID,
		&vm.Name,
		&vm.OS,
		&vm.Tier,
		&vm.CPU,
		&vm.RAM,
		&vm.Storage,
		&vm.Status,
		&vm.Domain,
		&vm.SSHPublicKey,
		&vm.CreatedAt,
		&vm.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create VM: %w", err)
	}

	return vm, nil
}

// GetVMByID gets a VM by ID
func (r *vmRepository) GetVMByID(ctx context.Context, id string) (*entity.VM, error) {
	query := `
		SELECT id, user_id, name, os, tier, cpu, ram, storage, status, domain, ssh_public_key, is_pinned, idle_warned_at, created_at, updated_at, started_at, stopped_at, deleted_at
		FROM vms
		WHERE id = $1 AND deleted_at IS NULL
	`

	vm := &entity.VM{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&vm.ID,
		&vm.UserID,
		&vm.Name,
		&vm.OS,
		&vm.Tier,
		&vm.CPU,
		&vm.RAM,
		&vm.Storage,
		&vm.Status,
		&vm.Domain,
		&vm.SSHPublicKey,
		&vm.IsPinned,
		&vm.IdleWarnedAt,
		&vm.CreatedAt,
		&vm.UpdatedAt,
		&vm.StartedAt,
		&vm.StoppedAt,
		&vm.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVMNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get VM by ID: %w", err)
	}

	return vm, nil
}

// GetVMByIDAndUser gets a VM by ID and user ID (ownership check)
func (r *vmRepository) GetVMByIDAndUser(ctx context.Context, id, userID string) (*entity.VM, error) {
	query := `
		SELECT id, user_id, name, os, tier, cpu, ram, storage, status, domain, ssh_public_key, is_pinned, idle_warned_at, created_at, updated_at, started_at, stopped_at, deleted_at
		FROM vms
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	vm := &entity.VM{}
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&vm.ID,
		&vm.UserID,
		&vm.Name,
		&vm.OS,
		&vm.Tier,
		&vm.CPU,
		&vm.RAM,
		&vm.Storage,
		&vm.Status,
		&vm.Domain,
		&vm.SSHPublicKey,
		&vm.IsPinned,
		&vm.IdleWarnedAt,
		&vm.CreatedAt,
		&vm.UpdatedAt,
		&vm.StartedAt,
		&vm.StoppedAt,
		&vm.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVMNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get VM by ID and user: %w", err)
	}

	return vm, nil
}

// GetUserVMs gets all VMs for a user
func (r *vmRepository) GetUserVMs(ctx context.Context, userID string) ([]*entity.VM, error) {
	query := `
		SELECT id, user_id, name, os, tier, cpu, ram, storage, status, domain, ssh_public_key, is_pinned, idle_warned_at, created_at, updated_at, started_at, stopped_at, deleted_at
		FROM vms
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user VMs: %w", err)
	}
	defer rows.Close()

	var vms []*entity.VM
	for rows.Next() {
		vm := &entity.VM{}
		err := rows.Scan(
			&vm.ID,
			&vm.UserID,
			&vm.Name,
			&vm.OS,
			&vm.Tier,
			&vm.CPU,
			&vm.RAM,
			&vm.Storage,
			&vm.Status,
			&vm.Domain,
			&vm.SSHPublicKey,
			&vm.IsPinned,
			&vm.IdleWarnedAt,
			&vm.CreatedAt,
			&vm.UpdatedAt,
			&vm.StartedAt,
			&vm.StoppedAt,
			&vm.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan VM: %w", err)
		}
		vms = append(vms, vm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate VMs: %w", err)
	}

	return vms, nil
}

// UpdateVMStatus updates the status of a VM
func (r *vmRepository) UpdateVMStatus(ctx context.Context, id, status string) error {
	now := time.Now()

	var query string
	if status == "running" {
		query = `
			UPDATE vms
			SET status = $1, started_at = $2, stopped_at = NULL, updated_at = NOW()
			WHERE id = $3
		`
		_, err := r.db.ExecContext(ctx, query, status, now, id)
		return err
	} else if status == "stopped" {
		query = `
			UPDATE vms
			SET status = $1, stopped_at = $2, updated_at = NOW()
			WHERE id = $3
		`
		_, err := r.db.ExecContext(ctx, query, status, now, id)
		return err
	} else {
		query = `
			UPDATE vms
			SET status = $1, updated_at = NOW()
			WHERE id = $2
		`
		_, err := r.db.ExecContext(ctx, query, status, id)
		return err
	}
}

// UpdateVM updates a VM with the given input
func (r *vmRepository) UpdateVM(ctx context.Context, id string, input VMUpdateInput) error {
	query := `
		UPDATE vms
		SET
			status = COALESCE($1, status),
			k8s_namespace = COALESCE($2, k8s_namespace),
			k8s_deployment = COALESCE($3, k8s_deployment),
			k8s_service = COALESCE($4, k8s_service),
			k8s_pvc = COALESCE($5, k8s_pvc),
			domain = COALESCE($6, domain),
			domain_status = COALESCE($7, domain_status),
			started_at = COALESCE($8, started_at),
			stopped_at = COALESCE($9, stopped_at),
			updated_at = NOW()
		WHERE id = $10
	`

	_, err := r.db.ExecContext(ctx, query,
		input.Status,
		input.K8sNamespace,
		input.K8sDeployment,
		input.K8sService,
		input.K8sPVC,
		input.Domain,
		input.DomainStatus,
		input.StartedAt,
		input.StoppedAt,
		id,
	)

	return err
}

// DeleteVM soft-deletes a VM
func (r *vmRepository) DeleteVM(ctx context.Context, id string) error {
	query := `
		UPDATE vms
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete VM: %w", err)
	}

	return nil
}

// SetPinned sets the pinned status of a VM
func (r *vmRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
	query := `
		UPDATE vms
		SET is_pinned = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, pinned, id)
	return err
}

// GetPinnedCount gets the count of pinned VMs for a user
func (r *vmRepository) GetPinnedCount(ctx context.Context, userID string) int {
	query := `
		SELECT COUNT(*)
		FROM vms
		WHERE user_id = $1 AND is_pinned = true AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// GetIdleVMs gets VMs that have been idle (no status updates) for the specified hours
func (r *vmRepository) GetIdleVMs(ctx context.Context, hours int) ([]*entity.VM, error) {
	query := `
		SELECT id, user_id, name, os, tier, cpu, ram, storage, status, domain, ssh_public_key, created_at, updated_at, started_at, stopped_at, deleted_at
		FROM vms
		WHERE deleted_at IS NULL
		AND updated_at < NOW() - INTERVAL '1 hour' * $1
		AND is_pinned = false
		ORDER BY updated_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, hours)
	if err != nil {
		return nil, fmt.Errorf("get idle VMs: %w", err)
	}
	defer rows.Close()

	var vms []*entity.VM
	for rows.Next() {
		vm := &entity.VM{}
		err := rows.Scan(
			&vm.ID,
			&vm.UserID,
			&vm.Name,
			&vm.OS,
			&vm.Tier,
			&vm.CPU,
			&vm.RAM,
			&vm.Storage,
			&vm.Status,
			&vm.Domain,
			&vm.SSHPublicKey,
			&vm.CreatedAt,
			&vm.UpdatedAt,
			&vm.StartedAt,
			&vm.StoppedAt,
			&vm.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan VM: %w", err)
		}
		vms = append(vms, vm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate VMs: %w", err)
	}

	return vms, nil
}

// SetIdleWarnedAt sets the idle_warned_at timestamp for a VM
func (r *vmRepository) SetIdleWarnedAt(ctx context.Context, id string, warnedAt time.Time) error {
	query := `
		UPDATE vms
		SET idle_warned_at = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, warnedAt, id)
	return err
}

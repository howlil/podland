package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/podland/backend/internal/entity"
)

// AuditRepository defines the interface for audit log data access
type AuditRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	GetRecent(ctx context.Context, limit int) ([]*entity.AuditLog, error)
	GetByUserID(ctx context.Context, userID string, limit int) ([]*entity.AuditLog, error)
}

// auditRepository implements AuditRepository
type auditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sql.DB) AuditRepository {
	return &auditRepository{db: db}
}

// Create creates a new audit log entry
func (r *auditRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, action, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`

	_, err := r.db.ExecContext(ctx, query, log.UserID, log.Action, log.IPAddress, log.UserAgent)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}

	return nil
}

// GetRecent gets recent audit log entries
func (r *auditRepository) GetRecent(ctx context.Context, limit int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, ip_address, user_agent, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}

	return logs, nil
}

// GetByUserID gets audit logs for a specific user
func (r *auditRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get user audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}

	return logs, nil
}

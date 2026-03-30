package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/podland/backend/internal/entity"
)

// NotificationRepository handles database operations for notifications
type NotificationRepository interface {
	Create(ctx context.Context, notification *entity.Notification) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetByVMID(ctx context.Context, vmID uuid.UUID, limit int) ([]*entity.Notification, error)
}

type notificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Create creates a new notification in the database
func (r *notificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, vm_id, alert_name, severity, title, message, is_read, created_at, resolved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var resolvedAt sql.NullTime
	if notification.ResolvedAt != nil {
		resolvedAt = sql.NullTime{Time: *notification.ResolvedAt, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		notification.ID,
		notification.UserID,
		notification.VMID,
		notification.AlertName,
		notification.Severity,
		notification.Title,
		notification.Message,
		notification.IsRead,
		notification.CreatedAt,
		resolvedAt,
	)

	return err
}

// GetByUserID retrieves notifications for a specific user
func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.Notification, error) {
	query := `
		SELECT id, user_id, vm_id, alert_name, severity, title, message, is_read, created_at, resolved_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entity.Notification
	for rows.Next() {
		var n entity.Notification
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.VMID,
			&n.AlertName,
			&n.Severity,
			&n.Title,
			&n.Message,
			&n.IsRead,
			&n.CreatedAt,
			&resolvedAt,
		)
		if err != nil {
			return nil, err
		}

		if resolvedAt.Valid {
			n.ResolvedAt = &resolvedAt.Time
		}

		notifications = append(notifications, &n)
	}

	return notifications, rows.Err()
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// MarkAsRead marks a notification as read
func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("notification not found")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE user_id = $1 AND is_read = false
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// GetByVMID retrieves notifications for a specific VM
func (r *notificationRepository) GetByVMID(ctx context.Context, vmID uuid.UUID, limit int) ([]*entity.Notification, error) {
	query := `
		SELECT id, user_id, vm_id, alert_name, severity, title, message, is_read, created_at, resolved_at
		FROM notifications
		WHERE vm_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, vmID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entity.Notification
	for rows.Next() {
		var n entity.Notification
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.VMID,
			&n.AlertName,
			&n.Severity,
			&n.Title,
			&n.Message,
			&n.IsRead,
			&n.CreatedAt,
			&resolvedAt,
		)
		if err != nil {
			return nil, err
		}

		if resolvedAt.Valid {
			n.ResolvedAt = &resolvedAt.Time
		}

		notifications = append(notifications, &n)
	}

	return notifications, rows.Err()
}

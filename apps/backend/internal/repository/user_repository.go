package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/podland/backend/internal/entity"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	CreateUser(ctx context.Context, input UserCreateInput) (*entity.User, error)
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	GetUserByGitHubID(ctx context.Context, githubID string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateUser(ctx context.Context, id string, input UserUpdateInput) error
	UpdateUserNIM(ctx context.Context, userID, nim string) error
}

// userRepository implements UserRepository
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *userRepository) CreateUser(ctx context.Context, input UserCreateInput) (*entity.User, error) {
	query := `
		INSERT INTO users (github_id, email, display_name, avatar_url, nim, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, github_id, email, display_name, avatar_url, nim, role, created_at, updated_at
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query,
		input.GithubID,
		input.Email,
		input.DisplayName,
		input.AvatarURL,
		input.NIM,
		input.Role,
	).Scan(
		&user.ID,
		&user.GithubID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.NIM,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// GetUserByID gets a user by ID
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, github_id, email, display_name, avatar_url, nim, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.GithubID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.NIM,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByGitHubID gets a user by GitHub ID
func (r *userRepository) GetUserByGitHubID(ctx context.Context, githubID string) (*entity.User, error) {
	query := `
		SELECT id, github_id, email, display_name, avatar_url, nim, role, created_at, updated_at
		FROM users
		WHERE github_id = $1
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, githubID).Scan(
		&user.ID,
		&user.GithubID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.NIM,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by GitHub ID: %w", err)
	}

	return user, nil
}

// GetUserByEmail gets a user by email
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, github_id, email, display_name, avatar_url, nim, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.GithubID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.NIM,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

// UpdateUser updates a user
func (r *userRepository) UpdateUser(ctx context.Context, id string, input UserUpdateInput) error {
	query := `
		UPDATE users
		SET display_name = COALESCE($1, display_name),
			avatar_url = COALESCE($2, avatar_url),
			updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, input.DisplayName, input.AvatarURL, id)
	return err
}

// UpdateUserNIM updates a user's NIM (for confirmation flow)
func (r *userRepository) UpdateUserNIM(ctx context.Context, userID, nim string) error {
	query := `
		UPDATE users
		SET nim = $1,
			role = CASE
				WHEN $1 LIKE '%1152%' THEN 'internal'
				ELSE 'external'
			END,
			updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, nim, userID)
	return err
}

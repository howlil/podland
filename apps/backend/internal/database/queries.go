package database

import (
	"database/sql"
	"encoding/json"
	"time"
)

// CreateUser creates a new user
func (d *sqlDB) CreateUser(input UserCreateInput) (*User, error) {
	query := `
		INSERT INTO users (github_id, email, display_name, avatar_url, nim, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, github_id, email, display_name, avatar_url, nim, role, created_at, updated_at
	`

	user := &User{}
	err := d.db.QueryRow(query,
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
		return nil, err
	}

	return user, nil
}

// GetUserByID gets a user by ID
func (d *sqlDB) GetUserByID(id string) (*User, error) {
	query := `
		SELECT id, github_id, email, display_name, avatar_url, nim, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	err := d.db.QueryRow(query, id).Scan(
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

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByGitHubID gets a user by GitHub ID
func (d *sqlDB) GetUserByGitHubID(githubID string) (*User, error) {
	query := `
		SELECT id, github_id, email, display_name, avatar_url, nim, role, created_at, updated_at
		FROM users
		WHERE github_id = $1
	`

	user := &User{}
	err := d.db.QueryRow(query, githubID).Scan(
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

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail gets a user by email
func (d *sqlDB) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, github_id, email, display_name, avatar_url, nim, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &User{}
	err := d.db.QueryRow(query, email).Scan(
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

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates a user
func (d *sqlDB) UpdateUser(id string, input UserUpdateInput) error {
	query := `
		UPDATE users
		SET display_name = COALESCE($1, display_name),
			avatar_url = COALESCE($2, avatar_url),
			updated_at = NOW()
		WHERE id = $3
	`

	_, err := d.db.Exec(query, input.DisplayName, input.AvatarURL, id)
	return err
}

// CreateSession creates a new session
func (d *sqlDB) CreateSession(session Session) error {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token_hash, jti, device_info, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	deviceInfoJSON, _ := json.Marshal(session.DeviceInfo)

	_, err := d.db.Exec(query,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.JTI,
		deviceInfoJSON,
		session.CreatedAt,
		session.ExpiresAt,
	)

	return err
}

// GetSessionByID gets a session by ID
func (d *sqlDB) GetSessionByID(id string) (*Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, jti, device_info, created_at, expires_at, revoked_at, replaced_by
		FROM sessions
		WHERE id = $1
	`

	session := &Session{}
	var deviceInfoJSON []byte

	err := d.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.JTI,
		&deviceInfoJSON,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.ReplacedBy,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	session.DeviceInfo = deviceInfoJSON
	return session, nil
}

// GetSessionByRefreshToken gets a session by refresh token hash
func (d *sqlDB) GetSessionByRefreshToken(hash string) (*Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, jti, device_info, created_at, expires_at, revoked_at, replaced_by
		FROM sessions
		WHERE refresh_token_hash = $1
	`

	session := &Session{}
	var deviceInfoJSON []byte

	err := d.db.QueryRow(query, hash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.JTI,
		&deviceInfoJSON,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.ReplacedBy,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	session.DeviceInfo = deviceInfoJSON
	return session, nil
}

// GetActiveSessionCount gets the count of active sessions for a user
func (d *sqlDB) GetActiveSessionCount(userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM sessions
		WHERE user_id = $1
		AND revoked_at IS NULL
		AND expires_at > NOW()
	`

	var count int
	err := d.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// RevokeSession revokes a session
func (d *sqlDB) RevokeSession(id string, revokedAt time.Time) error {
	query := `
		UPDATE sessions
		SET revoked_at = $2
		WHERE id = $1
	`

	_, err := d.db.Exec(query, id, revokedAt)
	return err
}

// RevokeOldestSession revokes the oldest session for a user
func (d *sqlDB) RevokeOldestSession(userID string) error {
	query := `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE id = (
			SELECT id
			FROM sessions
			WHERE user_id = $1
			AND revoked_at IS NULL
			AND expires_at > NOW()
			ORDER BY created_at ASC
			LIMIT 1
		)
	`

	_, err := d.db.Exec(query, userID)
	return err
}

// RevokeAllUserSessions revokes all sessions for a user
func (d *sqlDB) RevokeAllUserSessions(userID string) error {
	query := `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE user_id = $1
		AND revoked_at IS NULL
	`

	_, err := d.db.Exec(query, userID)
	return err
}

// LinkSessionReplacement links an old session to its replacement
func (d *sqlDB) LinkSessionReplacement(oldID, newID string) error {
	query := `
		UPDATE sessions
		SET replaced_by = $2
		WHERE id = $1
	`

	_, err := d.db.Exec(query, oldID, newID)
	return err
}

// CreateActivityLog creates an activity log entry
func (d *sqlDB) CreateActivityLog(userID string, action string, metadata map[string]interface{}) error {
	query := `
		INSERT INTO activity_logs (user_id, action, metadata, created_at)
		VALUES ($1, $2, $3, NOW())
	`

	var metadataJSON []byte
	if metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}

	_, err := d.db.Exec(query, userID, action, metadataJSON)
	return err
}

// GetUserActivity gets activity logs for a user
func (d *sqlDB) GetUserActivity(userID string, limit int) ([]ActivityLog, error) {
	query := `
		SELECT id, user_id, action, metadata, created_at
		FROM activity_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := d.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ActivityLog
	for rows.Next() {
		var log ActivityLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.Metadata,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// UpdateUserNIM updates a user's NIM (for confirmation flow)
func (d *sqlDB) UpdateUserNIM(userID, nim string) error {
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

	_, err := d.db.Exec(query, nim, userID)
	return err
}

// Init returns the sql.DB for direct access if needed
func (d *sqlDB) DB() *sql.DB {
	return d.db
}

// NewUserCreateInput creates a UserCreateInput with UUID generation
func NewUserCreateInput(githubID, email, displayName, avatarURL, nim, role string) UserCreateInput {
	return UserCreateInput{
		GithubID:    githubID,
		Email:       email,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		NIM:         nim,
		Role:        role,
	}
}

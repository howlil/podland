package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/sha3"
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

	deviceInfoJSON, err := json.Marshal(session.DeviceInfo)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(query,
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

// RotateSession atomically rotates a session token (revoke old, create new)
func (d *sqlDB) RotateSession(oldTokenHash string, userID string, deviceInfo json.RawMessage) (*Session, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Find the old session
	var oldID string
	query := `SELECT id FROM sessions WHERE refresh_token_hash = $1`
	if err := tx.QueryRow(query, oldTokenHash).Scan(&oldID); err != nil {
		return nil, err
	}

	// Revoke old session
	query = `UPDATE sessions SET revoked_at = NOW() WHERE id = $1`
	if _, err := tx.Exec(query, oldID); err != nil {
		return nil, err
	}

	// Generate new refresh token and JTI
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	newRefreshToken := base64.URLEncoding.EncodeToString(bytes)

	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return nil, err
	}
	newJTI := hex.EncodeToString(jtiBytes)

	// Hash the refresh token for storage
	hash := sha3.Sum256([]byte(newRefreshToken))
	newRefreshTokenHash := hex.EncodeToString(hash[:])

	newSessionID := uuid.New().String()

	query = `
		INSERT INTO sessions (id, user_id, refresh_token_hash, jti, device_info, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW() + INTERVAL '7 days')
	`
	if _, err := tx.Exec(query, newSessionID, userID, newRefreshTokenHash, newJTI, deviceInfo); err != nil {
		return nil, err
	}

	// Link old to new
	query = `UPDATE sessions SET replaced_by = $2 WHERE id = $1`
	if _, err := tx.Exec(query, oldID, newSessionID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return session with raw token for cookie
	return &Session{
		ID:               newSessionID,
		UserID:           userID,
		RefreshTokenHash: newRefreshToken, // Return raw token
		JTI:              newJTI,
		DeviceInfo:       deviceInfo,
		CreatedAt:        time.Now(),
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
	}, nil
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

// CreateVM creates a new VM
func (d *sqlDB) CreateVM(input VMCreateInput) (*VM, error) {
	query := `
		INSERT INTO vms (user_id, name, os, tier, cpu, ram, storage, status, ssh_public_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, NOW(), NOW())
		RETURNING id, user_id, name, os, tier, cpu, ram, storage, status, k8s_namespace, k8s_deployment, k8s_service, k8s_pvc, domain, ssh_public_key, created_at, updated_at, started_at, stopped_at, deleted_at
	`

	vm := &VM{}
	err := d.db.QueryRow(query,
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
		&vm.K8sNamespace,
		&vm.K8sDeployment,
		&vm.K8sService,
		&vm.K8sPVC,
		&vm.Domain,
		&vm.SSHPublicKey,
		&vm.CreatedAt,
		&vm.UpdatedAt,
		&vm.StartedAt,
		&vm.StoppedAt,
		&vm.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return vm, nil
}

// GetVMByID gets a VM by ID
func (d *sqlDB) GetVMByID(id string) (*VM, error) {
	query := `
		SELECT id, user_id, name, os, tier, cpu, ram, storage, status, k8s_namespace, k8s_deployment, k8s_service, k8s_pvc, domain, ssh_public_key, created_at, updated_at, started_at, stopped_at, deleted_at
		FROM vms
		WHERE id = $1 AND deleted_at IS NULL
	`

	vm := &VM{}
	err := d.db.QueryRow(query, id).Scan(
		&vm.ID,
		&vm.UserID,
		&vm.Name,
		&vm.OS,
		&vm.Tier,
		&vm.CPU,
		&vm.RAM,
		&vm.Storage,
		&vm.Status,
		&vm.K8sNamespace,
		&vm.K8sDeployment,
		&vm.K8sService,
		&vm.K8sPVC,
		&vm.Domain,
		&vm.SSHPublicKey,
		&vm.CreatedAt,
		&vm.UpdatedAt,
		&vm.StartedAt,
		&vm.StoppedAt,
		&vm.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// GetVMByIDAndUser gets a VM by ID and user ID (ownership check)
func (d *sqlDB) GetVMByIDAndUser(id, userID string) (*VM, error) {
	query := `
		SELECT id, user_id, name, os, tier, cpu, ram, storage, status, k8s_namespace, k8s_deployment, k8s_service, k8s_pvc, domain, ssh_public_key, created_at, updated_at, started_at, stopped_at, deleted_at
		FROM vms
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	vm := &VM{}
	err := d.db.QueryRow(query, id, userID).Scan(
		&vm.ID,
		&vm.UserID,
		&vm.Name,
		&vm.OS,
		&vm.Tier,
		&vm.CPU,
		&vm.RAM,
		&vm.Storage,
		&vm.Status,
		&vm.K8sNamespace,
		&vm.K8sDeployment,
		&vm.K8sService,
		&vm.K8sPVC,
		&vm.Domain,
		&vm.SSHPublicKey,
		&vm.CreatedAt,
		&vm.UpdatedAt,
		&vm.StartedAt,
		&vm.StoppedAt,
		&vm.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// GetUserVMs gets all VMs for a user
func (d *sqlDB) GetUserVMs(userID string) ([]*VM, error) {
	query := `
		SELECT id, user_id, name, os, tier, cpu, ram, storage, status, k8s_namespace, k8s_deployment, k8s_service, k8s_pvc, domain, ssh_public_key, created_at, updated_at, started_at, stopped_at, deleted_at
		FROM vms
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vms []*VM
	for rows.Next() {
		vm := &VM{}
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
			&vm.K8sNamespace,
			&vm.K8sDeployment,
			&vm.K8sService,
			&vm.K8sPVC,
			&vm.Domain,
			&vm.SSHPublicKey,
			&vm.CreatedAt,
			&vm.UpdatedAt,
			&vm.StartedAt,
			&vm.StoppedAt,
			&vm.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		vms = append(vms, vm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return vms, nil
}

// UpdateVM updates a VM
func (d *sqlDB) UpdateVM(id string, input VMUpdateInput) error {
	query := `
		UPDATE vms
		SET
			status = COALESCE($1, status),
			k8s_namespace = COALESCE($2, k8s_namespace),
			k8s_deployment = COALESCE($3, k8s_deployment),
			k8s_service = COALESCE($4, k8s_service),
			k8s_pvc = COALESCE($5, k8s_pvc),
			domain = COALESCE($6, domain),
			started_at = COALESCE($7, started_at),
			stopped_at = COALESCE($8, stopped_at),
			updated_at = NOW()
		WHERE id = $9
	`

	_, err := d.db.Exec(query,
		input.Status,
		input.K8sNamespace,
		input.K8sDeployment,
		input.K8sService,
		input.K8sPVC,
		input.Domain,
		input.StartedAt,
		input.StoppedAt,
		id,
	)

	return err
}

// UpdateVMStatus updates the status of a VM
func (d *sqlDB) UpdateVMStatus(id, status string) error {
	now := time.Now()

	var query string
	if status == "running" {
		query = `
			UPDATE vms
			SET status = $1, started_at = $2, stopped_at = NULL, updated_at = NOW()
			WHERE id = $3
		`
		_, err := d.db.Exec(query, status, now, id)
		return err
	} else if status == "stopped" {
		query = `
			UPDATE vms
			SET status = $1, stopped_at = $2, updated_at = NOW()
			WHERE id = $3
		`
		_, err := d.db.Exec(query, status, now, id)
		return err
	} else {
		query = `
			UPDATE vms
			SET status = $1, updated_at = NOW()
			WHERE id = $2
		`
		_, err := d.db.Exec(query, status, id)
		return err
	}
}

// DeleteVM soft-deletes a VM
func (d *sqlDB) DeleteVM(id string) error {
	query := `
		UPDATE vms
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := d.db.Exec(query, id)
	return err
}

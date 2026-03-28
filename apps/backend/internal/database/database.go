package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func Init() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

func Migrate(db *sql.DB) error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			github_id VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) NOT NULL,
			display_name VARCHAR(255),
			avatar_url VARCHAR(512),
			nim VARCHAR(20) NOT NULL,
			role VARCHAR(20) NOT NULL CHECK (role IN ('internal', 'external', 'superadmin')),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Users indexes
		`CREATE INDEX IF NOT EXISTS idx_users_github_id ON users(github_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_nim ON users(nim)`,

		// Sessions table
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			refresh_token_hash VARCHAR(255) NOT NULL,
			jti VARCHAR(255) NOT NULL,
			device_info JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMP NOT NULL,
			revoked_at TIMESTAMP,
			replaced_by UUID REFERENCES sessions(id)
		)`,

		// Sessions indexes
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`,

		// Activity logs table
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			action VARCHAR(100) NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Activity logs index
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at)`,
		// Composite index for efficient activity queries (filter by user, order by created_at)
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_user_created ON activity_logs(user_id, created_at DESC)`,

		// Updated_at trigger for users
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
			RETURNS TRIGGER AS $$
			BEGIN
				NEW.updated_at = NOW();
				RETURN NEW;
			END;
			$$ language 'plpgsql'`,

		`DROP TRIGGER IF EXISTS update_users_updated_at ON users`,
		`CREATE TRIGGER update_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column()`,

		// Phase 3: Domain status column for VMs
		`ALTER TABLE vms ADD COLUMN IF NOT EXISTS domain VARCHAR(255)`,
		`ALTER TABLE vms ADD COLUMN IF NOT EXISTS domain_status VARCHAR(20) DEFAULT 'pending' CHECK (domain_status IN ('pending', 'active', 'error'))`,
		`CREATE INDEX IF NOT EXISTS idx_vms_domain ON vms(domain)`,
		`CREATE INDEX IF NOT EXISTS idx_vms_domain_status ON vms(domain_status)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

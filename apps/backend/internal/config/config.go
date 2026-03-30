package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func Load() error {
	// Get the directory where the binary is running
	execDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Try to load .env file from current directory
	envPath := filepath.Join(execDir, ".env")
	if loadErr := godotenv.Load(envPath); loadErr != nil {
		// .env file is optional, but log if not found
		fmt.Printf("Note: .env file not loaded: %v\n", loadErr)
	}

	// Validate required environment variables
	required := []string{
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
		"GITHUB_CALLBACK_URL",
		"JWT_SECRET",
		"REFRESH_TOKEN_SECRET",
		"DATABASE_URL",
		"FRONTEND_URL",
	}

	for _, key := range required {
		if os.Getenv(key) == "" {
			return fmt.Errorf("missing required environment variable: %s", key)
		}
	}

	// Validate JWT secret length
	if len(os.Getenv("JWT_SECRET")) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	if len(os.Getenv("REFRESH_TOKEN_SECRET")) < 32 {
		return fmt.Errorf("REFRESH_TOKEN_SECRET must be at least 32 characters")
	}

	return nil
}

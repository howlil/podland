package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/podland/backend/internal/config"
	"github.com/podland/backend/internal/database"
	"github.com/podland/backend/handlers"
	"github.com/podland/backend/middleware"
)

var db *sql.DB

func main() {
	// Load environment variables
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	var err error
	db, err = database.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize middleware DB
	if err := middleware.InitDB(); err != nil {
		log.Fatalf("Failed to init middleware DB: %v", err)
	}

	// Set db for handlers
	handlers.SetDB(db)

	// Create router
	mux := http.NewServeMux()

	// Static files (avatars)
	fs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	// Auth routes
	mux.HandleFunc("GET /api/auth/login", handlers.HandleLogin)
	mux.HandleFunc("GET /api/auth/github/callback", handlers.HandleCallback)
	mux.HandleFunc("POST /api/auth/refresh", handlers.HandleRefresh)
	mux.HandleFunc("POST /api/auth/logout", handlers.HandleLogout)

	// User routes (protected)
	mux.Handle("GET /api/users/me", middleware.AuthMiddleware(http.HandlerFunc(handlers.HandleGetMe)))
	mux.Handle("GET /api/users/{id}", middleware.AuthMiddleware(http.HandlerFunc(handlers.HandleGetUser)))
	mux.Handle("POST /api/users/confirm-nim", middleware.AuthMiddleware(http.HandlerFunc(handlers.HandleConfirmNIM)))

	// Activity routes (protected)
	mux.Handle("GET /api/activity", middleware.AuthMiddleware(http.HandlerFunc(handlers.HandleGetActivity)))

	// Health check
	mux.HandleFunc("GET /api/health", handlers.HandleHealth)

	// Create server
	addr := fmt.Sprintf(":%s", os.Getenv("PORT"))
	server := &http.Server{
		Addr:         addr,
		Handler:      middleware.CSRFMiddleware(middleware.CORSMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

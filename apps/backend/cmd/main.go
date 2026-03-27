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

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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

	// Create chi router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Static files (avatars)
	fs := http.FileServer(http.Dir("./uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fs))

	// Auth routes
	r.Get("/api/auth/login", handlers.HandleLogin)
	r.Get("/api/auth/github/callback", handlers.HandleCallback)
	r.Post("/api/auth/refresh", handlers.HandleRefresh)
	r.Post("/api/auth/logout", handlers.HandleLogout)
	r.Get("/api/auth/welcome/user", handlers.HandleGetWelcomeUser)

	// User routes (protected)
	r.Route("/api/users", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})(w, r)
			})
		})
		r.Get("/me", handlers.HandleGetMe)
		r.Get("/{id}", handlers.HandleGetUser)
		r.Post("/confirm-nim", handlers.HandleConfirmNIM)
	})

	// Activity routes (protected)
	r.Route("/api/activity", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})(w, r)
			})
		})
		r.Get("/", handlers.HandleGetActivity)
	})

	// VM routes (protected)
	r.Route("/api/vms", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})(w, r)
			})
		})
		r.Post("/", handlers.HandleCreateVM)
		r.Get("/", handlers.HandleListVMs)
		r.Get("/{id}", handlers.HandleGetVM)
		r.Post("/{id}/start", handlers.HandleStartVM)
		r.Post("/{id}/stop", handlers.HandleStopVM)
		r.Post("/{id}/restart", handlers.HandleRestartVM)
		r.Delete("/{id}", handlers.HandleDeleteVM)
	})

	// Health check
	r.Get("/api/health", handlers.HandleHealth)

	// Create server
	addr := fmt.Sprintf(":%s", os.Getenv("PORT"))
	server := &http.Server{
		Addr:         addr,
		Handler:      middleware.CSRFMiddleware(middleware.CORSMiddleware(r)),
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

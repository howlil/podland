package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/podland/backend/internal/cloudflare"
	"github.com/podland/backend/internal/config"
	"github.com/podland/backend/internal/database"
	"github.com/podland/backend/internal/domain"
	"github.com/podland/backend/internal/handler"
	"github.com/podland/backend/internal/middleware"
	"github.com/podland/backend/internal/repository"
	"github.com/podland/backend/internal/usecase"
)

func main() {
	// Load environment variables
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create repositories (dependency injection)
	vmRepo := repository.NewVMRepository(db)
	quotaRepo := repository.NewQuotaRepository(db)
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Create usecases (dependency injection)
	vmUsecase := usecase.NewVMUsecase(vmRepo, quotaRepo, userRepo)

	// Create Cloudflare DNS manager (if credentials are configured)
	var dnsManager *cloudflare.DNSManager
	var dnsPoller *domain.DNSPoller
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" && os.Getenv("CLOUDFLARE_ZONE_ID") != "" {
		dnsManager = cloudflare.NewDNSManager(
			os.Getenv("CLOUDFLARE_API_TOKEN"),
			os.Getenv("CLOUDFLARE_ZONE_ID"),
		)
		dnsPoller = domain.NewDNSPoller(dnsManager, vmRepo)
	}

	// Create handlers (dependency injection)
	vmHandler := handler.NewVMHandler(vmUsecase, vmRepo, dnsManager, dnsPoller)
	authHandler := handler.NewAuthHandler(userRepo, sessionRepo, quotaRepo)

	// Create domain service and handler
	var domainHandler *handler.DomainHandler
	if dnsManager != nil {
		domainService := domain.NewDomainService(dnsManager, db, vmRepo)
		domainHandler = handler.NewDomainHandler(domainService)
	}

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
	r.Get("/api/auth/login", authHandler.HandleLogin)
	r.Get("/api/auth/github/callback", authHandler.HandleCallback)
	r.Post("/api/auth/refresh", authHandler.HandleRefresh)
	r.Post("/api/auth/logout", authHandler.HandleLogout)

	// User routes (protected)
	r.Route("/api/users", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})(w, r)
			})
		})
		r.Get("/me", authHandler.HandleGetMe)
		r.Get("/{id}", authHandler.HandleGetUser)
		r.Post("/confirm-nim", authHandler.HandleConfirmNIM)
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
		r.Get("/", authHandler.HandleGetActivity)
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
		r.Post("/", vmHandler.HandleCreateVM)
		r.Get("/", vmHandler.HandleListVMs)
		r.Get("/{id}", vmHandler.HandleGetVM)
		r.Post("/{id}/start", vmHandler.HandleStartVM)
		r.Post("/{id}/stop", vmHandler.HandleStopVM)
		r.Post("/{id}/restart", vmHandler.HandleRestartVM)
		r.Delete("/{id}", vmHandler.HandleDeleteVM)
	})

	// Domain routes (protected) - only if domain handler is configured
	if domainHandler != nil {
		r.Route("/api/domains", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
						next.ServeHTTP(w, r)
					})(w, r)
				})
			})
			r.Get("/", domainHandler.GetDomains)
			r.Delete("/{id}", domainHandler.DeleteDomain)
		})
	}

	// Health check
	r.Get("/api/health", handler.HandleHealth)

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

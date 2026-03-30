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
	"github.com/podland/backend/internal/email"
	"github.com/podland/backend/internal/handler"
	"github.com/podland/backend/internal/idle"
	"github.com/podland/backend/internal/middleware"
	"github.com/podland/backend/internal/repository"
	"github.com/podland/backend/internal/usecase"
)

// checkRequiredEnvVars validates that all required environment variables are set
func checkRequiredEnvVars() {
	required := []string{
		"DATABASE_URL",
		"JWT_SECRET",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
	}

	// Optional (only required for specific features)
	// - CLOUDFLARE_API_TOKEN, CLOUDFLARE_ZONE_ID: Domain automation
	// - ALERTMANAGER_WEBHOOK_SECRET: Monitoring alerts
	// - SENDGRID_API_KEY, SENDGRID_FROM_EMAIL: Email notifications

	missing := []string{}
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}

	if len(missing) > 0 {
		log.Fatalf("Missing required environment variables: %v", missing)
	}
}

func main() {
	// Load environment variables
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate required environment variables
	checkRequiredEnvVars()

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
	notificationRepo := repository.NewNotificationRepository(db)
	auditRepo := repository.NewAuditRepository(db)

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
	vmHandler := handler.NewVMHandler(vmUsecase, vmRepo, userRepo, dnsManager, dnsPoller)
	authHandler := handler.NewAuthHandler(userRepo, sessionRepo, quotaRepo)
	alertWebhookHandler := handler.NewAlertWebhookHandler(vmRepo, notificationRepo)
	metricsHandler := handler.NewMetricsHandler()
	logsHandler := handler.NewLogsHandler()
	notificationHandler := handler.NewNotificationHandler(notificationRepo)
	adminHandler := handler.NewAdminHandler(userRepo, auditRepo, vmRepo)

	// Create email service
	emailService := email.NewEmailService()

	// Create idle detector and start hourly cron job
	detector := idle.NewDetector(vmRepo, userRepo, notificationRepo, emailService)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			detector.Run()
		}
	}()

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
		// Pin routes
		r.Post("/{id}/pin", vmHandler.HandlePinVM)
		r.Delete("/{id}/pin", vmHandler.HandleUnpinVM)
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

	// Observability routes
	// Alert webhook (internal service only - no auth, uses service token)
	r.Post("/api/alerts/webhook", alertWebhookHandler.HandleAlert)

	// Metrics, logs, and notifications routes (protected)
	// These are nested under /api/vms/{id} for specific VM operations
	r.Route("/api/vms/{id}", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})(w, r)
			})
		})
		r.Get("/metrics", metricsHandler.GetVMMetrics)
		r.Get("/metrics/detail", metricsHandler.RedirectToGrafana)
		r.Get("/logs", logsHandler.GetVMLogs)
		r.Get("/logs/stream", logsHandler.StreamVMLogs)
		r.Get("/alerts", alertWebhookHandler.GetVMAlerts)
	})

	// Notifications routes (protected)
	r.Route("/api/notifications", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})(w, r)
			})
		})
		r.Get("/", notificationHandler.ListNotifications)
		r.Get("/unread-count", notificationHandler.GetUnreadCount)
		r.Post("/{id}/read", notificationHandler.MarkAsRead)
		r.Post("/read-all", notificationHandler.MarkAllAsRead)
	})

	// Admin routes (protected - superadmin only)
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})(w, r)
			})
		})
		r.Use(middleware.AdminOnly(userRepo))    // Require superadmin role
		r.Use(middleware.AuditLogger(auditRepo)) // Auto-log all actions

		r.Get("/users", adminHandler.ListUsers)
		r.Patch("/users/{id}/role", adminHandler.ChangeRole)
		r.Post("/users/{id}/ban", adminHandler.BanUser)
		r.Post("/users/{id}/unban", adminHandler.UnbanUser)
		r.Get("/health", adminHandler.SystemHealth)
		r.Get("/audit-log", adminHandler.AuditLog)
	})

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

// Sentinel NOC - Security Camera Network Operations Center
// Main entry point for the Go/Gin backend API server
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sentinel-noc/internal/config"
	"sentinel-noc/internal/handlers"
	"sentinel-noc/internal/middleware"
	"sentinel-noc/internal/repository"
	"sentinel-noc/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURL)
	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Verify connection
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully")

	// Get database reference
	db := mongoClient.Database(cfg.DBName)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	cameraRepo := repository.NewCameraRepository(db)
	eventRepo := repository.NewEventRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Create indexes
	if err := createIndexes(ctx, db); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Initialize services
	cryptoService := services.NewCryptoService(cfg.EncryptionKey)
	auditService := services.NewAuditService(auditRepo)
	authService := services.NewAuthService(userRepo, cryptoService, auditService, cfg)
	onvifService := services.NewONVIFService()
	webhookService := services.NewWebhookService(settingsRepo)
	alarmService := services.NewAlarmService(cameraRepo, eventRepo, settingsRepo, cryptoService, onvifService, webhookService, auditService)

	// Seed default admin user
	if err := seedAdminUser(ctx, userRepo, cryptoService); err != nil {
		log.Printf("Warning: Failed to seed admin user: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo, cryptoService, auditService, cfg)
	userHandler := handlers.NewUserHandler(userRepo, auditService)
	cameraHandler := handlers.NewCameraHandler(cameraRepo, cryptoService, auditService)
	eventHandler := handlers.NewEventHandler(eventRepo, cameraRepo, auditService)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo, auditService)
	dashboardHandler := handlers.NewDashboardHandler(cameraRepo, eventRepo, settingsRepo)
	systemModeHandler := handlers.NewSystemModeHandler(settingsRepo, cameraRepo, eventRepo, auditService, webhookService)
	onvifHandler := handlers.NewONVIFHandler(cameraRepo, cryptoService, onvifService, alarmService, auditService)

	mediamtxURL := os.Getenv("MEDIAMTX_URL")
	if mediamtxURL == "" {
		mediamtxURL = "http://localhost:9997"
	}
	streamingHandler := handlers.NewStreamingHandler(cameraRepo, cryptoService, mediamtxURL)

	// Initialize Gin router
	router := gin.New()

	// Apply global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.CORSOrigins))
	router.Use(middleware.SecureHeaders())
	router.Use(middleware.RateLimiter(cfg.RateLimitMax, time.Duration(cfg.RateLimitWindow)*time.Second))

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Sentinel NOC API", "version": "1.0.0"})
		})
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy", "timestamp": time.Now().UTC().Format(time.RFC3339)})
		})

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			// Auth routes (protected)
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.POST("/auth/2fa/setup", authHandler.Setup2FA)
			protected.POST("/auth/2fa/verify", authHandler.Verify2FA)
			protected.POST("/auth/2fa/disable", middleware.RequireRole("admin"), authHandler.Disable2FA)

			// Camera routes
			protected.GET("/cameras", cameraHandler.List)
			protected.GET("/cameras/:id", cameraHandler.Get)
			protected.GET("/cameras/:id/stream-url", cameraHandler.GetStreamURL)
			protected.POST("/cameras/:id/status", cameraHandler.UpdateStatus)
			protected.POST("/cameras", middleware.RequireRole("admin", "security_operator"), cameraHandler.Create)
			protected.PUT("/cameras/:id", middleware.RequireRole("admin", "security_operator"), cameraHandler.Update)
			protected.DELETE("/cameras/:id", middleware.RequireRole("admin"), cameraHandler.Delete)

			// Streaming routes
			protected.POST("/cameras/:id/stream/start", streamingHandler.StartStream)
			protected.POST("/cameras/:id/stream/stop", middleware.RequireRole("admin", "security_operator"), streamingHandler.StopStream)
			protected.GET("/cameras/:id/stream/status", streamingHandler.GetStreamStatus)

			// ONVIF routes
			protected.POST("/cameras/:id/onvif/detect", middleware.RequireRole("admin", "security_operator"), onvifHandler.DetectCapabilities)
			protected.POST("/cameras/:id/onvif/test", middleware.RequireRole("admin", "security_operator"), onvifHandler.TestConnection)
			protected.POST("/cameras/:id/onvif/credentials", middleware.RequireRole("admin", "security_operator"), onvifHandler.UpdateONVIFCredentials)
			protected.POST("/cameras/:id/alarm/trigger", middleware.RequireRole("admin", "security_operator"), onvifHandler.TriggerAlarm)
			protected.POST("/cameras/:id/alarm/stop", middleware.RequireRole("admin", "security_operator"), onvifHandler.StopAlarm)

			// System mode routes
			protected.GET("/system/mode", systemModeHandler.GetMode)
			protected.PUT("/system/mode", middleware.RequireRole("admin", "security_operator"), systemModeHandler.SetMode)

			// Event routes
			protected.GET("/events", eventHandler.List)
			protected.POST("/events", eventHandler.Create)
			protected.POST("/events/:id/acknowledge", middleware.RequireRole("admin", "security_operator"), eventHandler.Acknowledge)

			// User management routes (admin only)
			protected.GET("/users", middleware.RequireRole("admin"), userHandler.List)
			protected.PUT("/users/:id/role", middleware.RequireRole("admin"), userHandler.UpdateRole)
			protected.DELETE("/users/:id", middleware.RequireRole("admin"), userHandler.Delete)

			// Settings routes (admin only)
			protected.GET("/settings", middleware.RequireRole("admin"), settingsHandler.Get)
			protected.PUT("/settings", middleware.RequireRole("admin"), settingsHandler.Update)

			// Audit logs (admin only)
			protected.GET("/audit-logs", middleware.RequireRole("admin"), auditService.ListHandler)

			// Dashboard
			protected.GET("/dashboard/stats", dashboardHandler.GetStats)
		}
	}

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting Sentinel NOC API server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if err := mongoClient.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
	}

	log.Println("Server exited gracefully")
}

// createIndexes creates database indexes for optimal query performance
func createIndexes(ctx context.Context, db *mongo.Database) error {
	// Users collection indexes
	userIndexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"username": 1}, Options: options.Index().SetUnique(true)},
		{Keys: map[string]interface{}{"email": 1}, Options: options.Index().SetUnique(true)},
	}
	if _, err := db.Collection("users").Indexes().CreateMany(ctx, userIndexes); err != nil {
		return err
	}

	// Cameras collection indexes
	cameraIndexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"ip_address": 1}, Options: options.Index().SetUnique(true)},
	}
	if _, err := db.Collection("cameras").Indexes().CreateMany(ctx, cameraIndexes); err != nil {
		return err
	}

	// Events collection indexes
	eventIndexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"created_at": -1}},
		{Keys: map[string]interface{}{"camera_id": 1}},
		{Keys: map[string]interface{}{"severity": 1, "acknowledged": 1}},
	}
	if _, err := db.Collection("events").Indexes().CreateMany(ctx, eventIndexes); err != nil {
		return err
	}

	// Audit logs collection indexes
	auditIndexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"created_at": -1}},
		{Keys: map[string]interface{}{"user_id": 1}},
		{Keys: map[string]interface{}{"action": 1}},
	}
	if _, err := db.Collection("audit_logs").Indexes().CreateMany(ctx, auditIndexes); err != nil {
		return err
	}

	log.Println("Database indexes created successfully")
	return nil
}

// seedAdminUser creates a default admin user if one doesn't exist
func seedAdminUser(ctx context.Context, userRepo *repository.UserRepository, cryptoService *services.CryptoService) error {
	exists, err := userRepo.ExistsByUsername(ctx, "admin")
	if err != nil {
		return err
	}

	if !exists {
		hashedPassword, err := cryptoService.HashPassword("P@ssw0rd!")
		if err != nil {
			return err
		}

		adminUser := &repository.User{
			ID:                  services.GenerateUUID(),
			Username:            "admin",
			Email:               "admin@sentinel-noc.local",
			PasswordHash:        hashedPassword,
			Role:                "admin",
			TOTPEnabled:         false,
			CreatedAt:           time.Now().UTC(),
			FailedLoginAttempts: 0,
		}

		if err := userRepo.Create(ctx, adminUser); err != nil {
			return err
		}
		log.Println("Default admin user seeded. Change the password immediately via the Settings page.")
	}

	return nil
}

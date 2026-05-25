// Package config handles application configuration management
package config

import (
	"log"
	"os"
	"strconv"
)

const devJWTSecret = "dev-only-jwt-secret-do-not-use-in-production-sentinel-noc"

// Config holds all application configuration
type Config struct {
	// Server settings
	Port        string
	Environment string

	// Database settings
	MongoURL string
	DBName   string

	// Security settings
	JWTSecret               string
	EncryptionKey           string
	AccessTokenExpireMin    int
	RefreshTokenExpireDays  int

	// Rate limiting
	RateLimitMax    int
	RateLimitWindow int // seconds

	// CORS
	CORSOrigins string
}

// Load reads configuration from environment variables with secure defaults
func Load() *Config {
	env := getEnv("ENVIRONMENT", "development")
	isProd := env == "production"

	jwtSecret := os.Getenv("JWT_SECRET")
	encryptionKey := os.Getenv("ENCRYPTION_KEY")

	if isProd {
		if jwtSecret == "" {
			log.Fatal("FATAL: JWT_SECRET must be set in production")
		}
		if encryptionKey == "" {
			log.Fatal("FATAL: ENCRYPTION_KEY must be set in production")
		}
	}

	// Development-only fallbacks — never used when ENVIRONMENT=production
	if jwtSecret == "" {
		jwtSecret = devJWTSecret
		log.Println("WARNING: JWT_SECRET not set, using insecure development default")
	}
	if encryptionKey == "" {
		// Zero-filled key causes an error on first Encrypt call, surfacing the misconfiguration early
		encryptionKey = ""
		log.Println("WARNING: ENCRYPTION_KEY not set; encryption will fail until configured")
	}

	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:3000")
	if isProd && corsOrigins == "*" {
		log.Fatal("FATAL: CORS_ORIGINS must not be wildcard (*) in production")
	}
	if !isProd && corsOrigins == "*" {
		log.Println("WARNING: CORS allows all origins — acceptable only in development")
	}

	cfg := &Config{
		Port:        getEnv("PORT", "8001"),
		Environment: env,

		MongoURL: getEnv("MONGO_URL", "mongodb://localhost:27017"),
		DBName:   getEnv("DB_NAME", "sentinel_noc"),

		JWTSecret:              jwtSecret,
		EncryptionKey:          encryptionKey,
		AccessTokenExpireMin:   getEnvInt("ACCESS_TOKEN_EXPIRE_MINUTES", 30),
		RefreshTokenExpireDays: getEnvInt("REFRESH_TOKEN_EXPIRE_DAYS", 7),

		RateLimitMax:    getEnvInt("RATE_LIMIT_MAX", 100),
		RateLimitWindow: getEnvInt("RATE_LIMIT_WINDOW", 60),

		CORSOrigins: corsOrigins,
	}

	return cfg
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

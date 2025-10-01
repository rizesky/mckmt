package main

import (
	"fmt"
	"log"
	"strings"

	_ "github.com/rizesky/mckmt/docs" // Import docs for Swagger
	"github.com/rizesky/mckmt/internal/app/hub"
	"github.com/rizesky/mckmt/internal/config"
)

// @title MCKMT Hub API
// @version 1.0
// @description This is the API for the MCKMT Hub application.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @externalDocs.description OpenAPI
// @externalDocs.url https://swagger.io/resources/open-api/
func main() {
	// Load configuration
	cfg, err := config.LoadHubConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Display configuration and environment variables
	displayConfiguration(cfg)

	// Create and run hub application
	hubApp := hub.New(cfg)
	if err := hubApp.Run(); err != nil {
		log.Fatalf("Hub failed: %v", err)
	}
}

// displayConfiguration shows all configuration values being used
func displayConfiguration(cfg *config.HubConfig) {
	fmt.Println("=== MCKMT Configuration ===")

	// Server Configuration
	fmt.Println("\n🖥️  Server Configuration:")
	fmt.Printf("  Host: %s\n", cfg.Server.Host)
	fmt.Printf("  Port: %d\n", cfg.Server.Port)
	fmt.Printf("  Read Timeout: %s\n", cfg.Server.ReadTimeout)
	fmt.Printf("  Write Timeout: %s\n", cfg.Server.WriteTimeout)
	fmt.Printf("  Idle Timeout: %s\n", cfg.Server.IdleTimeout)

	// gRPC Configuration
	fmt.Println("\n🔌 gRPC Configuration:")
	fmt.Printf("  Host: %s\n", cfg.GRPC.Host)
	fmt.Printf("  Port: %d\n", cfg.GRPC.Port)
	fmt.Printf("  Read Timeout: %s\n", cfg.GRPC.ReadTimeout)
	fmt.Printf("  Write Timeout: %s\n", cfg.GRPC.WriteTimeout)
	fmt.Printf("  Idle Timeout: %s\n", cfg.GRPC.IdleTimeout)

	// Database Configuration
	fmt.Println("\n🗄️  Database Configuration:")
	fmt.Printf("  Host: %s\n", cfg.Database.Host)
	fmt.Printf("  Port: %d\n", cfg.Database.Port)
	fmt.Printf("  User: %s\n", cfg.Database.User)
	fmt.Printf("  Database: %s\n", cfg.Database.Database)
	fmt.Printf("  SSL Mode: %s\n", cfg.Database.SSLMode)

	// Redis Configuration
	fmt.Println("\n🔴 Redis Configuration:")
	fmt.Printf("  Host: %s\n", cfg.Redis.Host)
	fmt.Printf("  Port: %d\n", cfg.Redis.Port)
	fmt.Printf("  Database: %d\n", cfg.Redis.DB)

	// Authentication Configuration
	fmt.Println("\n🔐 Authentication Configuration:")
	fmt.Printf("  OIDC Enabled: %t\n", cfg.Auth.OIDC.Enabled)
	fmt.Printf("  OIDC Issuer: %s\n", cfg.Auth.OIDC.Issuer)
	fmt.Printf("  OIDC Client ID: %s\n", cfg.Auth.OIDC.ClientID)
	fmt.Printf("  OIDC Redirect URL: %s\n", cfg.Auth.OIDC.RedirectURL)
	fmt.Printf("  OIDC Scopes: %v\n", cfg.Auth.OIDC.Scopes)

	// JWT Configuration
	fmt.Println("\n🎫 JWT Configuration:")
	fmt.Printf("  Secret: %s\n", maskSecret(cfg.Auth.JWT.Secret))
	fmt.Printf("  Expiration: %s\n", cfg.Auth.JWT.Expiration)
	fmt.Printf("  Issuer: %s\n", cfg.Auth.JWT.Issuer)
	fmt.Printf("  Audience: %s\n", cfg.Auth.JWT.Audience)

	// RBAC Configuration
	fmt.Println("\n👥 RBAC Configuration:")
	fmt.Printf("  Strategy: %s\n", cfg.Auth.RBAC.Strategy)
	fmt.Printf("  Default Role: %s\n", cfg.Auth.RBAC.DefaultRole)
	fmt.Printf("  Casbin Enabled: %t\n", cfg.Auth.RBAC.Casbin.Enabled)
	fmt.Printf("  Casbin Model File: %s\n", cfg.Auth.RBAC.Casbin.ModelFile)
	fmt.Printf("  Casbin Policy File: %s\n", cfg.Auth.RBAC.Casbin.PolicyFile)
	fmt.Printf("  Casbin Auto Reload: %t\n", cfg.Auth.RBAC.Casbin.AutoReload)
	fmt.Printf("  Casbin Reload Interval: %d seconds\n", cfg.Auth.RBAC.Casbin.ReloadInterval)

	// Logging Configuration
	fmt.Println("\n📝 Logging Configuration:")
	fmt.Printf("  Level: %s\n", cfg.Logging.Level)
	fmt.Printf("  Format: %s\n", cfg.Logging.Format)

	// Metrics Configuration
	fmt.Println("\n📊 Metrics Configuration:")
	fmt.Printf("  Enabled: %t\n", cfg.Metrics.Enabled)
	fmt.Printf("  Port: %d\n", cfg.Metrics.Port)

	fmt.Println("\n🚀 Starting MCKMT Hub...")
	fmt.Println("=" + strings.Repeat("=", 50))
}

// maskSecret masks sensitive configuration values
func maskSecret(secret string) string {
	if secret == "" {
		return "***not set***"
	}
	if len(secret) <= 8 {
		return "***masked***"
	}
	return secret[:4] + "***masked***" + secret[len(secret)-4:]
}

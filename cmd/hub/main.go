package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	agentv1 "github.com/rizesky/mckmt/api/proto/agent/v1"
	_ "github.com/rizesky/mckmt/docs" // Import generated docs
	grpcserver "github.com/rizesky/mckmt/internal/api/grpc"
	apihandler "github.com/rizesky/mckmt/internal/api/http"
	"github.com/rizesky/mckmt/internal/app/clusters"
	"github.com/rizesky/mckmt/internal/app/operations"
	"github.com/rizesky/mckmt/internal/auth"
	"github.com/rizesky/mckmt/internal/config"
	"github.com/rizesky/mckmt/internal/metrics"
	"github.com/rizesky/mckmt/internal/orchestrator"
	"github.com/rizesky/mckmt/internal/repo/postgres"
	rediscache "github.com/rizesky/mckmt/internal/repo/redis"
)

// @title MCKMT Hub API
// @version 1.0
// @description Multi-Cluster Kubernetes Management API - Hub Service
// @termsOfService https://github.com/rizesky/mckmt/blob/main/LICENSE

// @contact.name API Support
// @contact.url https://github.com/rizesky/mckmt/issues
// @contact.email support@mckmt.dev

// @license.name MIT
// @license.url https://github.com/rizesky/mckmt/blob/main/LICENSE

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting MCKMT Hub", zap.String("version", "0.1.0"))

	// Initialize database
	db, err := postgres.NewDatabase(cfg.Database.DSN(), logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + fmt.Sprintf("%d", cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// Initialize cache manager
	cacheManager := rediscache.NewCacheAdapter(redisClient, logger)

	// Initialize metrics
	metricsManager := metrics.NewMetrics()

	// Initialize repositories
	clusterRepo := postgres.NewClusterRepository(db)
	operationRepo := postgres.NewOperationRepository(db)

	// Initialize cached repositories for better performance
	cachedClusterRepo := postgres.NewCachedClusterRepository(clusterRepo, cacheManager, metricsManager, logger)
	cachedOperationRepo := postgres.NewCachedOperationRepository(operationRepo, cacheManager, metricsManager, logger)

	// Initialize orchestrator (needed for operation service)
	orchestrator := orchestrator.NewOrchestrator(cachedOperationRepo, metricsManager, logger, cfg.Orchestrator.Workers)

	// Initialize app services
	clusterService := clusters.NewClusterService(cachedClusterRepo, cachedOperationRepo, cacheManager, logger, orchestrator)
	operationService := operations.NewOperationService(cachedOperationRepo, cacheManager, logger, orchestrator)

	// Initialize auth middleware
	jwtManager := auth.NewJWTManager(cfg.Auth.JWT.Secret, cfg.Auth.JWT.Expiration)
	authMiddleware := auth.NewAuthMiddleware(jwtManager, logger)

	// Initialize HTTP router
	router := apihandler.NewRouter(clusterService, operationService, logger, authMiddleware, cfg)

	// Initialize gRPC server
	grpcService := grpcserver.NewServer(cachedClusterRepo, cachedOperationRepo, metricsManager, logger)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    cfg.Server.Addr(),
		Handler: router.SetupRoutes(),
	}

	// Create gRPC server
	grpcListener, err := net.Listen("tcp", cfg.GRPC.Addr())
	if err != nil {
		logger.Fatal("Failed to listen on gRPC port", zap.Error(err))
	}
	grpcServer := grpc.NewServer()
	agentv1.RegisterAgentServiceServer(grpcServer, grpcService)

	// Start orchestrator
	go func() {
		if err := orchestrator.Start(ctx); err != nil {
			logger.Error("Orchestrator failed", zap.Error(err))
		}
	}()

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server", zap.String("addr", cfg.Server.Addr()))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Start gRPC server
	go func() {
		logger.Info("Starting gRPC server", zap.String("addr", cfg.GRPC.Addr()))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop orchestrator
	orchestrator.Stop()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown failed", zap.Error(err))
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	logger.Info("Servers stopped")
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	// TODO: Add environment and log level configuration
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	return config.Build()
}

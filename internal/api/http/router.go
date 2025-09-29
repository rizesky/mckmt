package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/app/clusters"
	"github.com/rizesky/mckmt/internal/app/operations"
	"github.com/rizesky/mckmt/internal/auth"
	"github.com/rizesky/mckmt/internal/config"
)

// Router composes all handlers and sets up routes
type Router struct {
	clusterHandler   *ClusterHandler
	operationHandler *OperationHandler
	systemHandler    *SystemHandler
	logger           *zap.Logger
	authMiddleware   *auth.AuthMiddleware
	cfg              *config.Config
}

// NewRouter creates a new router with all handlers
func NewRouter(
	clusterService *clusters.ClusterService,
	operationService *operations.OperationService,
	logger *zap.Logger,
	authMiddleware *auth.AuthMiddleware,
	cfg *config.Config,
) *Router {
	return &Router{
		clusterHandler:   NewClusterHandler(clusterService, logger),
		operationHandler: NewOperationHandler(operationService, logger),
		systemHandler:    NewSystemHandler(logger),
		logger:           logger,
		authMiddleware:   authMiddleware,
		cfg:              cfg,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() chi.Router {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// CORS
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			if req.Method == "OPTIONS" {
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	// API routes
	router.Route("/api/v1", func(apiRouter chi.Router) {
		// System routes (no auth required)
		apiRouter.Get("/health", r.systemHandler.HealthCheck)
		apiRouter.Get("/metrics", r.systemHandler.Metrics)

		// Swagger documentation
		apiRouter.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
		))

		// Register domain-specific routes
		r.clusterHandler.RegisterRoutes(apiRouter)
		r.operationHandler.RegisterRoutes(apiRouter)
	})

	return router
}

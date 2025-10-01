package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/auth"
	"github.com/rizesky/mckmt/internal/cluster"
	"github.com/rizesky/mckmt/internal/config"
	"github.com/rizesky/mckmt/internal/metrics"
	"github.com/rizesky/mckmt/internal/operation"
)

// Router composes all handlers and sets up routes
type Router struct {
	clusterHandler   *ClusterHandler
	operationHandler *OperationHandler
	systemHandler    *SystemHandler
	authHandler      *AuthHandler
	logger           *zap.Logger
	authMiddleware   *auth.Middleware
	cfg              *config.HubConfig
	metrics          *metrics.Metrics
	authzService     *auth.AuthorizationService
}

// NewRouter creates a new router with all handlers
func NewRouter(
	clusterService *cluster.Service,
	operationService *operation.Service,
	authService *auth.Service,
	logger *zap.Logger,
	authMiddleware *auth.Middleware,
	cfg *config.HubConfig,
	metricsMgr *metrics.Metrics,
	authzService *auth.AuthorizationService,
) *Router {
	return &Router{
		clusterHandler:   NewClusterHandler(clusterService, logger),
		operationHandler: NewOperationHandler(operationService, logger),
		systemHandler:    NewSystemHandler(logger),
		authHandler:      NewAuthHandler(authService, logger),
		logger:           logger,
		authMiddleware:   authMiddleware,
		cfg:              cfg,
		metrics:          metricsMgr,
		authzService:     authzService,
	}
}

// SetupRoutes configures all routes using Chi with proper grouping
func (r *Router) SetupRoutes() chi.Router {
	router := chi.NewRouter()

	// Apply common middleware to all routes
	r.applyCommonMiddlewares(router)

	// Swagger documentation
	router.Get("/swagger/*", httpSwagger.Handler())

	// API routes with versioning
	router.Route("/api/v1", func(api chi.Router) {
		// System routes (no auth required)
		r.registerSystemRoutes(api)

		// Authentication routes (no auth required)
		r.registerAuthRoutes(api)

		// Protected routes (require authentication)
		api.Group(func(protected chi.Router) {
			protected.Use(r.authMiddleware.RequireAuth)
			r.registerProtectedRoutes(protected)
		})
	})

	return router
}

// applyCommonMiddlewares applies common middleware to all routes
func (r *Router) applyCommonMiddlewares(router chi.Router) {
	// Chi built-in middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Custom middleware
	router.Use(r.corsMiddleware)
	router.Use(r.metricsMiddleware)
}

// registerSystemRoutes registers system routes that don't require authentication
func (r *Router) registerSystemRoutes(router chi.Router) {
	router.Route("/", func(system chi.Router) {
		system.Get("/health", r.systemHandler.HealthCheck)
		system.Get("/metrics", r.systemHandler.Metrics)
	})
}

// registerAuthRoutes registers authentication routes that don't require authentication
func (r *Router) registerAuthRoutes(router chi.Router) {
	router.Route("/auth", func(auth chi.Router) {
		// Auth methods discovery
		auth.Get("/methods", r.authHandler.GetAuthMethods)

		// OIDC routes
		auth.Get("/oidc/login", r.authHandler.OIDCLogin)
		auth.Get("/oidc/callback", r.authHandler.OIDCCallback)
		auth.Post("/oidc/logout", r.authHandler.OIDCLogout)

		// Custom auth routes
		auth.Post("/login", r.authHandler.Login)
		auth.Post("/register", r.authHandler.Register)
		auth.Post("/refresh", r.authHandler.RefreshToken)
		auth.Post("/logout", r.authHandler.Logout)
		auth.Post("/change-password", r.authHandler.ChangePassword)
	})
}

// registerProtectedRoutes registers routes that require authentication
func (r *Router) registerProtectedRoutes(router chi.Router) {
	// Protected auth routes (user profile)
	router.Get("/auth/profile", r.authHandler.GetProfile)

	// Cluster routes with Casbin permissions
	router.Route("/clusters", func(clusters chi.Router) {
		clusters.Get("/", r.authMiddleware.RequirePermission(r.authzService, "clusters", "read")(r.clusterHandler.ListClusters))
		clusters.Get("/{id}", r.authMiddleware.RequirePermission(r.authzService, "clusters", "read")(r.clusterHandler.GetCluster))
		clusters.Put("/{id}", r.authMiddleware.RequirePermission(r.authzService, "clusters", "write")(r.clusterHandler.UpdateCluster))
		clusters.Delete("/{id}", r.authMiddleware.RequirePermission(r.authzService, "clusters", "delete")(r.clusterHandler.DeleteCluster))
		clusters.Get("/{id}/resources", r.authMiddleware.RequirePermission(r.authzService, "clusters", "read")(r.clusterHandler.ListClusterResources))
		clusters.Post("/{id}/manifests", r.authMiddleware.RequirePermission(r.authzService, "clusters", "manage")(r.clusterHandler.ApplyManifests))
	})

	// Operation routes with Casbin permissions
	router.Route("/operations", func(operations chi.Router) {
		operations.Get("/{id}", r.authMiddleware.RequirePermission(r.authzService, "operations", "read")(r.operationHandler.GetOperation))
		operations.Get("/cluster/{clusterId}", r.authMiddleware.RequirePermission(r.authzService, "operations", "read")(r.operationHandler.ListOperationsByCluster))
		operations.Post("/{id}/cancel", r.authMiddleware.RequirePermission(r.authzService, "operations", "cancel")(r.operationHandler.CancelOperation))
	})
}

// Middleware functions
func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		if req.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (r *Router) metricsMiddleware(next http.Handler) http.Handler {
	if r.metrics != nil {
		return metrics.HTTPMiddlewareFactory(r.metrics, r.logger)(next)
	}
	return next
}

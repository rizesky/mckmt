package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// SystemHandler handles system-related HTTP requests (health, metrics, etc.)
type SystemHandler struct {
	logger *zap.Logger
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(logger *zap.Logger) *SystemHandler {
	return &SystemHandler{
		logger: logger,
	}
}

// RegisterRoutes registers system routes
func (h *SystemHandler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.HealthCheck)
	r.Get("/metrics", h.Metrics)
}

// HealthCheck handles health check endpoint
func (h *SystemHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

// Metrics handles metrics endpoint
func (h *SystemHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	// Serve Prometheus metrics
	promhttp.Handler().ServeHTTP(w, r)
}

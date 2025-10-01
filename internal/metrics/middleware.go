package metrics

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPMiddlewareFactory creates HTTP middleware with optional metrics
func HTTPMiddlewareFactory(metrics *Metrics, logger *zap.Logger) func(http.Handler) http.Handler {
	if metrics == nil {
		// Return no-op middleware if no metrics
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			method := req.Method
			endpoint := req.URL.Path

			// Increment in-flight requests
			metrics.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Inc()

			// Wrap response writer to capture status code
			rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rr, req)

			// Record metrics
			duration := time.Since(start).Seconds()
			statusCode := rr.statusCode
			statusText := http.StatusText(statusCode)

			metrics.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusText).Inc()
			metrics.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
			metrics.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Dec()
		})
	}
}

// responseRecorder wraps http.ResponseWriter to capture status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

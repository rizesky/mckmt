package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight *prometheus.GaugeVec

	// Cluster metrics
	ClustersTotal   *prometheus.GaugeVec
	ClusterStatus   *prometheus.GaugeVec
	ClusterLastSeen *prometheus.GaugeVec

	// Operation metrics
	OperationsTotal      *prometheus.CounterVec
	OperationsInProgress *prometheus.GaugeVec
	OperationDuration    *prometheus.HistogramVec

	// Agent metrics
	AgentsConnected    *prometheus.GaugeVec
	AgentHeartbeats    *prometheus.CounterVec
	AgentLastHeartbeat *prometheus.GaugeVec

	// Database metrics
	DatabaseConnections   *prometheus.GaugeVec
	DatabaseQueryDuration *prometheus.HistogramVec

	// Cache metrics
	CacheHits       *prometheus.CounterVec
	CacheMisses     *prometheus.CounterVec
	CacheOperations *prometheus.CounterVec
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mckmt_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mckmt_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mckmt_http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
			[]string{"method", "endpoint"},
		),

		// Cluster metrics
		ClustersTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mckmt_clusters_total",
				Help: "Total number of registered clusters",
			},
			[]string{"mode", "status"},
		),
		ClusterStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mckmt_cluster_status",
				Help: "Cluster status (1=connected, 0=disconnected)",
			},
			[]string{"cluster_id", "cluster_name", "mode"},
		),
		ClusterLastSeen: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mckmt_cluster_last_seen_timestamp",
				Help: "Timestamp of last cluster heartbeat",
			},
			[]string{"cluster_id", "cluster_name"},
		),

		// Operation metrics
		OperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mckmt_operations_total",
				Help: "Total number of operations",
			},
			[]string{"cluster_id", "type", "status"},
		),
		OperationsInProgress: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mckmt_operations_in_progress",
				Help: "Current number of operations in progress",
			},
			[]string{"cluster_id", "type"},
		),
		OperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mckmt_operation_duration_seconds",
				Help:    "Operation duration in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
			},
			[]string{"cluster_id", "type", "status"},
		),

		// Agent metrics
		AgentsConnected: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mckmt_agents_connected",
				Help: "Number of connected agents",
			},
			[]string{"cluster_id", "agent_version"},
		),
		AgentHeartbeats: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mckmt_agent_heartbeats_total",
				Help: "Total number of agent heartbeats",
			},
			[]string{"cluster_id", "status"},
		),
		AgentLastHeartbeat: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mckmt_agent_last_heartbeat_timestamp",
				Help: "Timestamp of last agent heartbeat",
			},
			[]string{"cluster_id"},
		),

		// Database metrics
		DatabaseConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mckmt_database_connections",
				Help: "Current number of database connections",
			},
			[]string{"state"},
		),
		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mckmt_database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),

		// Cache metrics
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mckmt_cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type", "key_pattern"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mckmt_cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type", "key_pattern"},
		),
		CacheOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mckmt_cache_operations_total",
				Help: "Total number of cache operations",
			},
			[]string{"operation", "cache_type"},
		),
	}
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// IncHTTPRequestsInFlight increments in-flight HTTP requests
func (m *Metrics) IncHTTPRequestsInFlight(method, endpoint string) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Inc()
}

// DecHTTPRequestsInFlight decrements in-flight HTTP requests
func (m *Metrics) DecHTTPRequestsInFlight(method, endpoint string) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Dec()
}

// SetClustersTotal sets the total number of clusters
func (m *Metrics) SetClustersTotal(mode, status string, count float64) {
	m.ClustersTotal.WithLabelValues(mode, status).Set(count)
}

// SetClusterStatus sets the cluster status
func (m *Metrics) SetClusterStatus(clusterID, clusterName, mode, status string) {
	value := 0.0
	if status == "connected" {
		value = 1.0
	}
	m.ClusterStatus.WithLabelValues(clusterID, clusterName, mode).Set(value)
}

// SetClusterLastSeen sets the cluster last seen timestamp
func (m *Metrics) SetClusterLastSeen(clusterID, clusterName string, timestamp float64) {
	m.ClusterLastSeen.WithLabelValues(clusterID, clusterName).Set(timestamp)
}

// RecordOperation records an operation
func (m *Metrics) RecordOperation(clusterID, operationType, status string, duration float64) {
	m.OperationsTotal.WithLabelValues(clusterID, operationType, status).Inc()
	if duration > 0 {
		m.OperationDuration.WithLabelValues(clusterID, operationType, status).Observe(duration)
	}
}

// IncOperationsInProgress increments operations in progress
func (m *Metrics) IncOperationsInProgress(clusterID, operationType string) {
	m.OperationsInProgress.WithLabelValues(clusterID, operationType).Inc()
}

// DecOperationsInProgress decrements operations in progress
func (m *Metrics) DecOperationsInProgress(clusterID, operationType string) {
	m.OperationsInProgress.WithLabelValues(clusterID, operationType).Dec()
}

// SetAgentsConnected sets the number of connected agents
func (m *Metrics) SetAgentsConnected(clusterID, agentVersion string, count float64) {
	m.AgentsConnected.WithLabelValues(clusterID, agentVersion).Set(count)
}

// RecordAgentHeartbeat records an agent heartbeat
func (m *Metrics) RecordAgentHeartbeat(clusterID, status string) {
	m.AgentHeartbeats.WithLabelValues(clusterID, status).Inc()
}

// SetAgentLastHeartbeat sets the last heartbeat timestamp
func (m *Metrics) SetAgentLastHeartbeat(clusterID string, timestamp float64) {
	m.AgentLastHeartbeat.WithLabelValues(clusterID).Set(timestamp)
}

// SetDatabaseConnections sets the number of database connections
func (m *Metrics) SetDatabaseConnections(state string, count float64) {
	m.DatabaseConnections.WithLabelValues(state).Set(count)
}

// RecordDatabaseQuery records a database query
func (m *Metrics) RecordDatabaseQuery(operation, table string, duration float64) {
	m.DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit(cacheType, keyPattern string) {
	m.CacheHits.WithLabelValues(cacheType, keyPattern).Inc()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss(cacheType, keyPattern string) {
	m.CacheMisses.WithLabelValues(cacheType, keyPattern).Inc()
}

// RecordCacheOperation records a cache operation
func (m *Metrics) RecordCacheOperation(operation, cacheType string) {
	m.CacheOperations.WithLabelValues(operation, cacheType).Inc()
}

// IncrementCounter increments a counter metric
func (m *Metrics) IncrementCounter(name string, tags map[string]string) {
	// This is a generic method - you might want to implement specific counters
	// For now, we'll use a default counter
	m.HTTPRequestsTotal.WithLabelValues("unknown", "unknown", "unknown").Inc()
}

// IncrementGauge increments a gauge metric
func (m *Metrics) IncrementGauge(name string, value float64, tags map[string]string) {
	// This is a generic method - you might want to implement specific gauges
	// For now, we'll use a default gauge
	m.HTTPRequestsInFlight.WithLabelValues("unknown").Add(value)
}

// RecordHistogram records a histogram metric
func (m *Metrics) RecordHistogram(name string, value float64, tags map[string]string) {
	// This is a generic method - you might want to implement specific histograms
	// For now, we'll use a default histogram
	m.HTTPRequestDuration.WithLabelValues("unknown", "unknown").Observe(value)
}

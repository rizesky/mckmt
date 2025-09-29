package orchestrator

//go:generate mockgen -destination=./mocks/mock_orchestrator.go -package=mocks github.com/rizesky/mckmt/internal/orchestrator MetricsProvider

// MetricsProvider defines the interface for metrics operations.
//
// This interface is defined here (in the orchestrator package) because:
//  1. The orchestrator is the primary consumer of metrics operations
//  2. It follows the "define where used" principle - interfaces should be defined
//     where they are consumed, not where they are implemented
//  3. The orchestrator only needs a subset of all possible metrics operations,
//     so defining it here allows for interface segregation
//  4. This makes testing easier as we can mock exactly what the orchestrator needs
//
// Note: While metrics could be considered a "shared" concern, the orchestrator
// is currently the only component that directly uses metrics. If other components
// need metrics in the future, we can either:
// - Move this interface to a shared location (e.g., internal/interfaces/)
// - Create separate interfaces for each consumer's specific needs
type MetricsProvider interface {
	// Generic metrics methods
	IncrementCounter(name string, tags map[string]string)
	IncrementGauge(name string, value float64, tags map[string]string)
	RecordHistogram(name string, value float64, tags map[string]string)

	// HTTP metrics
	RecordHTTPRequest(method, endpoint, statusCode string, duration float64)
	IncHTTPRequestsInFlight(method, endpoint string)
	DecHTTPRequestsInFlight(method, endpoint string)

	// Cluster metrics
	SetClustersTotal(mode, status string, count float64)
	SetClusterStatus(clusterID, clusterName, mode, status string)
	SetClusterLastSeen(clusterID, clusterName string, timestamp float64)

	// Operation metrics
	RecordOperation(clusterID, operationType, status string, duration float64)
	IncOperationsInProgress(clusterID, operationType string)
	DecOperationsInProgress(clusterID, operationType string)

	// Agent metrics
	SetAgentsConnected(clusterID, agentVersion string, count float64)
	RecordAgentHeartbeat(clusterID, status string)
	SetAgentLastHeartbeat(clusterID string, timestamp float64)

	// Database metrics
	SetDatabaseConnections(state string, count float64)
	RecordDatabaseQuery(operation, table string, duration float64)

	// Cache metrics
	RecordCacheHit(cacheType, keyPattern string)
	RecordCacheMiss(cacheType, keyPattern string)
	RecordCacheOperation(operation, cacheType string)
}

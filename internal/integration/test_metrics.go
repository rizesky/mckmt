package integration

import (
	"github.com/rizesky/mckmt/internal/orchestrator"
)

// TestMetrics is a mock implementation of MetricsProvider for integration tests
type TestMetrics struct{}

// NewTestMetrics creates a new test metrics instance
func NewTestMetrics() *TestMetrics {
	return &TestMetrics{}
}

// Generic metrics methods
func (m *TestMetrics) IncrementCounter(name string, tags map[string]string)               {}
func (m *TestMetrics) IncrementGauge(name string, value float64, tags map[string]string)  {}
func (m *TestMetrics) RecordHistogram(name string, value float64, tags map[string]string) {}

// HTTP metrics
func (m *TestMetrics) RecordHTTPRequest(method, endpoint, statusCode string, duration float64) {}
func (m *TestMetrics) IncHTTPRequestsInFlight(method, endpoint string)                         {}
func (m *TestMetrics) DecHTTPRequestsInFlight(method, endpoint string)                         {}

// Cluster metrics
func (m *TestMetrics) SetClustersTotal(mode, status string, count float64)                 {}
func (m *TestMetrics) SetClusterStatus(clusterID, clusterName, mode, status string)        {}
func (m *TestMetrics) SetClusterLastSeen(clusterID, clusterName string, timestamp float64) {}

// Operation metrics
func (m *TestMetrics) RecordOperation(clusterID, operationType, status string, duration float64) {}
func (m *TestMetrics) IncOperationsInProgress(clusterID, operationType string)                   {}
func (m *TestMetrics) DecOperationsInProgress(clusterID, operationType string)                   {}

// Agent metrics
func (m *TestMetrics) SetAgentsConnected(clusterID, agentVersion string, count float64) {}
func (m *TestMetrics) RecordAgentHeartbeat(clusterID, status string)                    {}
func (m *TestMetrics) SetAgentLastHeartbeat(clusterID string, timestamp float64)        {}

// Database metrics
func (m *TestMetrics) SetDatabaseConnections(state string, count float64)            {}
func (m *TestMetrics) RecordDatabaseQuery(operation, table string, duration float64) {}

// Cache metrics
func (m *TestMetrics) RecordCacheHit(cacheType, keyPattern string)      {}
func (m *TestMetrics) RecordCacheMiss(cacheType, keyPattern string)     {}
func (m *TestMetrics) RecordCacheOperation(operation, cacheType string) {}

// Ensure TestMetrics implements MetricsProvider interface
var _ orchestrator.MetricsProvider = (*TestMetrics)(nil)

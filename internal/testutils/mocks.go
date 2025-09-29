package testutils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
	"go.uber.org/zap"
)

// MockOperationRepository is a mock implementation of repo.OperationRepository
type MockOperationRepository struct {
	operations map[uuid.UUID]*repo.Operation
	createErr  error
	getErr     error
	updateErr  error
	listErr    error
}

// NewMockOperationRepository creates a new mock operation repository
func NewMockOperationRepository() *MockOperationRepository {
	return &MockOperationRepository{
		operations: make(map[uuid.UUID]*repo.Operation),
	}
}

// SetCreateError sets the error to return on Create
func (m *MockOperationRepository) SetCreateError(err error) {
	m.createErr = err
}

// SetGetError sets the error to return on GetByID
func (m *MockOperationRepository) SetGetError(err error) {
	m.getErr = err
}

// SetUpdateError sets the error to return on Update
func (m *MockOperationRepository) SetUpdateError(err error) {
	m.updateErr = err
}

// SetListError sets the error to return on ListByCluster
func (m *MockOperationRepository) SetListError(err error) {
	m.listErr = err
}

// AddOperation adds an operation to the mock repository
func (m *MockOperationRepository) AddOperation(operation *repo.Operation) {
	m.operations[operation.ID] = operation
}

// GetOperation returns an operation from the mock repository
func (m *MockOperationRepository) GetOperation(id uuid.UUID) *repo.Operation {
	return m.operations[id]
}

// Create implements repo.OperationRepository
func (m *MockOperationRepository) Create(ctx context.Context, operation *repo.Operation) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.operations[operation.ID] = operation
	return nil
}

// GetByID implements repo.OperationRepository
func (m *MockOperationRepository) GetByID(ctx context.Context, id uuid.UUID) (*repo.Operation, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	operation, exists := m.operations[id]
	if !exists {
		return nil, repo.ErrNotFound
	}
	return operation, nil
}

// ListByCluster implements repo.OperationRepository
func (m *MockOperationRepository) ListByCluster(ctx context.Context, clusterID uuid.UUID, limit, offset int) ([]*repo.Operation, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	var operations []*repo.Operation
	count := 0
	for _, op := range m.operations {
		if op.ClusterID == clusterID {
			if count >= offset && len(operations) < limit {
				operations = append(operations, op)
			}
			count++
		}
	}
	return operations, nil
}

// Update implements repo.OperationRepository
func (m *MockOperationRepository) Update(ctx context.Context, operation *repo.Operation) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.operations[operation.ID] = operation
	return nil
}

// UpdateStatus implements repo.OperationRepository
func (m *MockOperationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	operation, exists := m.operations[id]
	if !exists {
		return repo.ErrNotFound
	}
	operation.Status = status
	operation.UpdatedAt = time.Now()
	return nil
}

// UpdateResult implements repo.OperationRepository
func (m *MockOperationRepository) UpdateResult(ctx context.Context, id uuid.UUID, result repo.Payload) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	operation, exists := m.operations[id]
	if !exists {
		return repo.ErrNotFound
	}
	operation.Result = &result
	operation.UpdatedAt = time.Now()
	return nil
}

// SetStarted implements repo.OperationRepository
func (m *MockOperationRepository) SetStarted(ctx context.Context, id uuid.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	operation, exists := m.operations[id]
	if !exists {
		return repo.ErrNotFound
	}
	operation.Status = "running"
	now := time.Now()
	operation.StartedAt = &now
	operation.UpdatedAt = now
	return nil
}

// SetFinished implements repo.OperationRepository
func (m *MockOperationRepository) SetFinished(ctx context.Context, id uuid.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	operation, exists := m.operations[id]
	if !exists {
		return repo.ErrNotFound
	}
	now := time.Now()
	operation.FinishedAt = &now
	operation.UpdatedAt = now
	return nil
}

// CancelOperation implements repo.OperationRepository
func (m *MockOperationRepository) CancelOperation(ctx context.Context, id uuid.UUID, reason string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	operation, exists := m.operations[id]
	if !exists {
		return repo.ErrNotFound
	}
	operation.Status = "cancelled"
	operation.Result = &repo.Payload{
		"cancelled":    true,
		"reason":       reason,
		"cancelled_at": time.Now(),
	}
	operation.UpdatedAt = time.Now()
	return nil
}

// MockClusterRepository is a mock implementation of repo.ClusterRepository
type MockClusterRepository struct {
	clusters  map[uuid.UUID]*repo.Cluster
	createErr error
	getErr    error
	updateErr error
	listErr   error
}

// NewMockClusterRepository creates a new mock cluster repository
func NewMockClusterRepository() *MockClusterRepository {
	return &MockClusterRepository{
		clusters: make(map[uuid.UUID]*repo.Cluster),
	}
}

// SetCreateError sets the error to return on Create
func (m *MockClusterRepository) SetCreateError(err error) {
	m.createErr = err
}

// SetGetError sets the error to return on GetByID
func (m *MockClusterRepository) SetGetError(err error) {
	m.getErr = err
}

// SetUpdateError sets the error to return on Update
func (m *MockClusterRepository) SetUpdateError(err error) {
	m.updateErr = err
}

// SetListError sets the error to return on List
func (m *MockClusterRepository) SetListError(err error) {
	m.listErr = err
}

// AddCluster adds a cluster to the mock repository
func (m *MockClusterRepository) AddCluster(cluster *repo.Cluster) {
	m.clusters[cluster.ID] = cluster
}

// GetCluster returns a cluster from the mock repository
func (m *MockClusterRepository) GetCluster(id uuid.UUID) *repo.Cluster {
	return m.clusters[id]
}

// Create implements repo.ClusterRepository
func (m *MockClusterRepository) Create(ctx context.Context, cluster *repo.Cluster) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.clusters[cluster.ID] = cluster
	return nil
}

// GetByID implements repo.ClusterRepository
func (m *MockClusterRepository) GetByID(ctx context.Context, id uuid.UUID) (*repo.Cluster, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	cluster, exists := m.clusters[id]
	if !exists {
		return nil, repo.ErrNotFound
	}
	return cluster, nil
}

// GetByName implements repo.ClusterRepository
func (m *MockClusterRepository) GetByName(ctx context.Context, name string) (*repo.Cluster, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, cluster := range m.clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}
	return nil, repo.ErrNotFound
}

// List implements repo.ClusterRepository
func (m *MockClusterRepository) List(ctx context.Context, limit, offset int) ([]*repo.Cluster, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	var clusters []*repo.Cluster
	count := 0
	for _, cluster := range m.clusters {
		if count >= offset && len(clusters) < limit {
			clusters = append(clusters, cluster)
		}
		count++
	}
	return clusters, nil
}

// Update implements repo.ClusterRepository
func (m *MockClusterRepository) Update(ctx context.Context, cluster *repo.Cluster) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.clusters[cluster.ID] = cluster
	return nil
}

// Delete implements repo.ClusterRepository
func (m *MockClusterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	delete(m.clusters, id)
	return nil
}

// UpdateLastSeen implements repo.ClusterRepository
func (m *MockClusterRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cluster, exists := m.clusters[id]
	if !exists {
		return repo.ErrNotFound
	}
	now := time.Now()
	cluster.LastSeenAt = &now
	cluster.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus implements repo.ClusterRepository
func (m *MockClusterRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cluster, exists := m.clusters[id]
	if !exists {
		return repo.ErrNotFound
	}
	cluster.Status = status
	cluster.UpdatedAt = time.Now()
	return nil
}

// MockCache is a mock implementation of repo.Cache
type MockCache struct {
	data      map[string]interface{}
	getErr    error
	setErr    error
	deleteErr error
}

// NewMockCache creates a new mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
	}
}

// SetGetError sets the error to return on Get
func (m *MockCache) SetGetError(err error) {
	m.getErr = err
}

// SetSetError sets the error to return on Set
func (m *MockCache) SetSetError(err error) {
	m.setErr = err
}

// SetDeleteError sets the error to return on Delete
func (m *MockCache) SetDeleteError(err error) {
	m.deleteErr = err
}

// Get implements repo.Cache
func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	if m.getErr != nil {
		return m.getErr
	}
	value, exists := m.data[key]
	if !exists {
		return repo.ErrCacheMiss
	}

	// Handle different destination types
	switch d := dest.(type) {
	case *interface{}:
		*d = value
	case *repo.Cluster:
		if cluster, ok := value.(*repo.Cluster); ok {
			*d = *cluster
		} else {
			return fmt.Errorf("type mismatch: expected *repo.Cluster, got %T", value)
		}
	default:
		// For other types, try to assign directly
		// This is a simplified implementation for testing
		if destPtr, ok := dest.(*interface{}); ok {
			*destPtr = value
		}
	}
	return nil
}

// Set implements repo.Cache
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}

// Delete implements repo.Cache
func (m *MockCache) Delete(ctx context.Context, key string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.data, key)
	return nil
}

// ClusterKey implements repo.Cache
func (m *MockCache) ClusterKey(id string) string {
	return "cluster:" + id
}

// OperationKey implements repo.Cache
func (m *MockCache) OperationKey(id string) string {
	return "operation:" + id
}

// UserKey implements repo.Cache
func (m *MockCache) UserKey(id string) string {
	return "user:" + id
}

// SessionKey implements repo.Cache
func (m *MockCache) SessionKey(token string) string {
	return "session:" + token
}

// ClusterResourcesKey implements repo.Cache
func (m *MockCache) ClusterResourcesKey(clusterID, kind, namespace string) string {
	return fmt.Sprintf("cluster:%s:resources:%s:%s", clusterID, kind, namespace)
}

// ClusterResourceKey implements repo.Cache
func (m *MockCache) ClusterResourceKey(clusterID, kind, namespace, name string) string {
	return fmt.Sprintf("cluster:%s:resource:%s:%s:%s", clusterID, kind, namespace, name)
}

// ClusterStatusKey implements repo.Cache
func (m *MockCache) ClusterStatusKey(clusterID string) string {
	return "cluster:" + clusterID + ":status"
}

// ClusterMetricsKey implements repo.Cache
func (m *MockCache) ClusterMetricsKey(clusterID string) string {
	return "cluster:" + clusterID + ":metrics"
}

// Keys implements repo.Cache
func (m *MockCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}

	var keys []string
	for key := range m.data {
		// Simple pattern matching for testing
		if pattern == "*" || strings.Contains(key, pattern) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

// FlushDB implements repo.Cache
func (m *MockCache) FlushDB(ctx context.Context) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.data = make(map[string]interface{})
	return nil
}

// Ping implements repo.Cache
func (m *MockCache) Ping(ctx context.Context) error {
	return nil
}

// Health implements repo.Cache
func (m *MockCache) Health(ctx context.Context) error {
	return nil
}

// MockOrchestrator is a mock implementation of OrchestratorInterface
type MockOrchestrator struct {
	queueErr  error
	queuedOps []*repo.Operation
}

// NewMockOrchestrator creates a new mock orchestrator
func NewMockOrchestrator() *MockOrchestrator {
	return &MockOrchestrator{
		queuedOps: make([]*repo.Operation, 0),
	}
}

// SetQueueError sets the error to return on QueueOperation
func (m *MockOrchestrator) SetQueueError(err error) {
	m.queueErr = err
}

// GetQueuedOperations returns the queued operations
func (m *MockOrchestrator) GetQueuedOperations() []*repo.Operation {
	return m.queuedOps
}

// QueueOperation implements OrchestratorInterface
func (m *MockOrchestrator) QueueOperation(operation *repo.Operation) error {
	if m.queueErr != nil {
		return m.queueErr
	}
	m.queuedOps = append(m.queuedOps, operation)
	return nil
}

// NewTestLogger creates a test logger
func NewTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

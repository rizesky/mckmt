package clusters

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
	"go.uber.org/zap"
)

// ClusterService handles cluster business logic
type ClusterService struct {
	clusterRepo   repo.ClusterRepository
	operationRepo repo.OperationRepository
	cache         repo.Cache
	logger        *zap.Logger
	orchestrator  OrchestratorInterface
}

//go:generate mockgen -destination=./mocks/mock_clusters.go -package=mocks github.com/rizesky/mckmt/internal/app/clusters OrchestratorInterface

// OrchestratorInterface defines the interface for orchestrator operations.
//
// This interface is defined here (in the clusters service package) because:
//  1. The clusters service is the primary consumer of orchestrator operations
//  2. It follows the "define where used" principle - interfaces should be defined
//     where they are consumed, not where they are implemented
//  3. The clusters service only needs a specific subset of orchestrator operations,
//     so defining it here allows for interface segregation
//  4. This makes testing easier as we can mock exactly what the clusters service needs
//  5. It decouples the clusters service from the concrete orchestrator implementation
//
// Note: This interface represents the clusters service's view of orchestrator operations.
// The actual implementation is provided by the orchestrator, but the clusters service
// defines what it needs, not what the orchestrator provides.
type OrchestratorInterface interface {
	QueueOperation(operation *repo.Operation) error
}

// NewClusterService creates a new cluster service
func NewClusterService(clusterRepo repo.ClusterRepository, operationRepo repo.OperationRepository, cache repo.Cache, logger *zap.Logger, orchestrator OrchestratorInterface) *ClusterService {
	return &ClusterService{
		clusterRepo:   clusterRepo,
		operationRepo: operationRepo,
		cache:         cache,
		logger:        logger,
		orchestrator:  orchestrator,
	}
}

// GetCluster retrieves a cluster by ID with caching
func (s *ClusterService) GetCluster(ctx context.Context, id uuid.UUID) (*repo.Cluster, error) {
	// Try cache first
	key := s.cache.ClusterKey(id.String())
	var cluster repo.Cluster
	err := s.cache.Get(ctx, key, &cluster)
	if err == nil {
		s.logger.Debug("Cluster cache hit", zap.String("cluster_id", id.String()))
		return &cluster, nil
	}

	if err != repo.ErrCacheMiss {
		s.logger.Warn("Cache error, falling back to database", zap.Error(err))
	}

	// Cache miss, get from database
	clusterFromDB, err := s.clusterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.Set(ctx, key, clusterFromDB, 1*time.Hour); err != nil {
		s.logger.Warn("Failed to cache cluster", zap.Error(err))
	}

	return clusterFromDB, nil
}

// ListClusters retrieves clusters with pagination
func (s *ClusterService) ListClusters(ctx context.Context, limit, offset int) ([]*repo.Cluster, error) {
	// For list operations, we don't cache since they're complex to invalidate
	return s.clusterRepo.List(ctx, limit, offset)
}

// RegisterCluster registers an existing cluster for management
func (s *ClusterService) RegisterCluster(ctx context.Context, cluster *repo.Cluster) error {
	err := s.clusterRepo.Create(ctx, cluster)
	if err != nil {
		return err
	}

	// Cache the registered cluster
	key := s.cache.ClusterKey(cluster.ID.String())
	if err := s.cache.Set(ctx, key, cluster, 1*time.Hour); err != nil {
		s.logger.Warn("Failed to cache registered cluster", zap.Error(err))
	}

	return nil
}

// UpdateCluster updates an existing cluster
func (s *ClusterService) UpdateCluster(ctx context.Context, id uuid.UUID, name, description string, labels map[string]string) error {
	// Get existing cluster
	cluster, err := s.clusterRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update fields
	cluster.Name = name
	cluster.Description = description
	cluster.Labels = labels

	err = s.clusterRepo.Update(ctx, cluster)
	if err != nil {
		return err
	}

	// Update cache
	key := s.cache.ClusterKey(cluster.ID.String())
	if err := s.cache.Set(ctx, key, cluster, 1*time.Hour); err != nil {
		s.logger.Warn("Failed to update cluster cache", zap.Error(err))
	}

	return nil
}

// DeleteCluster deletes a cluster
func (s *ClusterService) DeleteCluster(ctx context.Context, id uuid.UUID) error {
	err := s.clusterRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Remove from cache
	key := s.cache.ClusterKey(id.String())
	if err := s.cache.Delete(ctx, key); err != nil {
		s.logger.Warn("Failed to remove cluster from cache", zap.Error(err))
	}

	return nil
}

// GetClusterResources retrieves cluster resources with caching
func (s *ClusterService) GetClusterResources(ctx context.Context, clusterID uuid.UUID, kind, namespace string) ([]map[string]interface{}, error) {
	// Check cache first
	cacheKey := s.cache.ClusterResourcesKey(clusterID.String(), kind, namespace)
	var resources []map[string]interface{}
	err := s.cache.Get(ctx, cacheKey, &resources)
	if err == nil {
		s.logger.Debug("Cluster resources cache hit",
			zap.String("cluster_id", clusterID.String()),
			zap.String("kind", kind),
			zap.String("namespace", namespace))
		return resources, nil
	}

	if err != repo.ErrCacheMiss {
		s.logger.Warn("Cache error, falling back to database", zap.Error(err))
	}

	// Cache miss - in a real implementation, this would query the cluster
	// For now, return mock data
	resources = []map[string]interface{}{
		{
			"kind":       "Pod",
			"name":       "example-pod",
			"namespace":  "default",
			"status":     "Running",
			"created_at": time.Now().Add(-1 * time.Hour),
		},
		{
			"kind":       "Service",
			"name":       "example-service",
			"namespace":  "default",
			"type":       "ClusterIP",
			"created_at": time.Now().Add(-2 * time.Hour),
		},
	}

	// Cache the result for 5 minutes
	if err := s.cache.Set(ctx, cacheKey, resources, 5*time.Minute); err != nil {
		s.logger.Warn("Failed to cache cluster resources", zap.Error(err))
	}

	return resources, nil
}

// CreateOperation creates a new operation
func (s *ClusterService) CreateOperation(ctx context.Context, operation *repo.Operation) error {
	return s.operationRepo.Create(ctx, operation)
}

// QueueOperation queues an operation for processing
func (s *ClusterService) QueueOperation(ctx context.Context, operation *repo.Operation) error {
	return s.orchestrator.QueueOperation(operation)
}

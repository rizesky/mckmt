package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/metrics"
	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/repo/redis"
	"go.uber.org/zap"
)

// cachedClusterRepository wraps repo.ClusterRepository with caching
type cachedClusterRepository struct {
	repo    repo.ClusterRepository
	cache   *redis.CacheAdapter
	metrics *metrics.Metrics
	logger  *zap.Logger
}

// NewCachedrepo.ClusterRepository creates a new cached cluster repository
func NewCachedClusterRepository(repo repo.ClusterRepository, cache *redis.CacheAdapter, metrics *metrics.Metrics, logger *zap.Logger) repo.ClusterRepository {
	return &cachedClusterRepository{
		repo:    repo,
		cache:   cache,
		metrics: metrics,
		logger:  logger,
	}
}

func (r *cachedClusterRepository) Create(ctx context.Context, cluster *repo.Cluster) error {
	err := r.repo.Create(ctx, cluster)
	if err != nil {
		return err
	}

	// Cache the created cluster
	key := r.cache.ClusterKey(cluster.ID.String())
	if err := r.cache.Set(ctx, key, cluster, 1*time.Hour); err != nil {
		r.logger.Warn("Failed to cache created cluster", zap.Error(err))
	}

	return nil
}

func (r *cachedClusterRepository) GetByID(ctx context.Context, id uuid.UUID) (*repo.Cluster, error) {
	// Try cache first
	key := r.cache.ClusterKey(id.String())
	var cluster repo.Cluster
	err := r.cache.Get(ctx, key, &cluster)
	if err == nil {
		r.logger.Debug("Cluster cache hit", zap.String("cluster_id", id.String()))
		r.metrics.RecordCacheHit("cluster", "single")
		return &cluster, nil
	}

	if err != repo.ErrCacheMiss {
		r.logger.Warn("Cache error, falling back to database", zap.Error(err))
		r.metrics.RecordCacheOperation("get", "cluster")
	} else {
		r.metrics.RecordCacheMiss("cluster", "single")
	}

	// Cache miss, get from database
	clusterFromDB, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, key, clusterFromDB, 1*time.Hour); err != nil {
		r.logger.Warn("Failed to cache cluster", zap.Error(err))
	}
	r.metrics.RecordCacheOperation("set", "cluster")

	return clusterFromDB, nil
}

func (r *cachedClusterRepository) GetByName(ctx context.Context, name string) (*repo.Cluster, error) {
	// For name lookups, we don't cache since it's less common
	// and would require additional cache management
	return r.repo.GetByName(ctx, name)
}

func (r *cachedClusterRepository) List(ctx context.Context, limit, offset int) ([]*repo.Cluster, error) {
	// For list operations, we don't cache since they're complex to invalidate
	// and the data changes frequently
	return r.repo.List(ctx, limit, offset)
}

func (r *cachedClusterRepository) Update(ctx context.Context, cluster *repo.Cluster) error {
	err := r.repo.Update(ctx, cluster)
	if err != nil {
		return err
	}

	// Update cache
	key := r.cache.ClusterKey(cluster.ID.String())
	if err := r.cache.Set(ctx, key, cluster, 1*time.Hour); err != nil {
		r.logger.Warn("Failed to update cluster cache", zap.Error(err))
	}

	return nil
}

func (r *cachedClusterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Remove from cache
	key := r.cache.ClusterKey(id.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn("Failed to remove cluster from cache", zap.Error(err))
	}

	return nil
}

func (r *cachedClusterRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	err := r.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	// Invalidate cache to force refresh
	key := r.cache.ClusterKey(id.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn("Failed to invalidate cluster cache", zap.Error(err))
	}

	return nil
}

func (r *cachedClusterRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	err := r.repo.UpdateLastSeen(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate cache to force refresh
	key := r.cache.ClusterKey(id.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn("Failed to invalidate cluster cache", zap.Error(err))
	}

	return nil
}

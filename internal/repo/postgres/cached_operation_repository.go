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

// cachedOperationRepository wraps repo.OperationRepository with caching
type cachedOperationRepository struct {
	repo    repo.OperationRepository
	cache   *redis.CacheAdapter
	metrics *metrics.Metrics
	logger  *zap.Logger
}

// NewCachedOperationRepository creates a new cached operation repository
func NewCachedOperationRepository(repo repo.OperationRepository, cache *redis.CacheAdapter, metrics *metrics.Metrics, logger *zap.Logger) repo.OperationRepository {
	return &cachedOperationRepository{
		repo:    repo,
		cache:   cache,
		metrics: metrics,
		logger:  logger,
	}
}

func (r *cachedOperationRepository) Create(ctx context.Context, operation *repo.Operation) error {
	err := r.repo.Create(ctx, operation)
	if err != nil {
		return err
	}

	// Cache the created operation
	key := r.cache.OperationKey(operation.ID.String())
	if err := r.cache.Set(ctx, key, operation, 30*time.Minute); err != nil {
		r.logger.Warn("Failed to cache created operation", zap.Error(err))
	}

	return nil
}

func (r *cachedOperationRepository) GetByID(ctx context.Context, id uuid.UUID) (*repo.Operation, error) {
	// Try cache first
	key := r.cache.OperationKey(id.String())
	var operation repo.Operation
	if err := r.cache.Get(ctx, key, &operation); err == nil {
		return &operation, nil
	}

	// Cache miss, get from database
	operationPtr, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, key, *operationPtr, 30*time.Minute); err != nil {
		r.logger.Warn("Failed to cache operation", zap.Error(err))
	}

	return operationPtr, nil
}

func (r *cachedOperationRepository) ListByCluster(ctx context.Context, clusterID uuid.UUID, limit, offset int) ([]*repo.Operation, error) {
	// For cluster-specific operations, we don't cache as they can be large and change frequently
	// Just delegate to the underlying repository
	return r.repo.ListByCluster(ctx, clusterID, limit, offset)
}

func (r *cachedOperationRepository) Update(ctx context.Context, operation *repo.Operation) error {
	err := r.repo.Update(ctx, operation)
	if err != nil {
		return err
	}

	// Invalidate cache for this operation
	key := r.cache.OperationKey(operation.ID.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn("Failed to invalidate operation cache", zap.Error(err))
	}

	return nil
}

func (r *cachedOperationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	err := r.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	// Invalidate cache for this operation
	key := r.cache.OperationKey(id.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn("Failed to invalidate operation cache", zap.Error(err))
	}

	return nil
}

func (r *cachedOperationRepository) UpdateResult(ctx context.Context, id uuid.UUID, result repo.Payload) error {
	err := r.repo.UpdateResult(ctx, id, result)
	if err != nil {
		return err
	}

	// Invalidate cache for this operation
	key := r.cache.OperationKey(id.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn("Failed to invalidate operation cache", zap.Error(err))
	}

	return nil
}

func (r *cachedOperationRepository) SetStarted(ctx context.Context, id uuid.UUID) error {
	err := r.repo.SetStarted(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate cache for this operation
	key := r.cache.OperationKey(id.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn("Failed to invalidate operation cache", zap.Error(err))
	}

	return nil
}

func (r *cachedOperationRepository) SetFinished(ctx context.Context, id uuid.UUID) error {
	err := r.repo.SetFinished(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate cache for this operation
	key := r.cache.OperationKey(id.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn("Failed to invalidate operation cache", zap.Error(err))
	}

	return nil
}

func (r *cachedOperationRepository) CancelOperation(ctx context.Context, id uuid.UUID, reason string) error {
	err := r.repo.CancelOperation(ctx, id, reason)
	if err != nil {
		return err
	}

	// Invalidate cache for this operation
	key := r.cache.OperationKey(id.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn("Failed to invalidate operation cache", zap.Error(err))
	}

	return nil
}

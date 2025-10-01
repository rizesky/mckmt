package operation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
	"go.uber.org/zap"
)

// Service handles operation business logic
type Service struct {
	operationRepo repo.OperationRepository
	cache         repo.Cache
	logger        *zap.Logger
	orchestrator  OrchestratorInterface
}

// OrchestratorInterface defines the interface for orchestrator operations
type OrchestratorInterface interface {
	CancelOperation(operationID uuid.UUID) error
}

// NewService creates a new operation service
func NewService(operationRepo repo.OperationRepository, cache repo.Cache, logger *zap.Logger, orchestrator OrchestratorInterface) *Service {
	return &Service{
		operationRepo: operationRepo,
		cache:         cache,
		logger:        logger,
		orchestrator:  orchestrator,
	}
}

// GetOperation retrieves an operation by ID with caching
func (s *Service) GetOperation(ctx context.Context, id uuid.UUID) (*repo.Operation, error) {
	// Try cache first
	key := s.cache.OperationKey(id.String())
	var operation repo.Operation
	err := s.cache.Get(ctx, key, &operation)
	if err == nil {
		s.logger.Debug("Operation cache hit", zap.String("operation_id", id.String()))
		return &operation, nil
	}

	if err != repo.ErrCacheMiss {
		s.logger.Warn("Cache error, falling back to database", zap.Error(err))
	}

	// Cache miss, get from database
	operationFromDB, err := s.operationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.Set(ctx, key, operationFromDB, 30*time.Minute); err != nil {
		s.logger.Warn("Failed to cache operation", zap.Error(err))
	}

	return operationFromDB, nil
}

// ListOperationsByCluster retrieves operations for a cluster
func (s *Service) ListOperationsByCluster(ctx context.Context, clusterID uuid.UUID, limit, offset int) ([]*repo.Operation, error) {
	// For list operations, we don't cache since they're complex to invalidate
	return s.operationRepo.ListByCluster(ctx, clusterID, limit, offset)
}

// CreateOperation creates a new operation
func (s *Service) CreateOperation(ctx context.Context, operation *repo.Operation) error {
	err := s.operationRepo.Create(ctx, operation)
	if err != nil {
		return err
	}

	// Cache the created operation
	key := s.cache.OperationKey(operation.ID.String())
	if err := s.cache.Set(ctx, key, operation, 30*time.Minute); err != nil {
		s.logger.Warn("Failed to cache created operation", zap.Error(err))
	}

	return nil
}

// UpdateOperation updates an existing operation
func (s *Service) UpdateOperation(ctx context.Context, operation *repo.Operation) error {
	err := s.operationRepo.Update(ctx, operation)
	if err != nil {
		return err
	}

	// Update cache
	key := s.cache.OperationKey(operation.ID.String())
	if err := s.cache.Set(ctx, key, operation, 30*time.Minute); err != nil {
		s.logger.Warn("Failed to update operation cache", zap.Error(err))
	}

	return nil
}

// UpdateOperationStatus updates operation status
func (s *Service) UpdateOperationStatus(ctx context.Context, id uuid.UUID, status string) error {
	err := s.operationRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	// Invalidate cache to force refresh
	key := s.cache.OperationKey(id.String())
	if err := s.cache.Delete(ctx, key); err != nil {
		s.logger.Warn("Failed to invalidate operation cache", zap.Error(err))
	}

	return nil
}

// UpdateOperationResult updates operation result
func (s *Service) UpdateOperationResult(ctx context.Context, id uuid.UUID, result repo.Payload) error {
	err := s.operationRepo.UpdateResult(ctx, id, result)
	if err != nil {
		return err
	}

	// Invalidate cache to force refresh
	key := s.cache.OperationKey(id.String())
	if err := s.cache.Delete(ctx, key); err != nil {
		s.logger.Warn("Failed to invalidate operation cache", zap.Error(err))
	}

	return nil
}

// CancelOperation cancels an operation
func (s *Service) CancelOperation(ctx context.Context, id uuid.UUID, reason string) error {
	// First, get the operation to check if it can be cancelled
	operation, err := s.GetOperation(ctx, id)
	if err != nil {
		return err
	}

	// Check if operation can be cancelled
	if operation.Status == string(repo.OperationStatusSuccess) ||
		operation.Status == string(repo.OperationStatusFailed) ||
		operation.Status == string(repo.OperationStatusCancelled) {
		return fmt.Errorf("operation cannot be cancelled, current status: %s", operation.Status)
	}

	// Use orchestrator to cancel the operation
	err = s.orchestrator.CancelOperation(id)
	if err != nil {
		return err
	}

	// Invalidate cache to force refresh
	key := s.cache.OperationKey(id.String())
	if err := s.cache.Delete(ctx, key); err != nil {
		s.logger.Warn("Failed to invalidate operation cache", zap.Error(err))
	}

	s.logger.Info("Operation cancellation requested",
		zap.String("operation_id", id.String()),
		zap.String("reason", reason))

	return nil
}

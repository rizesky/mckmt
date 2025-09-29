package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/repo"
)

// Orchestrator manages long-running operations for agent-based clusters
type Orchestrator struct {
	operations repo.OperationRepository
	metrics    MetricsProvider
	logger     *zap.Logger
	workers    int
	queue      chan *repo.Operation
	stopCh     chan struct{}
	cancelCh   chan uuid.UUID
	runningOps map[uuid.UUID]context.CancelFunc
}

// NewOrchestrator creates a new orchestrator for agent-based operations
func NewOrchestrator(operations repo.OperationRepository, metrics MetricsProvider, logger *zap.Logger, workers int) *Orchestrator {
	// Ensure at least 1 worker
	if workers <= 0 {
		workers = 1
	}

	return &Orchestrator{
		operations: operations,
		metrics:    metrics,
		logger:     logger,
		workers:    workers,
		queue:      make(chan *repo.Operation, 1000),
		stopCh:     make(chan struct{}),
		cancelCh:   make(chan uuid.UUID, 100),
		runningOps: make(map[uuid.UUID]context.CancelFunc),
	}
}

// Start starts the orchestrator
func (o *Orchestrator) Start(ctx context.Context) error {
	o.logger.Info("Starting orchestrator", zap.Int("workers", o.workers))

	// Start worker goroutines
	for i := 0; i < o.workers; i++ {
		go o.worker(ctx, i)
	}

	// Start operation processor
	go o.processOperations(ctx)

	o.logger.Info("Orchestrator started successfully")
	return nil
}

// Stop stops the orchestrator
func (o *Orchestrator) Stop() {
	o.logger.Info("Stopping orchestrator")
	close(o.stopCh)
	close(o.queue)
}

// QueueOperation queues an operation for processing
func (o *Orchestrator) QueueOperation(operation *repo.Operation) error {
	select {
	case o.queue <- operation:
		o.logger.Info("Operation queued",
			zap.String("operation_id", operation.ID.String()),
			zap.String("type", operation.Type),
			zap.String("cluster_id", operation.ClusterID.String()),
		)
		return nil
	default:
		return fmt.Errorf("operation queue is full")
	}
}

// CancelOperation cancels a running operation
func (o *Orchestrator) CancelOperation(operationID uuid.UUID) error {
	select {
	case o.cancelCh <- operationID:
		o.logger.Info("Operation cancellation requested",
			zap.String("operation_id", operationID.String()),
		)
		return nil
	default:
		return fmt.Errorf("cancellation queue is full")
	}
}

// processOperations processes operations from the queue
func (o *Orchestrator) processOperations(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopCh:
			return
		case operationID := <-o.cancelCh:
			o.handleCancellation(operationID)
		case operation := <-o.queue:
			if operation == nil {
				return
			}
			o.processOperation(ctx, operation)
		}
	}
}

// worker processes operations from the queue
func (o *Orchestrator) worker(ctx context.Context, workerID int) {
	o.logger.Info("Worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopCh:
			return
		case operation := <-o.queue:
			if operation == nil {
				return
			}
			o.processOperation(ctx, operation)
		}
	}
}

// processOperation processes a single operation
func (o *Orchestrator) processOperation(ctx context.Context, operation *repo.Operation) {
	startTime := time.Now()

	o.logger.Info("Processing operation",
		zap.String("operation_id", operation.ID.String()),
		zap.String("type", operation.Type),
		zap.String("cluster_id", operation.ClusterID.String()),
	)

	// Check if operation was already cancelled before processing
	existingOp, err := o.operations.GetByID(ctx, operation.ID)
	if err == nil && existingOp.Status == string(repo.OperationStatusCancelled) {
		o.logger.Info("Operation was already cancelled, skipping processing",
			zap.String("operation_id", operation.ID.String()),
		)
		return
	}

	// Create cancellable context for this operation
	opCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Store the cancel function for potential cancellation
	o.runningOps[operation.ID] = cancel
	defer func() {
		delete(o.runningOps, operation.ID)
	}()

	// Check if operation was cancelled before processing
	if opCtx.Err() == context.Canceled {
		o.logger.Info("Operation was cancelled before processing",
			zap.String("operation_id", operation.ID.String()),
		)
		return
	}

	// Mark operation as started
	if err := o.operations.SetStarted(opCtx, operation.ID); err != nil {
		o.logger.Error("Failed to mark operation as started", zap.Error(err))
		return
	}

	// Increment metrics
	o.metrics.IncOperationsInProgress(operation.ClusterID.String(), operation.Type)

	// Process based on operation type
	var result repo.Payload
	var success bool
	var message string

	switch operation.Type {
	case string(repo.OperationTypeApply):
		result, success, message = o.processApplyOperation(opCtx, operation)
	case string(repo.OperationTypeExec):
		result, success, message = o.processExecOperation(opCtx, operation)
	case string(repo.OperationTypeSync):
		result, success, message = o.processSyncOperation(opCtx, operation)
	case string(repo.OperationTypeDelete):
		result, success, message = o.processDeleteOperation(opCtx, operation)
	default:
		success = false
		message = fmt.Sprintf("unknown operation type: %s", operation.Type)
	}

	// Check if operation was cancelled during processing
	if opCtx.Err() == context.Canceled {
		success = false
		message = "Operation was cancelled"
		result = repo.Payload{
			"status":  "cancelled",
			"message": "Operation was cancelled",
		}
	}

	// Calculate duration
	duration := time.Since(startTime).Seconds()

	// Update operation status
	status := string(repo.OperationStatusSuccess)
	if !success {
		if opCtx.Err() == context.Canceled {
			status = string(repo.OperationStatusCancelled)
		} else {
			status = string(repo.OperationStatusFailed)
		}
	}

	if err := o.operations.UpdateStatus(ctx, operation.ID, status); err != nil {
		o.logger.Error("Failed to update operation status", zap.Error(err))
	}

	if err := o.operations.UpdateResult(ctx, operation.ID, result); err != nil {
		o.logger.Error("Failed to update operation result", zap.Error(err))
	}

	if err := o.operations.SetFinished(ctx, operation.ID); err != nil {
		o.logger.Error("Failed to mark operation as finished", zap.Error(err))
	}

	// Decrement metrics
	o.metrics.DecOperationsInProgress(operation.ClusterID.String(), operation.Type)
	o.metrics.RecordOperation(operation.ClusterID.String(), operation.Type, status, duration)

	o.logger.Info("Operation completed",
		zap.String("operation_id", operation.ID.String()),
		zap.String("status", status),
		zap.Bool("success", success),
		zap.String("message", message),
		zap.Float64("duration", duration),
	)
}

// processApplyOperation queues an apply operation for agent processing
func (o *Orchestrator) processApplyOperation(_ context.Context, operation *repo.Operation) (repo.Payload, bool, string) {
	// For agent mode, we just queue the operation and let the agent handle it
	// The actual processing will be done by the agent via gRPC
	o.logger.Info("Queued apply operation for agent processing",
		zap.String("operation_id", operation.ID.String()),
		zap.String("cluster_id", operation.ClusterID.String()),
	)

	return repo.Payload{
		"status":  "queued",
		"message": "Operation queued for agent processing",
	}, true, "Operation queued for agent processing"
}

// processExecOperation queues an exec operation for agent processing
func (o *Orchestrator) processExecOperation(ctx context.Context, operation *repo.Operation) (repo.Payload, bool, string) {
	// For agent mode, we just queue the operation and let the agent handle it
	o.logger.Info("Queued exec operation for agent processing",
		zap.String("operation_id", operation.ID.String()),
		zap.String("cluster_id", operation.ClusterID.String()),
	)

	return repo.Payload{
		"status":  "queued",
		"message": "Operation queued for agent processing",
	}, true, "Operation queued for agent processing"
}

// processSyncOperation queues a sync operation for agent processing
func (o *Orchestrator) processSyncOperation(ctx context.Context, operation *repo.Operation) (repo.Payload, bool, string) {
	// For agent mode, we just queue the operation and let the agent handle it
	o.logger.Info("Queued sync operation for agent processing",
		zap.String("operation_id", operation.ID.String()),
		zap.String("cluster_id", operation.ClusterID.String()),
	)

	return repo.Payload{
		"status":  "queued",
		"message": "Operation queued for agent processing",
	}, true, "Operation queued for agent processing"
}

// processDeleteOperation queues a delete operation for agent processing
func (o *Orchestrator) processDeleteOperation(_ context.Context, operation *repo.Operation) (repo.Payload, bool, string) {
	// For agent mode, we just queue the operation and let the agent handle it
	o.logger.Info("Queued delete operation for agent processing",
		zap.String("operation_id", operation.ID.String()),
		zap.String("cluster_id", operation.ClusterID.String()),
	)

	return repo.Payload{
		"status":  "queued",
		"message": "Operation queued for agent processing",
	}, true, "Operation queued for agent processing"
}

// handleCancellation handles operation cancellation requests
func (o *Orchestrator) handleCancellation(operationID uuid.UUID) {
	o.logger.Info("Handling operation cancellation",
		zap.String("operation_id", operationID.String()),
	)

	// Check if operation is currently running
	if cancel, exists := o.runningOps[operationID]; exists {
		// Cancel the running operation
		cancel()
		o.logger.Info("Operation cancelled",
			zap.String("operation_id", operationID.String()),
		)
	} else {
		// Operation is not running, update status directly
		if err := o.operations.UpdateStatus(context.Background(), operationID, string(repo.OperationStatusCancelled)); err != nil {
			o.logger.Error("Failed to update operation status to cancelled", zap.Error(err))
		} else {
			o.logger.Info("Operation marked as cancelled (was not running)",
				zap.String("operation_id", operationID.String()),
			)
		}
	}
}

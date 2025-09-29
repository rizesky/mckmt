package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
)

// TestOrchestratorProcessingIntegration tests the orchestrator's operation processing
func TestOrchestratorProcessingIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()

	// Test 1: Queue and process operations
	t.Run("queue_and_process_operations", func(t *testing.T) {
		operationID := uuid.New()

		// Create operation
		operation := &repo.Operation{
			ID:        operationID,
			ClusterID: suite.TestClusterID,
			Type:      "apply",
			Status:    "queued",
			Payload: repo.Payload{
				"manifests": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test-pod",
				"namespace": "default",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Create operation in repository
		err := suite.OperationRepo.Create(ctx, operation)
		if err != nil {
			t.Fatalf("Failed to create operation: %v", err)
		}

		// Queue operation
		err = suite.Orchestrator.QueueOperation(operation)
		if err != nil {
			t.Fatalf("Failed to queue operation: %v", err)
		}

		// Wait for processing
		time.Sleep(200 * time.Millisecond)

		// Verify operation was processed (status should be updated by orchestrator)
		processedOp, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get processed operation: %v", err)
		}

		// The orchestrator should have updated the operation
		if processedOp.Status == "queued" {
			t.Error("Expected operation status to be updated by orchestrator")
		}
	})

	// Test 2: Process multiple operations
	t.Run("process_multiple_operations", func(t *testing.T) {
		numOperations := 3
		operationIDs := make([]uuid.UUID, numOperations)

		// Create multiple operations
		for i := 0; i < numOperations; i++ {
			operationID := uuid.New()
			operationIDs[i] = operationID

			operation := &repo.Operation{
				ID:        operationID,
				ClusterID: suite.TestClusterID,
				Type:      "apply",
				Status:    "queued",
				Payload: repo.Payload{
					"index":     i,
					"manifests": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test-pod-" + string(rune(i)),
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := suite.OperationRepo.Create(ctx, operation)
			if err != nil {
				t.Fatalf("Failed to create operation %d: %v", i, err)
			}

			err = suite.Orchestrator.QueueOperation(operation)
			if err != nil {
				t.Fatalf("Failed to queue operation %d: %v", i, err)
			}
		}

		// Wait for processing
		time.Sleep(500 * time.Millisecond)

		// Verify all operations were processed
		for i, operationID := range operationIDs {
			operation, err := suite.OperationRepo.GetByID(ctx, operationID)
			if err != nil {
				t.Fatalf("Failed to get operation %d: %v", i, err)
			}

			if operation.Status == "queued" {
				t.Errorf("Expected operation %d to be processed, but status is still 'queued'", i)
			}
		}
	})
}

// TestOrchestratorCancellationIntegration tests operation cancellation through orchestrator
func TestOrchestratorCancellationIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()
	operationID := uuid.New()

	// Create and queue operation
	operation := &repo.Operation{
		ID:        operationID,
		ClusterID: suite.TestClusterID,
		Type:      "apply",
		Status:    "queued",
		Payload: repo.Payload{
			"manifests": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test-pod",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := suite.OperationRepo.Create(ctx, operation)
	if err != nil {
		t.Fatalf("Failed to create operation: %v", err)
	}

	err = suite.Orchestrator.QueueOperation(operation)
	if err != nil {
		t.Fatalf("Failed to queue operation: %v", err)
	}

	// Test cancellation
	t.Run("cancel_queued_operation", func(t *testing.T) {
		// Cancel operation through orchestrator BEFORE starting the orchestrator
		err := suite.Orchestrator.CancelOperation(operationID)
		if err != nil {
			t.Fatalf("Failed to cancel operation: %v", err)
		}

		// Wait a bit for cancellation to be processed
		time.Sleep(100 * time.Millisecond)

		// Verify operation was cancelled
		operation, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get cancelled operation: %v", err)
		}

		if operation.Status != "cancelled" {
			t.Errorf("Expected operation status 'cancelled', got '%s'", operation.Status)
		}
	})
}

// TestOrchestratorWorkerPoolIntegration tests the orchestrator's worker pool
func TestOrchestratorWorkerPoolIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()

	// Test worker pool with multiple operations
	t.Run("worker_pool_processing", func(t *testing.T) {
		numOperations := 5
		operationIDs := make([]uuid.UUID, numOperations)

		// Create operations with different types
		operationTypes := []string{"apply", "exec", "sync", "delete", "apply"}

		for i := 0; i < numOperations; i++ {
			operationID := uuid.New()
			operationIDs[i] = operationID

			operation := &repo.Operation{
				ID:        operationID,
				ClusterID: suite.TestClusterID,
				Type:      operationTypes[i],
				Status:    "queued",
				Payload: repo.Payload{
					"index":     i,
					"type":      operationTypes[i],
					"manifests": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test-pod-" + string(rune(i)),
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := suite.OperationRepo.Create(ctx, operation)
			if err != nil {
				t.Fatalf("Failed to create operation %d: %v", i, err)
			}

			err = suite.Orchestrator.QueueOperation(operation)
			if err != nil {
				t.Fatalf("Failed to queue operation %d: %v", i, err)
			}
		}

		// Wait for all operations to be processed
		time.Sleep(1 * time.Second)

		// Verify all operations were processed
		processedCount := 0
		for i, operationID := range operationIDs {
			operation, err := suite.OperationRepo.GetByID(ctx, operationID)
			if err != nil {
				t.Fatalf("Failed to get operation %d: %v", i, err)
			}

			if operation.Status != "queued" {
				processedCount++
			}
		}

		if processedCount == 0 {
			t.Error("Expected at least some operations to be processed by worker pool")
		}
	})
}

// TestOrchestratorMetricsIntegration tests metrics collection during operation processing
func TestOrchestratorMetricsIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()

	// Test metrics collection
	t.Run("metrics_collection", func(t *testing.T) {
		operationID := uuid.New()

		// Create operation
		operation := &repo.Operation{
			ID:        operationID,
			ClusterID: suite.TestClusterID,
			Type:      "apply",
			Status:    "queued",
			Payload: repo.Payload{
				"manifests": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test-pod",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := suite.OperationRepo.Create(ctx, operation)
		if err != nil {
			t.Fatalf("Failed to create operation: %v", err)
		}

		// Queue operation (this should trigger metrics)
		err = suite.Orchestrator.QueueOperation(operation)
		if err != nil {
			t.Fatalf("Failed to queue operation: %v", err)
		}

		// Wait for processing
		time.Sleep(200 * time.Millisecond)

		// Note: In a real integration test, we would verify metrics were recorded
		// For now, we just ensure the operation was processed without errors
		processedOp, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get processed operation: %v", err)
		}

		if processedOp.Status == "queued" {
			t.Error("Expected operation to be processed")
		}
	})
}

// TestOrchestratorErrorHandlingIntegration tests error handling in orchestrator
func TestOrchestratorErrorHandlingIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()

	// Test error handling
	t.Run("error_handling", func(t *testing.T) {
		operationID := uuid.New()

		// Create operation with invalid payload to trigger error
		operation := &repo.Operation{
			ID:        operationID,
			ClusterID: suite.TestClusterID,
			Type:      "invalid_type", // This should trigger an error
			Status:    "queued",
			Payload: repo.Payload{
				"invalid": "data",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := suite.OperationRepo.Create(ctx, operation)
		if err != nil {
			t.Fatalf("Failed to create operation: %v", err)
		}

		// Queue operation
		err = suite.Orchestrator.QueueOperation(operation)
		if err != nil {
			t.Fatalf("Failed to queue operation: %v", err)
		}

		// Wait for processing
		time.Sleep(200 * time.Millisecond)

		// Verify operation was handled (even if it failed)
		processedOp, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get processed operation: %v", err)
		}

		// The operation should have been processed (status changed from queued)
		if processedOp.Status == "queued" {
			t.Error("Expected operation status to be updated even if processing failed")
		}
	})
}

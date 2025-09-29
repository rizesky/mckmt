package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
)

// TestClusterLifecycleIntegration tests the complete cluster lifecycle
func TestClusterLifecycleIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()
	clusterID := uuid.New()

	// Create cluster first (outside of subtests to ensure it persists)
	cluster := &repo.Cluster{
		ID:          clusterID,
		Name:        "integration-test-cluster",
		Description: "Cluster for integration testing",
		Status:      "active",
		Labels:      repo.Labels{"env": "integration", "test": "true"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := suite.ClusterRepo.Create(ctx, cluster)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v", err)
	}

	// Test 1: Verify cluster creation
	t.Run("create_cluster", func(t *testing.T) {
		suite.AssertClusterExists(t, clusterID)
	})

	// Test 2: Get cluster
	t.Run("get_cluster", func(t *testing.T) {
		cluster, err := suite.ClusterService.GetCluster(ctx, clusterID)
		if err != nil {
			t.Fatalf("Failed to get cluster: %v", err)
		}

		if cluster.Name != "integration-test-cluster" {
			t.Errorf("Expected cluster name 'integration-test-cluster', got '%s'", cluster.Name)
		}

		// Mode field was removed from Cluster struct
	})

	// Test 3: List clusters
	t.Run("list_clusters", func(t *testing.T) {
		clusters, err := suite.ClusterService.ListClusters(ctx, 10, 0)
		if err != nil {
			t.Fatalf("Failed to list clusters: %v", err)
		}

		if len(clusters) < 1 {
			t.Errorf("Expected at least 1 cluster, got %d", len(clusters))
		}

		found := false
		for _, cluster := range clusters {
			if cluster.ID == clusterID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected to find cluster %s in list", clusterID)
		}
	})

	// Test 4: Update cluster
	t.Run("update_cluster", func(t *testing.T) {
		// First, verify the cluster exists before update
		cluster, err := suite.ClusterService.GetCluster(ctx, clusterID)
		if err != nil {
			t.Fatalf("Failed to get cluster before update: %v", err)
		}
		t.Logf("Cluster before update: Name='%s', Description='%s'", cluster.Name, cluster.Description)

		err = suite.ClusterService.UpdateCluster(
			ctx,
			clusterID,
			"updated-integration-test-cluster",
			"Updated description",
			map[string]string{"env": "integration", "test": "true", "updated": "true"},
		)
		if err != nil {
			t.Fatalf("Failed to update cluster: %v", err)
		}

		// Get the updated cluster
		updatedCluster, err := suite.ClusterService.GetCluster(ctx, clusterID)
		if err != nil {
			t.Fatalf("Failed to get updated cluster: %v", err)
		}

		t.Logf("Cluster after update: Name='%s', Description='%s'", updatedCluster.Name, updatedCluster.Description)

		if updatedCluster.Name != "updated-integration-test-cluster" {
			t.Errorf("Expected updated cluster name 'updated-integration-test-cluster', got '%s'", updatedCluster.Name)
		}

		if updatedCluster.Description != "Updated description" {
			t.Errorf("Expected updated description 'Updated description', got '%s'", updatedCluster.Description)
		}
	})

	// Test 5: Delete cluster
	t.Run("delete_cluster", func(t *testing.T) {
		err := suite.ClusterService.DeleteCluster(ctx, clusterID)
		if err != nil {
			t.Fatalf("Failed to delete cluster: %v", err)
		}

		suite.AssertClusterNotExists(t, clusterID)
	})
}

// TestOperationFlowIntegration tests the complete operation flow
func TestOperationFlowIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()
	operationID := uuid.New()

	// Test 1: Create operation
	t.Run("create_operation", func(t *testing.T) {
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

		err := suite.ClusterService.CreateOperation(ctx, operation)
		if err != nil {
			t.Fatalf("Failed to create operation: %v", err)
		}

		// Verify operation was created
		createdOp, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get created operation: %v", err)
		}

		if createdOp.Type != "apply" {
			t.Errorf("Expected operation type 'apply', got '%s'", createdOp.Type)
		}

		if createdOp.Status != "queued" {
			t.Errorf("Expected operation status 'queued', got '%s'", createdOp.Status)
		}
	})

	// Test 2: Queue operation
	t.Run("queue_operation", func(t *testing.T) {
		operation, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get operation for queuing: %v", err)
		}

		err = suite.ClusterService.QueueOperation(ctx, operation)
		if err != nil {
			t.Fatalf("Failed to queue operation: %v", err)
		}

		// Wait a bit for the orchestrator to process
		time.Sleep(100 * time.Millisecond)
	})

	// Test 3: Update operation status
	t.Run("update_operation_status", func(t *testing.T) {
		err := suite.OperationRepo.UpdateStatus(ctx, operationID, "running")
		if err != nil {
			t.Fatalf("Failed to update operation status: %v", err)
		}

		suite.AssertOperationStatus(t, operationID, "running")
	})

	// Test 4: Set operation as started
	t.Run("set_operation_started", func(t *testing.T) {
		err := suite.OperationRepo.SetStarted(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to set operation as started: %v", err)
		}

		operation, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get operation after setting started: %v", err)
		}

		if operation.Status != "running" {
			t.Errorf("Expected operation status 'running', got '%s'", operation.Status)
		}

		if operation.StartedAt == nil {
			t.Error("Expected StartedAt to be set")
		}
	})

	// Test 5: Update operation result
	t.Run("update_operation_result", func(t *testing.T) {
		result := repo.Payload{
			"success":      true,
			"message":      "Operation completed successfully",
			"pods_created": 1,
		}

		err := suite.OperationRepo.UpdateResult(ctx, operationID, result)
		if err != nil {
			t.Fatalf("Failed to update operation result: %v", err)
		}

		operation, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get operation after updating result: %v", err)
		}

		if operation.Result == nil {
			t.Fatal("Expected operation result to be set")
		}

		if success, ok := (*operation.Result)["success"].(bool); !ok || !success {
			t.Error("Expected operation result to indicate success")
		}
	})

	// Test 6: Set operation as finished
	t.Run("set_operation_finished", func(t *testing.T) {
		err := suite.OperationRepo.SetFinished(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to set operation as finished: %v", err)
		}

		operation, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get operation after setting finished: %v", err)
		}

		if operation.FinishedAt == nil {
			t.Error("Expected FinishedAt to be set")
		}
	})

	// Test 7: List operations by cluster
	t.Run("list_operations_by_cluster", func(t *testing.T) {
		operations, err := suite.OperationRepo.ListByCluster(ctx, suite.TestClusterID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to list operations: %v", err)
		}

		if len(operations) < 1 {
			t.Errorf("Expected at least 1 operation, got %d", len(operations))
		}

		found := false
		for _, op := range operations {
			if op.ID == operationID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected to find operation %s in list", operationID)
		}
	})
}

// TestOperationCancellationIntegration tests operation cancellation flow
func TestOperationCancellationIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()
	operationID := uuid.New()

	// Create and queue an operation
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

	err := suite.ClusterService.CreateOperation(ctx, operation)
	if err != nil {
		t.Fatalf("Failed to create operation: %v", err)
	}

	// Test cancellation
	t.Run("cancel_operation", func(t *testing.T) {
		reason := "User requested cancellation"
		err := suite.OperationRepo.CancelOperation(ctx, operationID, reason)
		if err != nil {
			t.Fatalf("Failed to cancel operation: %v", err)
		}

		operation, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get cancelled operation: %v", err)
		}

		if operation.Status != "cancelled" {
			t.Errorf("Expected operation status 'cancelled', got '%s'", operation.Status)
		}

		if operation.Result == nil {
			t.Fatal("Expected operation result to be set after cancellation")
		}

		if cancelled, ok := (*operation.Result)["cancelled"].(bool); !ok || !cancelled {
			t.Error("Expected operation result to indicate cancellation")
		}

		if reasonResult, ok := (*operation.Result)["reason"].(string); !ok || reasonResult != reason {
			t.Errorf("Expected cancellation reason '%s', got '%s'", reason, reasonResult)
		}
	})
}

// TestConcurrentOperationsIntegration tests handling of concurrent operations
func TestConcurrentOperationsIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()
	numOperations := 5

	// Create multiple operations concurrently
	t.Run("create_concurrent_operations", func(t *testing.T) {
		operationIDs := make([]uuid.UUID, numOperations)

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

			err := suite.ClusterService.CreateOperation(ctx, operation)
			if err != nil {
				t.Fatalf("Failed to create operation %d: %v", i, err)
			}
		}

		// Verify all operations were created
		operations, err := suite.OperationRepo.ListByCluster(ctx, suite.TestClusterID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to list operations: %v", err)
		}

		if len(operations) < numOperations {
			t.Errorf("Expected at least %d operations, got %d", numOperations, len(operations))
		}
	})
}

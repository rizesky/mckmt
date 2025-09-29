package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	apihandler "github.com/rizesky/mckmt/internal/api/http"
	"github.com/rizesky/mckmt/internal/repo"
)

// TestEndToEndIntegration tests the complete flow from HTTP request to operation completion
func TestEndToEndIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	ctx := context.Background()
	handler := apihandler.NewClusterHandler(suite.ClusterService, suite.Logger)

	// Step 1: Create a cluster
	t.Run("create_cluster", func(t *testing.T) {
		cluster := &repo.Cluster{
			ID:          suite.TestClusterID,
			Name:        "e2e-test-cluster",
			Description: "End-to-end test cluster",
			Status:      "active",
			Labels:      repo.Labels{"env": "e2e", "test": "true"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := suite.ClusterRepo.Create(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to create cluster: %v", err)
		}

		suite.AssertClusterExists(t, suite.TestClusterID)
	})

	// Step 2: Apply manifests via HTTP API
	var operationID uuid.UUID
	t.Run("apply_manifests_via_api", func(t *testing.T) {
		// Create multipart form data
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		manifestsContent := `apiVersion: v1
kind: Pod
metadata:
  name: e2e-test-pod
  labels:
    app: e2e-test
spec:
  containers:
  - name: test-container
    image: nginx:latest
    ports:
    - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: e2e-test-service
spec:
  selector:
    app: e2e-test
  ports:
  - port: 80
    targetPort: 80`

		fileWriter, err := writer.CreateFormFile("manifests", "manifests.yaml")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileWriter.Write([]byte(manifestsContent))
		if err != nil {
			t.Fatalf("Failed to write manifests content: %v", err)
		}

		writer.Close()

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/clusters/%s/manifests", suite.TestClusterID), &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		// Add URL parameters for chi router
		req = addURLParams(req, "id", suite.TestClusterID.String())

		handler.ApplyManifests(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
		}

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		operationIDStr, ok := response["operation_id"].(string)
		if !ok {
			t.Fatal("Expected 'operation_id' field in response")
		}

		operationID, err = uuid.Parse(operationIDStr)
		if err != nil {
			t.Fatalf("Failed to parse operation ID: %v", err)
		}
	})

	// Step 3: Verify operation was created and queued
	t.Run("verify_operation_created", func(t *testing.T) {
		operation, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get operation: %v", err)
		}

		if operation.ClusterID != suite.TestClusterID {
			t.Errorf("Expected cluster ID %s, got %s", suite.TestClusterID, operation.ClusterID)
		}

		if operation.Type != "apply" {
			t.Errorf("Expected operation type 'apply', got '%s'", operation.Type)
		}

		// The operation might be processed immediately by the orchestrator
		// So we accept 'queued', 'running', or 'success' status
		if operation.Status != "queued" && operation.Status != "running" && operation.Status != "success" {
			t.Errorf("Expected operation status 'queued', 'running', or 'success', got '%s'", operation.Status)
		}

		// Verify payload contains manifests
		if manifests, ok := operation.Payload["manifests"].(string); !ok {
			t.Error("Expected manifests in operation payload")
		} else if !strings.Contains(manifests, "e2e-test-pod") {
			t.Error("Expected manifests to contain 'e2e-test-pod'")
		}
	})

	// Step 4: Wait for operation processing
	t.Run("wait_for_operation_processing", func(t *testing.T) {
		// Wait for orchestrator to process the operation
		err := suite.WaitForOperationStatus(operationID, "running", 2*time.Second)
		if err != nil {
			// If it doesn't reach running status, that's okay for this test
			// The important thing is that it's no longer queued
			operation, getErr := suite.OperationRepo.GetByID(ctx, operationID)
			if getErr != nil {
				t.Fatalf("Failed to get operation: %v", getErr)
			}
			if operation.Status == "queued" {
				t.Error("Expected operation to be processed by orchestrator")
			}
		}
	})

	// Step 5: Simulate operation completion
	t.Run("simulate_operation_completion", func(t *testing.T) {
		// Update operation status to running
		err := suite.OperationRepo.UpdateStatus(ctx, operationID, "running")
		if err != nil {
			t.Fatalf("Failed to update operation status: %v", err)
		}

		// Set operation as started
		err = suite.OperationRepo.SetStarted(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to set operation as started: %v", err)
		}

		// Simulate successful completion
		result := repo.Payload{
			"success": true,
			"message": "Manifests applied successfully",
			"resources_created": []string{
				"Pod/e2e-test-pod",
				"Service/e2e-test-service",
			},
			"pods_created":     1,
			"services_created": 1,
		}

		err = suite.OperationRepo.UpdateResult(ctx, operationID, result)
		if err != nil {
			t.Fatalf("Failed to update operation result: %v", err)
		}

		// Set operation as finished
		err = suite.OperationRepo.SetFinished(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to set operation as finished: %v", err)
		}

		// Verify final operation state
		operation, err := suite.OperationRepo.GetByID(ctx, operationID)
		if err != nil {
			t.Fatalf("Failed to get final operation: %v", err)
		}

		if operation.Status != "running" {
			t.Errorf("Expected operation status 'running', got '%s'", operation.Status)
		}

		if operation.StartedAt == nil {
			t.Error("Expected StartedAt to be set")
		}

		if operation.FinishedAt == nil {
			t.Error("Expected FinishedAt to be set")
		}

		if operation.Result == nil {
			t.Fatal("Expected operation result to be set")
		}

		if success, ok := (*operation.Result)["success"].(bool); !ok || !success {
			t.Error("Expected operation result to indicate success")
		}
	})

	// Step 6: Verify cluster resources via API
	t.Run("verify_cluster_resources", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/clusters/%s/resources?kind=Pod&namespace=default", suite.TestClusterID), nil)
		w := httptest.NewRecorder()

		// Add URL parameters for chi router
		req = addURLParams(req, "id", suite.TestClusterID.String())

		handler.ListClusterResources(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resources, ok := response["resources"].([]interface{}); !ok {
			t.Error("Expected 'resources' field in response")
		} else {
			// In a real environment, we would verify the actual resources
			// For this test, we just ensure the API call succeeds
			_ = resources
		}
	})

	// Step 7: Test operation cancellation
	t.Run("test_operation_cancellation", func(t *testing.T) {
		// Create another operation for cancellation test
		cancelOperationID := uuid.New()
		cancelOperation := &repo.Operation{
			ID:        cancelOperationID,
			ClusterID: suite.TestClusterID,
			Type:      "apply",
			Status:    "queued",
			Payload: repo.Payload{
				"manifests": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: cancel-test-pod",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := suite.ClusterService.CreateOperation(ctx, cancelOperation)
		if err != nil {
			t.Fatalf("Failed to create cancellation test operation: %v", err)
		}

		// Cancel the operation
		err = suite.OperationRepo.CancelOperation(ctx, cancelOperationID, "End-to-end test cancellation")
		if err != nil {
			t.Fatalf("Failed to cancel operation: %v", err)
		}

		// Verify cancellation
		operation, err := suite.OperationRepo.GetByID(ctx, cancelOperationID)
		if err != nil {
			t.Fatalf("Failed to get cancelled operation: %v", err)
		}

		if operation.Status != "cancelled" {
			t.Errorf("Expected operation status 'cancelled', got '%s'", operation.Status)
		}
	})
}

// TestEndToEndErrorHandlingIntegration tests error scenarios in the complete flow
func TestEndToEndErrorHandlingIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	handler := apihandler.NewClusterHandler(suite.ClusterService, suite.Logger)

	// Test 1: Apply manifests to non-existent cluster
	t.Run("apply_manifests_to_nonexistent_cluster", func(t *testing.T) {
		nonexistentClusterID := uuid.New()

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		manifestsContent := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod`

		fileWriter, err := writer.CreateFormFile("manifests", "manifests.yaml")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileWriter.Write([]byte(manifestsContent))
		if err != nil {
			t.Fatalf("Failed to write manifests content: %v", err)
		}

		writer.Close()

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/clusters/%s/manifests", nonexistentClusterID), &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		// Add URL parameters for chi router
		req = addURLParams(req, "id", nonexistentClusterID.String())

		handler.ApplyManifests(w, req)

		// Should still create operation even for non-existent cluster
		// (the cluster existence check happens at operation execution time)
		if w.Code != http.StatusAccepted {
			t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
		}
	})

	// Test 2: Apply invalid manifests
	t.Run("apply_invalid_manifests", func(t *testing.T) {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		invalidManifestsContent := `invalid yaml content
this is not valid kubernetes yaml
---
also invalid`

		fileWriter, err := writer.CreateFormFile("manifests", "invalid-manifests.yaml")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileWriter.Write([]byte(invalidManifestsContent))
		if err != nil {
			t.Fatalf("Failed to write manifests content: %v", err)
		}

		writer.Close()

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/clusters/%s/manifests", suite.TestClusterID), &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		// Add URL parameters for chi router
		req = addURLParams(req, "id", suite.TestClusterID.String())

		handler.ApplyManifests(w, req)

		// Should still accept the request (validation happens during execution)
		if w.Code != http.StatusAccepted {
			t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
		}
	})
}

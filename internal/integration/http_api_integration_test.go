package integration

import (
	"bytes"
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
)

// TestHTTPAPIIntegration tests the complete HTTP API flow
func TestHTTPAPIIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Create HTTP handler
	handler := apihandler.NewClusterHandler(suite.ClusterService, suite.Logger)

	// Test 1: GET /clusters
	t.Run("list_clusters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/clusters?limit=10&offset=0", nil)
		w := httptest.NewRecorder()

		handler.ListClusters(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if clusters, ok := response["clusters"].([]interface{}); !ok {
			t.Error("Expected 'clusters' field in response")
		} else if len(clusters) < 1 {
			t.Error("Expected at least 1 cluster in response")
		}
	})

	// Test 2: GET /clusters/{id}
	t.Run("get_cluster", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/clusters/%s", suite.TestClusterID), nil)
		w := httptest.NewRecorder()

		// Add URL parameters for chi router
		req = addURLParams(req, "id", suite.TestClusterID.String())

		handler.GetCluster(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var cluster map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &cluster)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if cluster["id"] != suite.TestClusterID.String() {
			t.Errorf("Expected cluster ID %s, got %s", suite.TestClusterID, cluster["id"])
		}
	})

	// Test 3: POST /clusters/{id}/manifests (ApplyManifests)
	t.Run("apply_manifests", func(t *testing.T) {
		// Create multipart form data
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		// Add manifests file
		manifestsContent := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: nginx:latest`

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

		if operationID, ok := response["operation_id"].(string); !ok {
			t.Error("Expected 'operation_id' field in response")
		} else if operationID == "" {
			t.Error("Expected non-empty operation ID")
		}
	})

	// Test 4: GET /clusters/{id}/resources
	t.Run("list_cluster_resources", func(t *testing.T) {
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
			// Resources might be empty in test environment, that's okay
			_ = resources
		}
	})
}

// TestHTTPAPIErrorHandlingIntegration tests error handling in HTTP API
func TestHTTPAPIErrorHandlingIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	handler := apihandler.NewClusterHandler(suite.ClusterService, suite.Logger)

	// Test 1: GET /clusters/{invalid-id}
	t.Run("get_nonexistent_cluster", func(t *testing.T) {
		invalidID := uuid.New()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/clusters/%s", invalidID), nil)
		w := httptest.NewRecorder()

		// Add URL parameters for chi router
		req = addURLParams(req, "id", invalidID.String())

		handler.GetCluster(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if errorMsg, ok := response["error"].(string); !ok {
			t.Error("Expected 'error' field in response")
		} else if !strings.Contains(errorMsg, "not found") {
			t.Errorf("Expected error message to contain 'not found', got: %s", errorMsg)
		}
	})

	// Test 2: GET /clusters with invalid parameters
	t.Run("list_clusters_invalid_params", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/clusters?limit=invalid&offset=invalid", nil)
		w := httptest.NewRecorder()

		handler.ListClusters(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if errorMsg, ok := response["error"].(string); !ok {
			t.Error("Expected 'error' field in response")
		} else if !strings.Contains(errorMsg, "Invalid") {
			t.Errorf("Expected error message to contain 'Invalid', got: %s", errorMsg)
		}
	})

	// Test 3: POST /clusters/{id}/manifests with no file
	t.Run("apply_manifests_no_file", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/clusters/%s/manifests", suite.TestClusterID), nil)
		req.Header.Set("Content-Type", "multipart/form-data")
		w := httptest.NewRecorder()

		// Add URL parameters for chi router
		req = addURLParams(req, "id", suite.TestClusterID.String())

		handler.ApplyManifests(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if errorMsg, ok := response["error"].(string); !ok {
			t.Error("Expected 'error' field in response")
		} else if !strings.Contains(errorMsg, "Failed to parse multipart form") {
			t.Errorf("Expected error message to contain 'Failed to parse multipart form', got: %s", errorMsg)
		}
	})
}

// TestHTTPAPIPerformanceIntegration tests API performance with multiple requests
func TestHTTPAPIPerformanceIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	handler := apihandler.NewClusterHandler(suite.ClusterService, suite.Logger)

	// Test concurrent requests
	t.Run("concurrent_requests", func(t *testing.T) {
		numRequests := 10
		done := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(index int) {
				req := httptest.NewRequest("GET", "/api/v1/clusters?limit=10&offset=0", nil)
				w := httptest.NewRecorder()

				handler.ListClusters(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Request %d failed with status %d", index, w.Code)
				}

				done <- true
			}(i)
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			select {
			case <-done:
				// Request completed
			case <-time.After(5 * time.Second):
				t.Fatalf("Request %d timed out", i)
			}
		}
	})
}

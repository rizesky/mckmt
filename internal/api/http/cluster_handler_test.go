package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/api/http/mocks"
	"github.com/rizesky/mckmt/internal/repo"
)

func TestClusterHandler_GetCluster(t *testing.T) {
	tests := []struct {
		name           string
		clusterID      string
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "successful get cluster",
			clusterID:      uuid.New().String(),
			serviceError:   nil,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid cluster ID",
			clusterID:      "invalid-uuid",
			serviceError:   nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "cluster not found",
			clusterID:      uuid.New().String(),
			serviceError:   errors.New("not found"),
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name:           "service error",
			clusterID:      uuid.New().String(),
			serviceError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClusterService := mocks.NewMockClusterManager(ctrl)
			logger := zap.NewNop()

			// Setup expectations
			if tt.clusterID != "invalid-uuid" {
				clusterID, err := uuid.Parse(tt.clusterID)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if tt.serviceError != nil {
					mockClusterService.EXPECT().
						GetCluster(gomock.Any(), clusterID).
						Return(nil, tt.serviceError)
				} else {
					mockClusterService.EXPECT().
						GetCluster(gomock.Any(), clusterID).
						Return(&repo.Cluster{
							ID:   clusterID,
							Name: "test-cluster",
						}, nil)
				}
			}

			handler := NewClusterHandler(mockClusterService, logger)

			// Create request with chi router context
			req := httptest.NewRequest("GET", fmt.Sprintf("/clusters/%s", tt.clusterID), nil)
			req = req.WithContext(context.Background())

			// Set up chi router context with URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.clusterID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handler.GetCluster(w, req)

			// Verify
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Errorf("Expected response to contain 'error' field: %+v", response)
				}
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if _, ok := response["id"]; !ok {
					t.Errorf("Expected response to contain 'id' field: %+v", response)
				}
				if _, ok := response["name"]; !ok {
					t.Errorf("Expected response to contain 'name' field: %+v", response)
				}
			}
		})
	}
}

func TestClusterHandler_ListClusters(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "successful list clusters",
			queryParams:    "?limit=10&offset=0",
			serviceError:   nil,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid limit parameter",
			queryParams:    "?limit=invalid",
			serviceError:   nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "service error",
			queryParams:    "?limit=10&offset=0",
			serviceError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClusterService := mocks.NewMockClusterManager(ctrl)
			logger := zap.NewNop()

			// Setup expectations
			if tt.queryParams != "?limit=invalid" {
				if tt.serviceError != nil {
					mockClusterService.EXPECT().
						ListClusters(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil, tt.serviceError)
				} else {
					mockClusterService.EXPECT().
						ListClusters(gomock.Any(), gomock.Any(), gomock.Any()).
						Return([]*repo.Cluster{
							{ID: uuid.New(), Name: "cluster1"},
							{ID: uuid.New(), Name: "cluster2"},
						}, nil)
				}
			}

			handler := NewClusterHandler(mockClusterService, logger)

			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/clusters%s", tt.queryParams), nil)
			req = req.WithContext(context.Background())

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handler.ListClusters(w, req)

			// Verify
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Errorf("Expected response to contain 'error' field: %+v", response)
				}
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if _, ok := response["clusters"]; !ok {
					t.Errorf("Expected response to contain 'clusters' field: %+v", response)
				}
			}
		})
	}
}

func TestClusterHandler_ApplyManifests(t *testing.T) {
	tests := []struct {
		name           string
		clusterID      string
		manifests      string
		createError    error
		queueError     error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "successful apply manifests",
			clusterID:      uuid.New().String(),
			manifests:      "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test-pod",
			createError:    nil,
			queueError:     nil,
			expectedStatus: http.StatusAccepted,
			expectedError:  false,
		},
		{
			name:           "invalid cluster ID",
			clusterID:      "invalid-uuid",
			manifests:      "apiVersion: v1\nkind: Pod",
			createError:    nil,
			queueError:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "no manifests file",
			clusterID:      uuid.New().String(),
			manifests:      "",
			createError:    nil,
			queueError:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "create operation error",
			clusterID:      uuid.New().String(),
			manifests:      "apiVersion: v1\nkind: Pod",
			createError:    errors.New("database error"),
			queueError:     nil,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:           "queue operation error",
			clusterID:      uuid.New().String(),
			manifests:      "apiVersion: v1\nkind: Pod",
			createError:    nil,
			queueError:     errors.New("queue full"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClusterService := mocks.NewMockClusterManager(ctrl)
			logger := zap.NewNop()

			// Setup expectations
			if tt.clusterID != "invalid-uuid" && tt.manifests != "" {
				_, err := uuid.Parse(tt.clusterID)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Expect CreateOperation call
				if tt.createError != nil {
					mockClusterService.EXPECT().
						CreateOperation(gomock.Any(), gomock.Any()).
						Return(tt.createError)
				} else {
					mockClusterService.EXPECT().
						CreateOperation(gomock.Any(), gomock.Any()).
						Return(nil)

					// Expect QueueOperation call only if create succeeded
					if tt.queueError != nil {
						mockClusterService.EXPECT().
							QueueOperation(gomock.Any(), gomock.Any()).
							Return(tt.queueError)
					} else {
						mockClusterService.EXPECT().
							QueueOperation(gomock.Any(), gomock.Any()).
							Return(nil)
					}
				}
			}

			handler := NewClusterHandler(mockClusterService, logger)

			// Create multipart form request
			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			if tt.manifests != "" {
				fw, err := w.CreateFormFile("manifests", "test.yaml")
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				_, err = fw.Write([]byte(tt.manifests))
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			err := w.Close()
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Create request with chi router context
			req := httptest.NewRequest("POST", fmt.Sprintf("/clusters/%s/manifests", tt.clusterID), &b)
			req.Header.Set("Content-Type", w.FormDataContentType())
			req = req.WithContext(context.Background())

			// Set up chi router context with URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.clusterID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute
			handler.ApplyManifests(rr, req)

			// Verify
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Errorf("Expected response to contain 'error' field: %+v", response)
				}
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if _, ok := response["operation_id"]; !ok {
					t.Errorf("Expected response to contain 'operation_id' field: %+v", response)
				}
				if _, ok := response["status"]; !ok {
					t.Errorf("Expected response to contain 'status' field: %+v", response)
				}
				if _, ok := response["message"]; !ok {
					t.Errorf("Expected response to contain 'message' field: %+v", response)
				}
			}
		})
	}
}

func TestClusterHandler_ListClusterResources(t *testing.T) {
	tests := []struct {
		name           string
		clusterID      string
		queryParams    string
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "successful list resources",
			clusterID:      uuid.New().String(),
			queryParams:    "?kind=Pod&namespace=default",
			serviceError:   nil,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid cluster ID",
			clusterID:      "invalid-uuid",
			queryParams:    "",
			serviceError:   nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "service error",
			clusterID:      uuid.New().String(),
			queryParams:    "",
			serviceError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClusterService := mocks.NewMockClusterManager(ctrl)
			logger := zap.NewNop()

			// Setup expectations
			if tt.clusterID != "invalid-uuid" {
				clusterID, err := uuid.Parse(tt.clusterID)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if tt.serviceError != nil {
					mockClusterService.EXPECT().
						GetClusterResources(gomock.Any(), clusterID, gomock.Any(), gomock.Any()).
						Return(nil, tt.serviceError)
				} else {
					mockClusterService.EXPECT().
						GetClusterResources(gomock.Any(), clusterID, gomock.Any(), gomock.Any()).
						Return([]map[string]interface{}{
							{
								"kind":      "Pod",
								"name":      "test-pod",
								"namespace": "default",
								"status":    "Running",
							},
							{
								"kind":      "Service",
								"name":      "test-service",
								"namespace": "default",
								"type":      "ClusterIP",
							},
						}, nil)
				}
			}

			handler := NewClusterHandler(mockClusterService, logger)

			// Create request with chi router context
			req := httptest.NewRequest("GET", fmt.Sprintf("/clusters/%s/resources%s", tt.clusterID, tt.queryParams), nil)
			req = req.WithContext(context.Background())

			// Set up chi router context with URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.clusterID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			handler.ListClusterResources(w, req)

			// Verify
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Errorf("Expected response to contain 'error' field: %+v", response)
				}
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if _, ok := response["resources"]; !ok {
					t.Errorf("Expected response to contain 'resources' field: %+v", response)
				}
			}
		})
	}
}

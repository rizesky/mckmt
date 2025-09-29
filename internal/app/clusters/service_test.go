package clusters

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	clustermocks "github.com/rizesky/mckmt/internal/app/clusters/mocks"
	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/repo/mocks"
)

func TestClusterService_CreateOperation(t *testing.T) {
	tests := []struct {
		name          string
		operation     *repo.Operation
		repoError     error
		expectedError bool
	}{
		{
			name: "successful operation creation",
			operation: &repo.Operation{
				ID:        uuid.New(),
				ClusterID: uuid.New(),
				Type:      "apply",
				Status:    "queued",
				Payload:   repo.Payload{"test": "data"},
			},
			repoError:     nil,
			expectedError: false,
		},
		{
			name: "repository error",
			operation: &repo.Operation{
				ID:        uuid.New(),
				ClusterID: uuid.New(),
				Type:      "apply",
				Status:    "queued",
				Payload:   repo.Payload{"test": "data"},
			},
			repoError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOpRepo := mocks.NewMockOperationRepository(ctrl)
			mockClusterRepo := mocks.NewMockClusterRepository(ctrl)
			mockCache := mocks.NewMockCache(ctrl)
			mockOrchestrator := clustermocks.NewMockOrchestratorInterface(ctrl)
			logger := zap.NewNop()

			// Setup expectations
			if tt.repoError != nil {
				mockOpRepo.EXPECT().
					Create(gomock.Any(), tt.operation).
					Return(tt.repoError)
			} else {
				mockOpRepo.EXPECT().
					Create(gomock.Any(), tt.operation).
					Return(nil)
			}

			service := NewClusterService(mockClusterRepo, mockOpRepo, mockCache, logger, mockOrchestrator)

			// Execute
			err := service.CreateOperation(context.Background(), tt.operation)

			// Verify
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestClusterService_QueueOperation(t *testing.T) {
	tests := []struct {
		name              string
		operation         *repo.Operation
		orchestratorError error
		expectedError     bool
	}{
		{
			name: "successful operation queuing",
			operation: &repo.Operation{
				ID:        uuid.New(),
				ClusterID: uuid.New(),
				Type:      "apply",
				Status:    "queued",
				Payload:   repo.Payload{"test": "data"},
			},
			orchestratorError: nil,
			expectedError:     false,
		},
		{
			name: "orchestrator error",
			operation: &repo.Operation{
				ID:        uuid.New(),
				ClusterID: uuid.New(),
				Type:      "apply",
				Status:    "queued",
				Payload:   repo.Payload{"test": "data"},
			},
			orchestratorError: errors.New("queue full"),
			expectedError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOpRepo := mocks.NewMockOperationRepository(ctrl)
			mockClusterRepo := mocks.NewMockClusterRepository(ctrl)
			mockCache := mocks.NewMockCache(ctrl)
			mockOrchestrator := clustermocks.NewMockOrchestratorInterface(ctrl)
			logger := zap.NewNop()

			// Setup expectations
			if tt.orchestratorError != nil {
				mockOrchestrator.EXPECT().
					QueueOperation(tt.operation).
					Return(tt.orchestratorError)
			} else {
				mockOrchestrator.EXPECT().
					QueueOperation(tt.operation).
					Return(nil)
			}

			service := NewClusterService(mockClusterRepo, mockOpRepo, mockCache, logger, mockOrchestrator)

			// Execute
			err := service.QueueOperation(context.Background(), tt.operation)

			// Verify
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestClusterService_GetCluster(t *testing.T) {
	tests := []struct {
		name            string
		clusterID       uuid.UUID
		setupCluster    *repo.Cluster
		cacheError      error
		repoError       error
		expectedError   bool
		expectedCluster *repo.Cluster
	}{
		{
			name:      "successful get from cache",
			clusterID: uuid.New(),
			setupCluster: &repo.Cluster{
				ID:   uuid.New(),
				Name: "test-cluster",
			},
			cacheError:    nil,
			repoError:     nil,
			expectedError: false,
			expectedCluster: &repo.Cluster{
				ID:   uuid.New(),
				Name: "test-cluster",
			},
		},
		{
			name:      "cache miss, get from repository",
			clusterID: uuid.New(),
			setupCluster: &repo.Cluster{
				ID:   uuid.New(),
				Name: "test-cluster",
			},
			cacheError:    repo.ErrCacheMiss,
			repoError:     nil,
			expectedError: false,
			expectedCluster: &repo.Cluster{
				ID:   uuid.New(),
				Name: "test-cluster",
			},
		},
		{
			name:            "repository error",
			clusterID:       uuid.New(),
			cacheError:      repo.ErrCacheMiss,
			repoError:       errors.New("database error"),
			expectedError:   true,
			expectedCluster: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOpRepo := mocks.NewMockOperationRepository(ctrl)
			mockClusterRepo := mocks.NewMockClusterRepository(ctrl)
			mockCache := mocks.NewMockCache(ctrl)
			mockOrchestrator := clustermocks.NewMockOrchestratorInterface(ctrl)
			logger := zap.NewNop()

			// Setup cache expectations
			mockCache.EXPECT().
				ClusterKey(gomock.Any()).
				Return("cluster:" + tt.clusterID.String()).
				AnyTimes()

			if tt.cacheError != nil {
				mockCache.EXPECT().
					Get(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(tt.cacheError)
			} else {
				mockCache.EXPECT().
					Get(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, key string, dest interface{}) error {
						if tt.setupCluster != nil {
							// Simulate cache hit by setting the dest
							if destPtr, ok := dest.(*repo.Cluster); ok {
								*destPtr = *tt.setupCluster
							}
						}
						return nil
					})
			}

			// Setup repository expectations for cache miss
			if tt.cacheError == repo.ErrCacheMiss {
				if tt.repoError != nil {
					mockClusterRepo.EXPECT().
						GetByID(gomock.Any(), tt.clusterID).
						Return(nil, tt.repoError)
				} else {
					mockClusterRepo.EXPECT().
						GetByID(gomock.Any(), tt.clusterID).
						Return(tt.setupCluster, nil)

					// Expect cache set
					mockCache.EXPECT().
						Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				}
			}

			service := NewClusterService(mockClusterRepo, mockOpRepo, mockCache, logger, mockOrchestrator)

			// Execute
			cluster, err := service.GetCluster(context.Background(), tt.clusterID)

			// Verify
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				if cluster != nil {
					t.Errorf("Expected nil cluster but got %+v", cluster)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.expectedCluster != nil {
					if cluster == nil {
						t.Errorf("Expected cluster but got nil")
					} else if cluster.Name != tt.expectedCluster.Name {
						t.Errorf("Expected cluster name %s but got %s", tt.expectedCluster.Name, cluster.Name)
					}
				}
			}
		})
	}
}

func TestClusterService_ListClusters(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		offset        int
		setupClusters []*repo.Cluster
		repoError     error
		expectedError bool
		expectedCount int
	}{
		{
			name:   "successful list",
			limit:  10,
			offset: 0,
			setupClusters: []*repo.Cluster{
				{ID: uuid.New(), Name: "cluster1"},
				{ID: uuid.New(), Name: "cluster2"},
			},
			repoError:     nil,
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:          "repository error",
			limit:         10,
			offset:        0,
			repoError:     errors.New("database error"),
			expectedError: true,
			expectedCount: 0,
		},
		{
			name:   "pagination test",
			limit:  1,
			offset: 1,
			setupClusters: []*repo.Cluster{
				{ID: uuid.New(), Name: "cluster1"},
				{ID: uuid.New(), Name: "cluster2"},
				{ID: uuid.New(), Name: "cluster3"},
			},
			repoError:     nil,
			expectedError: false,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOpRepo := mocks.NewMockOperationRepository(ctrl)
			mockClusterRepo := mocks.NewMockClusterRepository(ctrl)
			mockCache := mocks.NewMockCache(ctrl)
			mockOrchestrator := clustermocks.NewMockOrchestratorInterface(ctrl)
			logger := zap.NewNop()

			// Setup expectations
			if tt.repoError != nil {
				mockClusterRepo.EXPECT().
					List(gomock.Any(), tt.limit, tt.offset).
					Return(nil, tt.repoError)
			} else {
				mockClusterRepo.EXPECT().
					List(gomock.Any(), tt.limit, tt.offset).
					DoAndReturn(func(ctx context.Context, limit, offset int) ([]*repo.Cluster, error) {
						// Simulate pagination
						start := offset
						end := offset + limit
						if start >= len(tt.setupClusters) {
							return []*repo.Cluster{}, nil
						}
						if end > len(tt.setupClusters) {
							end = len(tt.setupClusters)
						}
						return tt.setupClusters[start:end], nil
					})
			}

			service := NewClusterService(mockClusterRepo, mockOpRepo, mockCache, logger, mockOrchestrator)

			// Execute
			clusters, err := service.ListClusters(context.Background(), tt.limit, tt.offset)

			// Verify
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				if clusters != nil {
					t.Errorf("Expected nil clusters but got %+v", clusters)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if len(clusters) != tt.expectedCount {
					t.Errorf("Expected %d clusters but got %d", tt.expectedCount, len(clusters))
				}
			}
		})
	}
}

func TestClusterService_GetClusterResources(t *testing.T) {
	tests := []struct {
		name          string
		clusterID     uuid.UUID
		kind          string
		namespace     string
		expectedError bool
		expectedCount int
	}{
		{
			name:          "successful resource list",
			clusterID:     uuid.New(),
			kind:          "Pod",
			namespace:     "default",
			expectedError: false,
			expectedCount: 2, // Based on mock implementation
		},
		{
			name:          "empty kind filter",
			clusterID:     uuid.New(),
			kind:          "",
			namespace:     "",
			expectedError: false,
			expectedCount: 2, // Based on mock implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOpRepo := mocks.NewMockOperationRepository(ctrl)
			mockClusterRepo := mocks.NewMockClusterRepository(ctrl)
			mockCache := mocks.NewMockCache(ctrl)
			mockOrchestrator := clustermocks.NewMockOrchestratorInterface(ctrl)
			logger := zap.NewNop()

			// Setup cache expectations
			mockCache.EXPECT().
				ClusterResourcesKey(gomock.Any(), gomock.Any(), gomock.Any()).
				Return("resources:" + tt.clusterID.String() + ":" + tt.kind + ":" + tt.namespace).
				AnyTimes()

			mockCache.EXPECT().
				Get(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(repo.ErrCacheMiss) // Always cache miss for this test

			mockCache.EXPECT().
				Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil) // Cache set after getting resources

			service := NewClusterService(mockClusterRepo, mockOpRepo, mockCache, logger, mockOrchestrator)

			// Execute
			resources, err := service.GetClusterResources(context.Background(), tt.clusterID, tt.kind, tt.namespace)

			// Verify
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				if resources != nil {
					t.Errorf("Expected nil resources but got %+v", resources)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if len(resources) != tt.expectedCount {
					t.Errorf("Expected %d resources but got %d", tt.expectedCount, len(resources))
				}

				// Verify resource structure
				for _, resource := range resources {
					if _, ok := resource["kind"]; !ok {
						t.Errorf("Resource missing 'kind' field: %+v", resource)
					}
					if _, ok := resource["name"]; !ok {
						t.Errorf("Resource missing 'name' field: %+v", resource)
					}
					if _, ok := resource["namespace"]; !ok {
						t.Errorf("Resource missing 'namespace' field: %+v", resource)
					}
					if _, ok := resource["created_at"]; !ok {
						t.Errorf("Resource missing 'created_at' field: %+v", resource)
					}
				}
			}
		})
	}
}

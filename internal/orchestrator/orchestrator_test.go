package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/orchestrator/mocks"
	"github.com/rizesky/mckmt/internal/repo"
	repomocks "github.com/rizesky/mckmt/internal/repo/mocks"
)

func TestOrchestrator_QueueOperation(t *testing.T) {
	tests := []struct {
		name          string
		operation     *repo.Operation
		repoError     error
		expectedError bool
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
			repoError:     nil,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOpRepo := repomocks.NewMockOperationRepository(ctrl)
			mockMetrics := mocks.NewMockMetricsProvider(ctrl)
			logger := zap.NewNop()

			// QueueOperation doesn't call Create, it just queues the operation
			// No expectations needed for this test

			orchestrator := NewOrchestrator(mockOpRepo, mockMetrics, logger, 1)

			// Execute
			err := orchestrator.QueueOperation(tt.operation)

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

func TestOrchestrator_CancelOperation(t *testing.T) {
	tests := []struct {
		name          string
		operationID   uuid.UUID
		expectedError bool
	}{
		{
			name:          "successful cancellation request",
			operationID:   uuid.New(),
			expectedError: false,
		},
		{
			name:          "another successful cancellation request",
			operationID:   uuid.New(),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOpRepo := repomocks.NewMockOperationRepository(ctrl)
			mockMetrics := mocks.NewMockMetricsProvider(ctrl)
			logger := zap.NewNop()

			// CancelOperation just queues the cancellation request
			// No repository calls are made directly

			orchestrator := NewOrchestrator(mockOpRepo, mockMetrics, logger, 1)

			// Execute
			err := orchestrator.CancelOperation(tt.operationID)

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

func TestOrchestrator_ProcessOperations(t *testing.T) {
	tests := []struct {
		name           string
		operations     []*repo.Operation
		expectedStatus string
	}{
		{
			name: "process apply operation",
			operations: []*repo.Operation{
				{
					ID:        uuid.New(),
					ClusterID: uuid.New(),
					Type:      "apply",
					Status:    "queued",
					Payload:   repo.Payload{"manifests": "apiVersion: v1\nkind: Pod"},
				},
			},
			expectedStatus: "queued", // Will be queued for agent processing
		},
		{
			name: "process exec operation",
			operations: []*repo.Operation{
				{
					ID:        uuid.New(),
					ClusterID: uuid.New(),
					Type:      "exec",
					Status:    "queued",
					Payload:   repo.Payload{"command": "kubectl get pods"},
				},
			},
			expectedStatus: "queued", // Will be queued for agent processing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOpRepo := repomocks.NewMockOperationRepository(ctrl)
			mockMetrics := mocks.NewMockMetricsProvider(ctrl)
			logger := zap.NewNop()

			// Setup expectations for operation processing
			for _, op := range tt.operations {
				// Expect GetByID call to retrieve the operation
				mockOpRepo.EXPECT().
					GetByID(gomock.Any(), op.ID).
					Return(op, nil)

				// Expect SetStarted call
				mockOpRepo.EXPECT().
					SetStarted(gomock.Any(), op.ID).
					Return(nil)

				// Expect metrics calls
				mockMetrics.EXPECT().
					IncOperationsInProgress(op.ClusterID.String(), op.Type).
					Return()

				mockMetrics.EXPECT().
					DecOperationsInProgress(op.ClusterID.String(), op.Type).
					Return()

				mockMetrics.EXPECT().
					RecordOperation(op.ClusterID.String(), op.Type, "success", gomock.Any()).
					Return()

				// Expect UpdateStatus call with final status (success)
				mockOpRepo.EXPECT().
					UpdateStatus(gomock.Any(), op.ID, "success").
					Return(nil)

				// Expect UpdateResult call
				mockOpRepo.EXPECT().
					UpdateResult(gomock.Any(), op.ID, gomock.Any()).
					Return(nil)

				// Expect SetFinished call
				mockOpRepo.EXPECT().
					SetFinished(gomock.Any(), op.ID).
					Return(nil)
			}

			orchestrator := NewOrchestrator(mockOpRepo, mockMetrics, logger, 1)

			// Start orchestrator in background
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			go orchestrator.Start(ctx)

			// Queue operations for processing
			for _, op := range tt.operations {
				err := orchestrator.QueueOperation(op)
				if err != nil {
					t.Errorf("Failed to queue operation: %v", err)
				}
			}

			// Wait a bit for processing
			time.Sleep(50 * time.Millisecond)
		})
	}
}

func TestOrchestrator_WorkerPool(t *testing.T) {
	tests := []struct {
		name     string
		workers  int
		expected int
	}{
		{
			name:     "single worker",
			workers:  1,
			expected: 1,
		},
		{
			name:     "multiple workers",
			workers:  5,
			expected: 5,
		},
		{
			name:     "zero workers (should default to 1)",
			workers:  0,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOpRepo := repomocks.NewMockOperationRepository(ctrl)
			mockMetrics := mocks.NewMockMetricsProvider(ctrl)
			logger := zap.NewNop()

			orchestrator := NewOrchestrator(mockOpRepo, mockMetrics, logger, tt.workers)

			// Verify worker count
			if orchestrator.workers != tt.expected {
				t.Errorf("Expected workers %d, got %d", tt.expected, orchestrator.workers)
			}
		})
	}
}

func TestOrchestrator_ContextCancellation(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOpRepo := repomocks.NewMockOperationRepository(ctrl)
	mockMetrics := mocks.NewMockMetricsProvider(ctrl)
	logger := zap.NewNop()

	orchestrator := NewOrchestrator(mockOpRepo, mockMetrics, logger, 1)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Start orchestrator
	done := make(chan error)
	go func() {
		orchestrator.Start(ctx)
		done <- nil
	}()

	// Wait for context cancellation
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("orchestrator should have stopped due to context cancellation")
	}
}

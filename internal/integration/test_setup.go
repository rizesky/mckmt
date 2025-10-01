package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/cluster"
	"github.com/rizesky/mckmt/internal/config"
	"github.com/rizesky/mckmt/internal/orchestrator"
	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/testutils"
	"go.uber.org/zap"
)

// TestSuite holds all the components needed for integration tests
type TestSuite struct {
	// Configuration
	Config *config.HubConfig

	// Repositories
	ClusterRepo   repo.ClusterRepository
	OperationRepo repo.OperationRepository
	Cache         repo.Cache

	// Services
	ClusterService *cluster.Service
	Orchestrator   *orchestrator.Orchestrator

	// Metrics
	Metrics *TestMetrics

	// Logger
	Logger *zap.Logger

	// Test data
	TestClusterID   uuid.UUID
	TestOperationID uuid.UUID
}

// SetupTestSuite creates a new test suite with all dependencies
func SetupTestSuite(t *testing.T) *TestSuite {
	t.Helper()

	// Skip integration tests if not requested
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Load test configuration
	cfg := loadTestConfig()

	// Create logger
	logger := createTestLogger(t)

	// Create repositories (using mocks for now, can be extended to use real DB)
	clusterRepo := testutils.NewMockClusterRepository()
	operationRepo := testutils.NewMockOperationRepository()
	cache := testutils.NewMockCache()

	// Create test metrics (mock implementation to avoid Prometheus conflicts)
	metricsManager := NewTestMetrics()

	// Create orchestrator
	orchestrator := orchestrator.NewOrchestrator(operationRepo, metricsManager, logger, 2)

	// Create cluster service
	clusterService := cluster.NewService(
		clusterRepo,
		operationRepo,
		cache,
		logger,
		orchestrator,
	)

	// Start orchestrator
	ctx := context.Background()
	go orchestrator.Start(ctx)

	// Create test data
	testClusterID := uuid.New()
	testOperationID := uuid.New()

	// Add test cluster
	testCluster := &repo.Cluster{
		ID:          testClusterID,
		Name:        "test-cluster",
		Description: "Test cluster for integration tests",
		Status:      "active",
		Labels:      repo.Labels{"env": "test"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	clusterRepo.Create(ctx, testCluster)

	return &TestSuite{
		Config:          cfg,
		ClusterRepo:     clusterRepo,
		OperationRepo:   operationRepo,
		Cache:           cache,
		ClusterService:  clusterService,
		Orchestrator:    orchestrator,
		Metrics:         metricsManager,
		Logger:          logger,
		TestClusterID:   testClusterID,
		TestOperationID: testOperationID,
	}
}

// loadTestConfig loads test configuration
func loadTestConfig() *config.HubConfig {
	// Return default test config
	return &config.HubConfig{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Database: "mckmt_test",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       1,
		},
		Orchestrator: config.OrchestratorConfig{
			Workers: 2,
		},
	}
}

// createTestLogger creates a test logger
func createTestLogger(t *testing.T) *zap.Logger {
	t.Helper()

	cfg := config.LoggingConfig{
		Level:      "info",
		Format:     "console",
		Caller:     false,
		Stacktrace: false,
	}

	logger, err := config.InitLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	return logger
}

// Cleanup performs cleanup after tests
func (ts *TestSuite) Cleanup() {
	// Stop orchestrator (with panic recovery)
	defer func() {
		if r := recover(); r != nil {
			// Ignore panic from stopping already stopped orchestrator
		}
	}()
	ts.Orchestrator.Stop()

	// Close logger
	ts.Logger.Sync()
}

// CreateTestOperation creates a test operation
func (ts *TestSuite) CreateTestOperation(operationType, status string) *repo.Operation {
	operation := &repo.Operation{
		ID:        ts.TestOperationID,
		ClusterID: ts.TestClusterID,
		Type:      operationType,
		Status:    status,
		Payload: repo.Payload{
			"test": true,
			"type": operationType,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ts.OperationRepo.Create(context.Background(), operation)
	return operation
}

// WaitForOperationStatus waits for an operation to reach a specific status
func (ts *TestSuite) WaitForOperationStatus(operationID uuid.UUID, expectedStatus string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for operation %s to reach status %s", operationID, expectedStatus)
		case <-ticker.C:
			operation, err := ts.OperationRepo.GetByID(ctx, operationID)
			if err != nil {
				continue
			}
			if operation.Status == expectedStatus {
				return nil
			}
		}
	}
}

// AssertOperationStatus checks if an operation has the expected status
func (ts *TestSuite) AssertOperationStatus(t *testing.T, operationID uuid.UUID, expectedStatus string) {
	t.Helper()

	operation, err := ts.OperationRepo.GetByID(context.Background(), operationID)
	if err != nil {
		t.Fatalf("Failed to get operation %s: %v", operationID, err)
	}

	if operation.Status != expectedStatus {
		t.Errorf("Expected operation status %s, got %s", expectedStatus, operation.Status)
	}
}

// AssertClusterExists checks if a cluster exists
func (ts *TestSuite) AssertClusterExists(t *testing.T, clusterID uuid.UUID) {
	t.Helper()

	_, err := ts.ClusterRepo.GetByID(context.Background(), clusterID)
	if err != nil {
		t.Errorf("Expected cluster %s to exist, but got error: %v", clusterID, err)
	}
}

// AssertClusterNotExists checks if a cluster doesn't exist
func (ts *TestSuite) AssertClusterNotExists(t *testing.T, clusterID uuid.UUID) {
	t.Helper()

	_, err := ts.ClusterRepo.GetByID(context.Background(), clusterID)
	if err == nil {
		t.Errorf("Expected cluster %s to not exist, but it does", clusterID)
	}
}

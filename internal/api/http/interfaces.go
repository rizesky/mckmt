package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
)

//go:generate mockgen -destination=./mocks/mock_http.go -package=mocks github.com/rizesky/mckmt/internal/api/http ClusterManager

// ClusterManager defines the interface for cluster management operations.
//
// This interface is defined here (in the HTTP handlers package) because:
//  1. HTTP handlers are the primary consumers of cluster management operations
//  2. It follows the "define where used" principle - interfaces should be defined
//     where they are consumed, not where they are implemented
//  3. HTTP handlers only need a specific subset of cluster operations (HTTP-specific),
//     so defining it here allows for interface segregation
//  4. This makes testing easier as we can mock exactly what the HTTP handlers need
//  5. It decouples HTTP handlers from the concrete cluster service implementation
//
// Note: This interface represents the HTTP layer's view of cluster operations.
// The actual implementation is provided by the cluster service, but the HTTP layer
// defines what it needs, not what the service provides.
type ClusterManager interface {
	GetCluster(ctx context.Context, id uuid.UUID) (*repo.Cluster, error)
	ListClusters(ctx context.Context, limit, offset int) ([]*repo.Cluster, error)
	UpdateCluster(ctx context.Context, id uuid.UUID, name, description string, labels map[string]string) error
	DeleteCluster(ctx context.Context, id uuid.UUID) error
	GetClusterResources(ctx context.Context, clusterID uuid.UUID, kind, namespace string) ([]map[string]interface{}, error)
	CreateOperation(ctx context.Context, operation *repo.Operation) error
	QueueOperation(ctx context.Context, operation *repo.Operation) error
}

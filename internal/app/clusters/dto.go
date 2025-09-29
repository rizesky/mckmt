package clusters

import (
	"time"

	"github.com/google/uuid"
)

// UpdateClusterRequest represents a request to update a cluster
type UpdateClusterRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Status      string            `json:"status"`
}

// ClusterResponse represents a cluster response
type ClusterResponse struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Mode        string            `json:"mode"` // Always "agent"
	Labels      map[string]string `json:"labels"`
	Status      string            `json:"status"`
	LastSeenAt  *time.Time        `json:"last_seen_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ListClustersRequest represents a request to list clusters
type ListClustersRequest struct {
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}

// ListClustersResponse represents a response to list clusters
type ListClustersResponse struct {
	Clusters   []ClusterResponse `json:"clusters"`
	TotalCount int               `json:"total_count"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}

// GetClusterResourcesRequest represents a request to get cluster resources
type GetClusterResourcesRequest struct {
	ClusterID uuid.UUID `json:"cluster_id"`
	Kind      string    `json:"kind"`
	Namespace string    `json:"namespace"`
}

// GetClusterResourcesResponse represents a response to get cluster resources
type GetClusterResourcesResponse struct {
	ClusterID  uuid.UUID                `json:"cluster_id"`
	Resources  []map[string]interface{} `json:"resources"`
	TotalCount int                      `json:"total_count"`
	Kind       string                   `json:"kind"`
	Namespace  string                   `json:"namespace"`
	Cached     bool                     `json:"cached"`
}

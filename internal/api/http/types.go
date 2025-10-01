package http

import (
	"time"

	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/user"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// ClusterDTO represents a cluster in HTTP responses
type ClusterDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Endpoint    string            `json:"endpoint"`
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	LastSeen    *time.Time        `json:"last_seen,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// OperationDTO represents an operation in HTTP responses
type OperationDTO struct {
	ID          string                 `json:"id"`
	ClusterID   string                 `json:"cluster_id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	FinishedAt  *time.Time             `json:"finished_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// UserDTO represents a user in HTTP responses
type UserDTO struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	AuthSource string    `json:"auth_source"`
	Roles      []string  `json:"roles"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Mapping functions to convert from domain entities to DTOs

// ToClusterDTO converts a repo.Cluster to ClusterDTO
func ToClusterDTO(cluster *repo.Cluster) *ClusterDTO {
	return &ClusterDTO{
		ID:          cluster.ID.String(),
		Name:        cluster.Name,
		Description: cluster.Description,
		Endpoint:    cluster.Endpoint,
		Status:      cluster.Status,
		Labels:      map[string]string(cluster.Labels),
		LastSeen:    cluster.LastSeenAt,
		CreatedAt:   cluster.CreatedAt,
		UpdatedAt:   cluster.UpdatedAt,
	}
}

// ToClusterDTOs converts a slice of repo.Cluster to []ClusterDTO
func ToClusterDTOs(clusters []*repo.Cluster) []*ClusterDTO {
	dtos := make([]*ClusterDTO, len(clusters))
	for i, cluster := range clusters {
		dtos[i] = ToClusterDTO(cluster)
	}
	return dtos
}

// ToOperationDTO converts a repo.Operation to OperationDTO
func ToOperationDTO(operation *repo.Operation) *OperationDTO {
	var result map[string]interface{}
	if operation.Result != nil {
		result = map[string]interface{}(*operation.Result)
	}

	return &OperationDTO{
		ID:          operation.ID.String(),
		ClusterID:   operation.ClusterID.String(),
		Type:        operation.Type,
		Status:      operation.Status,
		Description: "", // Operation doesn't have description field
		Parameters:  map[string]interface{}(operation.Payload),
		Result:      result,
		Error:       "", // Operation doesn't have error field
		StartedAt:   operation.StartedAt,
		FinishedAt:  operation.FinishedAt,
		CreatedAt:   operation.CreatedAt,
		UpdatedAt:   operation.UpdatedAt,
	}
}

// ToOperationDTOs converts a slice of repo.Operation to []OperationDTO
func ToOperationDTOs(operations []*repo.Operation) []*OperationDTO {
	dtos := make([]*OperationDTO, len(operations))
	for i, operation := range operations {
		dtos[i] = ToOperationDTO(operation)
	}
	return dtos
}

// ToUserDTO converts a user.User to UserDTO
func ToUserDTO(user *user.User) *UserDTO {
	// Convert []*user.Role to []string
	roleNames := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleNames[i] = role.Name
	}

	return &UserDTO{
		ID:         user.ID.String(),
		Username:   user.Username,
		Email:      user.Email,
		AuthSource: string(user.AuthSource),
		Roles:      roleNames,
		Active:     user.Active,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

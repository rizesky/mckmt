package operations

import (
	"time"

	"github.com/google/uuid"
)

// CreateOperationRequest represents a request to create an operation
type CreateOperationRequest struct {
	ClusterID uuid.UUID              `json:"cluster_id" validate:"required"`
	Type      string                 `json:"type" validate:"required"`
	Payload   map[string]interface{} `json:"payload"`
}

// UpdateOperationRequest represents a request to update an operation
type UpdateOperationRequest struct {
	Status string                 `json:"status" validate:"omitempty,oneof=pending running completed failed cancelled"`
	Result map[string]interface{} `json:"result"`
}

// OperationResponse represents an operation response
type OperationResponse struct {
	ID         uuid.UUID              `json:"id"`
	ClusterID  uuid.UUID              `json:"cluster_id"`
	Type       string                 `json:"type"`
	Status     string                 `json:"status"`
	Payload    map[string]interface{} `json:"payload"`
	Result     map[string]interface{} `json:"result"`
	StartedAt  *time.Time             `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// ListOperationsRequest represents a request to list operations
type ListOperationsRequest struct {
	ClusterID uuid.UUID `json:"cluster_id"`
	Limit     int       `json:"limit" validate:"min=1,max=100"`
	Offset    int       `json:"offset" validate:"min=0"`
}

// ListOperationsResponse represents a response to list operations
type ListOperationsResponse struct {
	Operations []OperationResponse `json:"operations"`
	TotalCount int                 `json:"total_count"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
}

// UpdateOperationStatusRequest represents a request to update operation status
type UpdateOperationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending running completed failed cancelled"`
}

// UpdateOperationResultRequest represents a request to update operation result
type UpdateOperationResultRequest struct {
	Result map[string]interface{} `json:"result" validate:"required"`
}

// CancelOperationRequest represents a request to cancel an operation
type CancelOperationRequest struct {
	Reason string `json:"reason" validate:"omitempty,max=500"`
}

// CancelOperationResponse represents a response to cancel operation
type CancelOperationResponse struct {
	ID      uuid.UUID `json:"id"`
	Status  string    `json:"status"`
	Message string    `json:"message"`
}

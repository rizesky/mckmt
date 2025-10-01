package postgres

import (
	"time"

	"github.com/google/uuid"
)

// Cluster represents a managed Kubernetes cluster
type Cluster struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	Name                 string     `json:"name" db:"name"`
	Description          string     `json:"description" db:"description"`
	Labels               Labels     `json:"labels" db:"labels"`
	EncryptedCredentials []byte     `json:"-" db:"encrypted_credentials"`
	Status               string     `json:"status" db:"status"`
	LastSeenAt           *time.Time `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// ClusterAgent represents an agent running in a cluster
type ClusterAgent struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	ClusterID     uuid.UUID  `json:"cluster_id" db:"cluster_id"`
	AgentVersion  string     `json:"agent_version" db:"agent_version"`
	ConnectedAt   *time.Time `json:"connected_at" db:"connected_at"`
	LastHeartbeat *time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	Fingerprint   string     `json:"fingerprint" db:"fingerprint"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// Operation represents a long-running operation
type Operation struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	ClusterID  uuid.UUID  `json:"cluster_id" db:"cluster_id"`
	Type       string     `json:"type" db:"type"`     // "apply", "exec", "sync", etc.
	Status     string     `json:"status" db:"status"` // "queued", "running", "success", "failed"
	Payload    Payload    `json:"payload" db:"payload"`
	Result     *Payload   `json:"result,omitempty" db:"result"`
	StartedAt  *time.Time `json:"started_at,omitempty" db:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty" db:"finished_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          string    `json:"user_id" db:"user_id"`
	Action          string    `json:"action" db:"action"`
	ResourceType    string    `json:"resource_type" db:"resource_type"`
	ResourceID      string    `json:"resource_id" db:"resource_id"`
	RequestPayload  *Payload  `json:"request_payload,omitempty" db:"request_payload"`
	ResponsePayload *Payload  `json:"response_payload,omitempty" db:"response_payload"`
	IPAddress       string    `json:"ip_address" db:"ip_address"`
	UserAgent       string    `json:"user_agent" db:"user_agent"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// Labels represents cluster labels as JSONB
type Labels map[string]string

// Payload represents arbitrary JSON data
type Payload map[string]interface{}

// ClusterStatus represents the status of a cluster
type ClusterStatus string

const (
	ClusterStatusPending      ClusterStatus = "pending"
	ClusterStatusConnected    ClusterStatus = "connected"
	ClusterStatusDisconnected ClusterStatus = "disconnected"
	ClusterStatusError        ClusterStatus = "error"
)

// OperationStatus represents the status of an operation
type OperationStatus string

const (
	OperationStatusQueued    OperationStatus = "queued"
	OperationStatusRunning   OperationStatus = "running"
	OperationStatusSuccess   OperationStatus = "success"
	OperationStatusFailed    OperationStatus = "failed"
	OperationStatusCancelled OperationStatus = "cancelled"
)

// OperationType represents the type of an operation
type OperationType string

const (
	OperationTypeApply  OperationType = "apply"
	OperationTypeExec   OperationType = "exec"
	OperationTypeSync   OperationType = "sync"
	OperationTypeDelete OperationType = "delete"
)

// ClusterMode represents the connection mode for a cluster
type ClusterMode string

const (
	ClusterModeAgent  ClusterMode = "agent"
	ClusterModeDirect ClusterMode = "direct"
)

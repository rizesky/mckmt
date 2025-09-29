package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=./mocks/mock_repo.go -package=mocks github.com/rizesky/mckmt/internal/repo ClusterRepository,OperationRepository,AuditLogRepository,UserRepository,RoleRepository,Cache,EventBus

// Repository interfaces are defined here because they are shared across multiple services.
// These interfaces represent the data access layer and are consumed by:
// - Cluster service (uses ClusterRepository, OperationRepository, Cache)
// - Operations service (uses OperationRepository, AuditLogRepository)
// - Auth service (uses UserRepository, RoleRepository)
// - Orchestrator (uses OperationRepository)
// - HTTP handlers (indirectly through services)
//
// Following the "define where used" principle, these interfaces are placed in a shared
// location since they are used by multiple consumers across different layers of the application.
// This avoids duplication and ensures consistency across the codebase.

// ClusterRepository defines the interface for cluster operations
type ClusterRepository interface {
	Create(ctx context.Context, cluster *Cluster) error
	GetByID(ctx context.Context, id uuid.UUID) (*Cluster, error)
	GetByName(ctx context.Context, name string) (*Cluster, error)
	List(ctx context.Context, limit, offset int) ([]*Cluster, error)
	Update(ctx context.Context, cluster *Cluster) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error
}

// OperationRepository defines the interface for operation operations
type OperationRepository interface {
	Create(ctx context.Context, operation *Operation) error
	GetByID(ctx context.Context, id uuid.UUID) (*Operation, error)
	ListByCluster(ctx context.Context, clusterID uuid.UUID, limit, offset int) ([]*Operation, error)
	Update(ctx context.Context, operation *Operation) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateResult(ctx context.Context, id uuid.UUID, result Payload) error
	SetStarted(ctx context.Context, id uuid.UUID) error
	SetFinished(ctx context.Context, id uuid.UUID) error
	CancelOperation(ctx context.Context, id uuid.UUID, reason string) error
}

// AuditLogRepository defines the interface for audit log operations
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, userID string, limit, offset int) ([]*AuditLog, error)
	ListByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*AuditLog, error)
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RoleRepository defines the interface for role operations
type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context, limit, offset int) ([]*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// Cache defines the interface for cache operations
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	FlushDB(ctx context.Context) error
	Ping(ctx context.Context) error
	Health(ctx context.Context) error

	// Key generation methods
	ClusterKey(id string) string
	OperationKey(id string) string
	UserKey(id string) string
	SessionKey(token string) string
	ClusterResourcesKey(clusterID, kind, namespace string) string
	ClusterResourceKey(clusterID, kind, namespace, name string) string
	ClusterStatusKey(clusterID string) string
	ClusterMetricsKey(clusterID string) string
}

// EventBus defines the interface for event publishing
type EventBus interface {
	Publish(ctx context.Context, topic string, event interface{}) error
	Subscribe(ctx context.Context, topic string, handler func(event interface{}) error) error
}

// Cluster represents a cluster entity
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

// Operation represents an operation entity
type Operation struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	ClusterID  uuid.UUID  `json:"cluster_id" db:"cluster_id"`
	Type       string     `json:"type" db:"type"`
	Status     string     `json:"status" db:"status"`
	Payload    Payload    `json:"payload" db:"payload"`
	Result     *Payload   `json:"result,omitempty" db:"result"`
	StartedAt  *time.Time `json:"started_at" db:"started_at"`
	FinishedAt *time.Time `json:"finished_at" db:"finished_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// AuditLog represents an audit log entity
type AuditLog struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	Details      Payload   `json:"details" db:"details"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// User represents a user entity
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Role represents a role entity
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Permissions []string  `json:"permissions" db:"permissions"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Payload represents a generic payload
type Payload map[string]interface{}

// Labels represents cluster labels
type Labels map[string]string

// Operation types
const (
	OperationTypeApply  = "apply"
	OperationTypeExec   = "exec"
	OperationTypeSync   = "sync"
	OperationTypeDelete = "delete"
)

// Operation statuses
const (
	OperationStatusPending   = "pending"
	OperationStatusRunning   = "running"
	OperationStatusSuccess   = "success"
	OperationStatusFailed    = "failed"
	OperationStatusCancelled = "cancelled"
)

// Common errors
var (
	ErrNotFound  = fmt.Errorf("not found")
	ErrCacheMiss = fmt.Errorf("cache miss")
)

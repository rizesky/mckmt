package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/user"
)

//go:generate mockgen -destination=./mocks/mock_repo.go -package=mocks github.com/rizesky/mckmt/internal/repo ClusterRepository,OperationRepository,AuditLogRepository,UserRepository,RoleRepository,PermissionRepository,Cache,EventBus

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

	// Additional audit log methods
	ListAll(ctx context.Context, limit, offset int) ([]*AuditLog, error)
	ListByAction(ctx context.Context, action string, limit, offset int) ([]*AuditLog, error)
	ListByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*AuditLog, error)
	Count(ctx context.Context) (int64, error)
	CountByUser(ctx context.Context, userID string) (int64, error)
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(ctx context.Context, user *user.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	List(ctx context.Context, limit, offset int) ([]*user.User, error)
	Update(ctx context.Context, user *user.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RoleRepository defines the interface for role operations
type RoleRepository interface {
	Create(ctx context.Context, role *user.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*user.Role, error)
	GetByName(ctx context.Context, name string) (*user.Role, error)
	List(ctx context.Context, limit, offset int) ([]*user.Role, error)
	Update(ctx context.Context, role *user.Role) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User-role relationship methods
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*user.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error

	// Role-permission relationship methods
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*user.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID, grantedBy *uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
}

// PermissionRepository defines the interface for permission operations
type PermissionRepository interface {
	Create(ctx context.Context, permission *user.Permission) error
	GetByID(ctx context.Context, id uuid.UUID) (*user.Permission, error)
	GetByName(ctx context.Context, name string) (*user.Permission, error)
	GetByResourceAction(ctx context.Context, resource, action string) (*user.Permission, error)
	List(ctx context.Context, limit, offset int) ([]*user.Permission, error)
	ListByResource(ctx context.Context, resource string) ([]*user.Permission, error)
	Update(ctx context.Context, permission *user.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User permission methods
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*user.Permission, error)
	GetUserPermissionsByResource(ctx context.Context, userID uuid.UUID, resource string) ([]*user.Permission, error)
	CheckUserPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
	CheckUserPermissionExact(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
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
	Endpoint             string     `json:"endpoint" db:"endpoint"`
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

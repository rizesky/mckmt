package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Domain errors
var (
	ErrNotFound = errors.New("not found")
)

// AuthSource represents the authentication source
type AuthSource string

const (
	AuthSourcePassword AuthSource = "password" // Local username/password
	AuthSourceOIDC     AuthSource = "oidc"     // OIDC provider
)

// User represents a user entity
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`         // Hidden from JSON, stored in DB
	AuthSource   AuthSource `json:"auth_source" db:"auth_source"` // Where the user came from
	Active       bool       `json:"active" db:"active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	// Relationships (loaded separately)
	Roles         []*Role       `json:"roles,omitempty" db:"-"`
	Permissions   []*Permission `json:"permissions,omitempty" db:"-"`
	OwnedClusters []uuid.UUID   `json:"owned_clusters,omitempty" db:"-"`
}

// Role represents a role entity
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relationships (loaded separately)
	Permissions []*Permission `json:"permissions,omitempty" db:"-"`
}

// Permission represents a permission entity
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents the relationship between users and roles
type UserRole struct {
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	RoleID     uuid.UUID  `json:"role_id" db:"role_id"`
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"`
	AssignedBy *uuid.UUID `json:"assigned_by,omitempty" db:"assigned_by"`
}

// RolePermission represents the relationship between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID  `json:"role_id" db:"role_id"`
	PermissionID uuid.UUID  `json:"permission_id" db:"permission_id"`
	GrantedAt    time.Time  `json:"granted_at" db:"granted_at"`
	GrantedBy    *uuid.UUID `json:"granted_by,omitempty" db:"granted_by"`
}

// UserCluster represents the relationship between users and clusters (for ABAC)
type UserCluster struct {
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	ClusterID   uuid.UUID  `json:"cluster_id" db:"cluster_id"`
	AccessLevel string     `json:"access_level" db:"access_level"`
	GrantedAt   time.Time  `json:"granted_at" db:"granted_at"`
	GrantedBy   *uuid.UUID `json:"granted_by,omitempty" db:"granted_by"`
}

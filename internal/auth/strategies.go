package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/repo"
)

// NoAuthStrategy - No authorization (allow all requests)
type NoAuthStrategy struct {
	logger *zap.Logger
}

func NewNoAuthStrategy(logger *zap.Logger) *NoAuthStrategy {
	return &NoAuthStrategy{logger: logger}
}

func (s *NoAuthStrategy) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	s.logger.Debug("NoAuthStrategy: allowing all requests",
		zap.String("user_id", userID.String()),
		zap.String("resource", resource),
		zap.String("action", action),
	)
	return true, nil
}

func (s *NoAuthStrategy) IsEnabled() bool {
	return true
}

func (s *NoAuthStrategy) GetName() string {
	return "no-auth"
}

// DatabaseRBACStrategy - Uses database for RBAC authorization
type DatabaseRBACStrategy struct {
	permissionRepo repo.PermissionRepository
	logger         *zap.Logger
}

func NewDatabaseRBACStrategy(permissionRepo repo.PermissionRepository, logger *zap.Logger) *DatabaseRBACStrategy {
	return &DatabaseRBACStrategy{
		permissionRepo: permissionRepo,
		logger:         logger,
	}
}

func (s *DatabaseRBACStrategy) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	allowed, err := s.permissionRepo.CheckUserPermission(ctx, userID, resource, action)
	if err != nil {
		s.logger.Error("DatabaseRBACStrategy: failed to check permission",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("resource", resource),
			zap.String("action", action),
		)
		return false, err
	}

	s.logger.Debug("DatabaseRBACStrategy: permission check result",
		zap.Bool("allowed", allowed),
		zap.String("user_id", userID.String()),
		zap.String("resource", resource),
		zap.String("action", action),
	)

	return allowed, nil
}

func (s *DatabaseRBACStrategy) IsEnabled() bool {
	return true
}

func (s *DatabaseRBACStrategy) GetName() string {
	return "database-rbac"
}

// CasbinStrategy - Uses Casbin for advanced authorization
type CasbinStrategy struct {
	enforcer CasbinEnforcer
	logger   *zap.Logger
}

func NewCasbinStrategy(enforcer CasbinEnforcer, logger *zap.Logger) *CasbinStrategy {
	return &CasbinStrategy{
		enforcer: enforcer,
		logger:   logger,
	}
}

func (s *CasbinStrategy) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	// Get user from context
	user, ok := GetUserFromContext(ctx)
	if !ok {
		s.logger.Error("CasbinStrategy: user not found in context")
		return false, ErrUserNotFound
	}

	allowed, err := s.enforcer.Enforce(user.Username, resource, action)
	if err != nil {
		s.logger.Error("CasbinStrategy: failed to enforce permission",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("username", user.Username),
			zap.String("resource", resource),
			zap.String("action", action),
		)
		return false, err
	}

	s.logger.Debug("CasbinStrategy: permission check result",
		zap.Bool("allowed", allowed),
		zap.String("user_id", userID.String()),
		zap.String("username", user.Username),
		zap.String("resource", resource),
		zap.String("action", action),
	)

	return allowed, nil
}

func (s *CasbinStrategy) IsEnabled() bool {
	return s.enforcer != nil
}

func (s *CasbinStrategy) GetName() string {
	return "casbin"
}

// CasbinEnforcer wraps the Casbin enforcer for easier testing and mocking
type CasbinEnforcer interface {
	Enforce(subject, object, action string) (bool, error)
	LoadPolicy() error
}

// Error definitions
var (
	ErrUserNotFound = fmt.Errorf("user not found in context")
)

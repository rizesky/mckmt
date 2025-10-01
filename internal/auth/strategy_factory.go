package auth

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/repo"
)

// StrategyType represents the type of authorization strategy
type StrategyType string

const (
	StrategyNoAuth       StrategyType = "no-auth"
	StrategyDatabaseRBAC StrategyType = "database-rbac"
	StrategyCasbin       StrategyType = "casbin"
)

// StrategyConfig holds configuration for creating authorization strategies
type StrategyConfig struct {
	Type           StrategyType
	PermissionRepo repo.PermissionRepository
	RoleRepo       repo.RoleRepository
	UserRepo       repo.UserRepository
	CasbinEnforcer CasbinEnforcer
	Logger         *zap.Logger
}

// NewAuthorizationStrategy creates an authorization strategy based on the configuration
func NewAuthorizationStrategy(config StrategyConfig) (AuthorizationStrategy, error) {
	switch config.Type {
	case StrategyNoAuth:
		return NewNoAuthStrategy(config.Logger), nil

	case StrategyDatabaseRBAC:
		if config.PermissionRepo == nil {
			return nil, fmt.Errorf("permission repository is required for database RBAC strategy")
		}
		return NewDatabaseRBACStrategy(config.PermissionRepo, config.Logger), nil

	case StrategyCasbin:
		if config.CasbinEnforcer == nil {
			return nil, fmt.Errorf("casbin enforcer is required for casbin strategy")
		}
		return NewCasbinStrategy(config.CasbinEnforcer, config.Logger), nil

	default:
		return nil, fmt.Errorf("unknown authorization strategy: %s", config.Type)
	}
}

// CreateAuthorizationService creates an authorization service with the specified strategy
func CreateAuthorizationService(config StrategyConfig) (*AuthorizationService, error) {
	strategy, err := NewAuthorizationStrategy(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization strategy: %w", err)
	}

	return NewAuthorizationService(strategy, config.Logger), nil
}

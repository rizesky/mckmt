package auth

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthorizationStrategy defines the interface for different authorization strategies
type AuthorizationStrategy interface {
	// CheckPermission checks if a user has permission to perform an action on a resource
	CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)

	// IsEnabled returns whether this strategy is active
	IsEnabled() bool

	// GetName returns the name of this strategy
	GetName() string
}

// AuthorizationService manages different authorization strategies
type AuthorizationService struct {
	strategy AuthorizationStrategy
	logger   *zap.Logger
}

// NewAuthorizationService creates a new authorization service with the specified strategy
func NewAuthorizationService(strategy AuthorizationStrategy, logger *zap.Logger) *AuthorizationService {
	return &AuthorizationService{
		strategy: strategy,
		logger:   logger,
	}
}

// CheckPermission delegates to the current strategy
func (a *AuthorizationService) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	if !a.strategy.IsEnabled() {
		a.logger.Warn("Authorization strategy is disabled, allowing request",
			zap.String("strategy", a.strategy.GetName()),
			zap.String("user_id", userID.String()),
			zap.String("resource", resource),
			zap.String("action", action),
		)
		return true, nil
	}

	return a.strategy.CheckPermission(ctx, userID, resource, action)
}

// GetStrategy returns the current strategy
func (a *AuthorizationService) GetStrategy() AuthorizationStrategy {
	return a.strategy
}

// SetStrategy changes the authorization strategy
func (a *AuthorizationService) SetStrategy(strategy AuthorizationStrategy) {
	a.strategy = strategy
	a.logger.Info("Authorization strategy changed", zap.String("strategy", strategy.GetName()))
}

// IsEnabled returns whether the current strategy is enabled
func (a *AuthorizationService) IsEnabled() bool {
	return a.strategy.IsEnabled()
}

// GetName returns the name of the current strategy
func (a *AuthorizationService) GetName() string {
	return a.strategy.GetName()
}

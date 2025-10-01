package auth

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/user"
)

// RoleMapper handles mapping between OIDC roles and database roles
type RoleMapper struct {
	roleRepo repo.RoleRepository
	logger   *zap.Logger
}

// NewRoleMapper creates a new role mapper
func NewRoleMapper(roleRepo repo.RoleRepository, logger *zap.Logger) *RoleMapper {
	return &RoleMapper{
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// MapOIDCRolesToDatabaseRoles maps OIDC roles to database roles
func (rm *RoleMapper) MapOIDCRolesToDatabaseRoles(ctx context.Context, oidcRoles []string) ([]*user.Role, error) {
	if len(oidcRoles) == 0 {
		// If no OIDC roles provided, return default role
		return rm.getDefaultRole(ctx)
	}

	var databaseRoles []*user.Role

	for _, oidcRole := range oidcRoles {
		// Map OIDC role to database role
		dbRole, err := rm.mapSingleRole(ctx, oidcRole)
		if err != nil {
			rm.logger.Warn("Failed to map OIDC role to database role",
				zap.String("oidc_role", oidcRole),
				zap.Error(err),
			)
			continue
		}

		if dbRole != nil {
			databaseRoles = append(databaseRoles, dbRole)
		}
	}

	// If no roles were mapped successfully, use default role
	if len(databaseRoles) == 0 {
		rm.logger.Warn("No OIDC roles could be mapped, using default role")
		return rm.getDefaultRole(ctx)
	}

	return databaseRoles, nil
}

// mapSingleRole maps a single OIDC role to a database role
func (rm *RoleMapper) mapSingleRole(ctx context.Context, oidcRole string) (*user.Role, error) {

	roleMappings := map[string]string{
		// Common OIDC role patterns
		"admin":         "admin",
		"administrator": "admin",
		"superuser":     "admin",
		"operator":      "operator",
		"ops":           "operator",
		"viewer":        "viewer",
		"readonly":      "viewer",
		"user":          "viewer",
		"member":        "viewer",

		// Group-based mappings
		"mckmt-admin":    "admin",
		"mckmt-operator": "operator",
		"mckmt-viewer":   "viewer",

		// You can add more mappings as needed
	}

	// Check if we have a direct mapping
	if dbRoleName, exists := roleMappings[oidcRole]; exists {
		return rm.getRoleByName(ctx, dbRoleName)
	}

	// Check for partial matches (e.g., "mckmt-admin" -> "admin")
	for oidcPattern, dbRoleName := range roleMappings {
		if contains(oidcRole, oidcPattern) || contains(oidcPattern, oidcRole) {
			return rm.getRoleByName(ctx, dbRoleName)
		}
	}

	// If no mapping found, return nil (will fall back to default role)
	rm.logger.Debug("No mapping found for OIDC role", zap.String("oidc_role", oidcRole))
	return nil, nil
}

// getRoleByName retrieves a role by name from the database
func (rm *RoleMapper) getRoleByName(ctx context.Context, roleName string) (*user.Role, error) {
	roles, err := rm.roleRepo.List(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	for _, role := range roles {
		if role.Name == roleName {
			return role, nil
		}
	}

	return nil, fmt.Errorf("role not found: %s", roleName)
}

// getDefaultRole returns the default role (viewer)
func (rm *RoleMapper) getDefaultRole(ctx context.Context) ([]*user.Role, error) {
	defaultRole, err := rm.getRoleByName(ctx, "viewer")
	if err != nil {
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}

	return []*user.Role{defaultRole}, nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

// containsSubstring is a simple substring check
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

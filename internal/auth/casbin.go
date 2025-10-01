package auth

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/user"
)

// CasbinService provides RBAC/ABAC authorization using Casbin
type CasbinService struct {
	enforcer       *casbin.Enforcer
	enabled        bool
	rbacEnabled    bool
	defaultRole    string
	logger         *zap.Logger
	permissionRepo repo.PermissionRepository
}

// CasbinAdapter implements persist.Adapter for our database
type CasbinAdapter struct {
	roleRepo       repo.RoleRepository
	permissionRepo repo.PermissionRepository
	userRepo       repo.UserRepository
	logger         *zap.Logger
}

// NewCasbinAdapter creates a new Casbin adapter
func NewCasbinAdapter(roleRepo repo.RoleRepository, permissionRepo repo.PermissionRepository, userRepo repo.UserRepository, logger *zap.Logger) *CasbinAdapter {
	return &CasbinAdapter{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		userRepo:       userRepo,
		logger:         logger,
	}
}

// LoadPolicy loads all policy rules from the database
func (a *CasbinAdapter) LoadPolicy(model model.Model) error {
	ctx := context.Background()

	// Load all users first
	users, err := a.userRepo.List(ctx, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	// For each user, load their roles and permissions
	for _, user := range users {
		// Get user roles
		userRoles, err := a.roleRepo.GetUserRoles(ctx, user.ID)
		if err != nil {
			a.logger.Warn("Failed to load user roles", zap.String("user_id", user.ID.String()), zap.Error(err))
			continue
		}

		// For each role, add user-role assignment and role permissions
		for _, role := range userRoles {
			// Add user-role assignment: user, role
			model.AddPolicy("g", "g", []string{user.Username, role.Name})

			// Get permissions for this role
			permissions, err := a.roleRepo.GetRolePermissions(ctx, role.ID)
			if err != nil {
				a.logger.Warn("Failed to load role permissions", zap.String("role", role.Name), zap.Error(err))
				continue
			}

			// Add role-permission assignments: role, resource, action
			for _, permission := range permissions {
				model.AddPolicy("p", "p", []string{role.Name, permission.Resource, permission.Action})
			}
		}
	}

	return nil
}

// SavePolicy saves all policy rules to the database
func (a *CasbinAdapter) SavePolicy(model model.Model) error {
	// For now, we don't implement saving policies back to the database
	// as our permissions are managed through the repository layer
	return nil
}

// AddPolicy adds a policy rule to the database
func (a *CasbinAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	// Not implemented - we manage policies through repositories
	return nil
}

// RemovePolicy removes a policy rule from the database
func (a *CasbinAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	// Not implemented - we manage policies through repositories
	return nil
}

// RemoveFilteredPolicy removes policy rules that match the filter from the database
func (a *CasbinAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	// Not implemented - we manage policies through repositories
	return nil
}

// NewCasbinService creates a new Casbin service
func NewCasbinService(roleRepo repo.RoleRepository, permissionRepo repo.PermissionRepository, userRepo repo.UserRepository, logger *zap.Logger, rbacEnabled bool, casbinEnabled bool, defaultRole string, modelFile string) (*CasbinService, error) {
	if !rbacEnabled {
		// If RBAC is completely disabled, allow all requests
		return &CasbinService{
			enabled:        false,
			rbacEnabled:    false,
			defaultRole:    defaultRole,
			logger:         logger,
			permissionRepo: permissionRepo,
		}, nil
	}

	if !casbinEnabled {
		// If RBAC is enabled but Casbin is disabled, use database RBAC
		return &CasbinService{
			enabled:        false,
			rbacEnabled:    true,
			defaultRole:    defaultRole,
			logger:         logger,
			permissionRepo: permissionRepo,
		}, nil
	}

	// Load RBAC model from file or use default
	var m model.Model
	var err error

	if modelFile != "" {
		// Load model from file
		m, err = model.NewModelFromFile(modelFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load Casbin model from file %s: %w", modelFile, err)
		}
	} else {
		// Use default built-in model
		modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && (p.obj == r.obj || p.obj == "*") && (p.act == r.act || p.act == "*")
`
		m, err = model.NewModelFromString(modelText)
		if err != nil {
			return nil, fmt.Errorf("failed to create Casbin model: %w", err)
		}
	}

	// Create adapter
	adapter := NewCasbinAdapter(roleRepo, permissionRepo, userRepo, logger)

	// Create enforcer
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Load policies
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load Casbin policies: %w", err)
	}

	return &CasbinService{
		enforcer:       enforcer,
		enabled:        true,
		rbacEnabled:    true,
		defaultRole:    defaultRole,
		logger:         logger,
		permissionRepo: permissionRepo,
	}, nil
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (c *CasbinService) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	if !c.rbacEnabled {
		// If RBAC is completely disabled, allow all requests
		return true, nil
	}

	if !c.enabled {
		// If RBAC is enabled but Casbin is disabled, use database permission check
		return c.checkBasicPermission(ctx, userID, resource, action)
	}

	// Get user from context
	authenticatedUser, ok := GetUserFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("user not found in context")
	}

	// Check permission using Casbin
	allowed, err := c.enforcer.Enforce(authenticatedUser.Username, resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to enforce permission: %w", err)
	}

	return allowed, nil
}

// checkBasicPermission provides a fallback permission check when Casbin is disabled
func (c *CasbinService) checkBasicPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	// When Casbin is disabled, use database permission checks
	// This provides the same security as Casbin but with simpler logic
	return c.permissionRepo.CheckUserPermission(ctx, userID, resource, action)
}

// ReloadPolicies reloads all policies from the database
func (c *CasbinService) ReloadPolicies() error {
	if !c.enabled {
		return nil
	}

	return c.enforcer.LoadPolicy()
}

// AddUserRole adds a role to a user
func (c *CasbinService) AddUserRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if !c.enabled {
		// If Casbin is disabled, this would be handled by the repository layer
		return nil
	}

	// Get role name
	role, err := c.getRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Get user
	user, err := c.getUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Add policy: user, role
	_, err = c.enforcer.AddPolicy(user.Username, role.Name, "")
	if err != nil {
		return fmt.Errorf("failed to add user role policy: %w", err)
	}

	return nil
}

// RemoveUserRole removes a role from a user
func (c *CasbinService) RemoveUserRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if !c.enabled {
		// If Casbin is disabled, this would be handled by the repository layer
		return nil
	}

	// Get role name
	role, err := c.getRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Get user
	user, err := c.getUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Remove policy: user, role
	_, err = c.enforcer.RemovePolicy(user.Username, role.Name, "")
	if err != nil {
		return fmt.Errorf("failed to remove user role policy: %w", err)
	}

	return nil
}

// AddRolePermission adds a permission to a role
func (c *CasbinService) AddRolePermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	if !c.enabled {
		// If Casbin is disabled, this would be handled by the repository layer
		return nil
	}

	// Get role name
	role, err := c.getRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Get permission
	permission, err := c.getPermissionByID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}

	// Add policy: role, resource, action
	_, err = c.enforcer.AddPolicy(role.Name, permission.Resource, permission.Action)
	if err != nil {
		return fmt.Errorf("failed to add role permission policy: %w", err)
	}

	return nil
}

// RemoveRolePermission removes a permission from a role
func (c *CasbinService) RemoveRolePermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	if !c.enabled {
		// If Casbin is disabled, this would be handled by the repository layer
		return nil
	}

	// Get role name
	role, err := c.getRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Get permission
	permission, err := c.getPermissionByID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}

	// Remove policy: role, resource, action
	_, err = c.enforcer.RemovePolicy(role.Name, permission.Resource, permission.Action)
	if err != nil {
		return fmt.Errorf("failed to remove role permission policy: %w", err)
	}

	return nil
}

// GetUserPermissions returns all permissions for a user
func (c *CasbinService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*user.Permission, error) {
	if !c.enabled {
		// If Casbin is disabled, fall back to repository
		return nil, fmt.Errorf("Casbin is disabled")
	}

	// Get user
	authenticatedUser, err := c.getUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get all policies for the user
	policies, _ := c.enforcer.GetFilteredPolicy(0, authenticatedUser.Username)

	var permissions []*user.Permission
	for _, policy := range policies {
		if len(policy) >= 3 {
			permission := &user.Permission{
				Resource: policy[1],
				Action:   policy[2],
			}
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// Helper methods (these would need to be implemented with proper repository access)
func (c *CasbinService) getRoleByID(ctx context.Context, roleID uuid.UUID) (*user.Role, error) {
	// This would need access to the role repository
	// For now, return a placeholder
	return &user.Role{ID: roleID, Name: "placeholder"}, nil
}

func (c *CasbinService) getUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	// This would need access to the user repository
	// For now, return a placeholder
	return &user.User{ID: userID, Username: "placeholder"}, nil
}

func (c *CasbinService) getPermissionByID(ctx context.Context, permissionID uuid.UUID) (*user.Permission, error) {
	// This would need access to the permission repository
	// For now, return a placeholder
	return &user.Permission{ID: permissionID, Resource: "placeholder", Action: "placeholder"}, nil
}

// IsEnabled returns whether Casbin is enabled
func (c *CasbinService) IsEnabled() bool {
	return c.enabled
}

// IsRBACEnabled returns whether RBAC is enabled
func (c *CasbinService) IsRBACEnabled() bool {
	return c.rbacEnabled
}

// GetDefaultRole returns the default role for new users
func (c *CasbinService) GetDefaultRole() string {
	return c.defaultRole
}

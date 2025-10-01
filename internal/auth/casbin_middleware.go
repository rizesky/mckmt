package auth

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequirePermission creates middleware that requires a specific permission
func (m *Middleware) RequirePermission(authzService *AuthorizationService, resource, action string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				m.writeErrorResponse(w, http.StatusUnauthorized, "User not found in context")
				return
			}

			// Parse user ID to UUID
			userID, err := uuid.Parse(user.ID)
			if err != nil {
				m.logger.Error("Invalid user ID format",
					zap.Error(err),
					zap.String("user_id", user.ID),
				)
				m.writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
				return
			}

			// Check permission using authorization service
			allowed, err := authzService.CheckPermission(r.Context(), userID, resource, action)
			if err != nil {
				m.logger.Error("Failed to check permission",
					zap.Error(err),
					zap.String("user_id", user.ID),
					zap.String("resource", resource),
					zap.String("action", action),
				)
				m.writeErrorResponse(w, http.StatusInternalServerError, "Failed to check permission")
				return
			}

			if !allowed {
				m.logger.Warn("Permission denied",
					zap.String("user_id", user.ID),
					zap.String("username", user.Username),
					zap.String("resource", resource),
					zap.String("action", action),
				)
				m.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			// Permission granted, proceed to next handler
			next(w, r)
		}
	}
}

// RequireAnyPermission creates middleware that requires any of the specified permissions
func (m *Middleware) RequireAnyPermission(casbinService *CasbinService, permissions ...Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				m.writeErrorResponse(w, http.StatusUnauthorized, "User not found in context")
				return
			}

			// Check if Casbin is enabled
			if !casbinService.IsEnabled() {
				// If Casbin is disabled, allow the request to proceed
				next.ServeHTTP(w, r)
				return
			}

			// Parse user ID to UUID
			userID, err := uuid.Parse(user.ID)
			if err != nil {
				m.logger.Error("Invalid user ID format",
					zap.Error(err),
					zap.String("user_id", user.ID),
				)
				m.writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
				return
			}

			// Check if user has any of the required permissions
			for _, perm := range permissions {
				allowed, err := casbinService.CheckPermission(r.Context(), userID, perm.Resource, perm.Action)
				if err != nil {
					m.logger.Error("Failed to check permission",
						zap.Error(err),
						zap.String("user_id", user.ID),
						zap.String("resource", perm.Resource),
						zap.String("action", perm.Action),
					)
					continue
				}

				if allowed {
					// User has at least one required permission
					next.ServeHTTP(w, r)
					return
				}
			}

			// User doesn't have any of the required permissions
			m.logger.Warn("Permission denied - no matching permissions",
				zap.String("user_id", user.ID),
				zap.String("username", user.Username),
				zap.Int("required_permissions", len(permissions)),
			)
			m.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		}
	}
}

// RequireAllPermissions creates middleware that requires all of the specified permissions
func (m *Middleware) RequireAllPermissions(casbinService *CasbinService, permissions ...Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				m.writeErrorResponse(w, http.StatusUnauthorized, "User not found in context")
				return
			}

			// Check if Casbin is enabled
			if !casbinService.IsEnabled() {
				// If Casbin is disabled, allow the request to proceed
				next.ServeHTTP(w, r)
				return
			}

			// Parse user ID to UUID
			userID, err := uuid.Parse(user.ID)
			if err != nil {
				m.logger.Error("Invalid user ID format",
					zap.Error(err),
					zap.String("user_id", user.ID),
				)
				m.writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
				return
			}

			// Check if user has all required permissions
			for _, perm := range permissions {
				allowed, err := casbinService.CheckPermission(r.Context(), userID, perm.Resource, perm.Action)
				if err != nil {
					m.logger.Error("Failed to check permission",
						zap.Error(err),
						zap.String("user_id", user.ID),
						zap.String("resource", perm.Resource),
						zap.String("action", perm.Action),
					)
					m.writeErrorResponse(w, http.StatusInternalServerError, "Failed to check permission")
					return
				}

				if !allowed {
					m.logger.Warn("Permission denied - missing required permission",
						zap.String("user_id", user.ID),
						zap.String("username", user.Username),
						zap.String("resource", perm.Resource),
						zap.String("action", perm.Action),
					)
					m.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
					return
				}
			}

			// User has all required permissions
			next(w, r)
		}
	}
}

// RequireResourceOwnership creates middleware that requires ownership of a specific resource
func (m *Middleware) RequireResourceOwnership(casbinService *CasbinService, resourceType string, resourceIDParam string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				m.writeErrorResponse(w, http.StatusUnauthorized, "User not found in context")
				return
			}

			// Extract resource ID from URL parameters
			resourceID := r.URL.Query().Get(resourceIDParam)
			if resourceID == "" {
				// Try to get from path parameters (for Chi router)
				resourceID = r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
			}

			if resourceID == "" {
				m.writeErrorResponse(w, http.StatusBadRequest, "Resource ID not provided")
				return
			}

			// Parse resource ID as UUID (for validation)
			_, err := uuid.Parse(resourceID)
			if err != nil {
				m.writeErrorResponse(w, http.StatusBadRequest, "Invalid resource ID")
				return
			}

			// Check if Casbin is enabled
			if !casbinService.IsEnabled() {
				// If Casbin is disabled, allow the request to proceed
				next.ServeHTTP(w, r)
				return
			}

			// Parse user ID to UUID
			userID, err := uuid.Parse(user.ID)
			if err != nil {
				m.logger.Error("Invalid user ID format",
					zap.Error(err),
					zap.String("user_id", user.ID),
				)
				m.writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
				return
			}

			// Check ownership permission
			allowed, err := casbinService.CheckPermission(r.Context(), userID, resourceType, "own")
			if err != nil {
				m.logger.Error("Failed to check ownership permission",
					zap.Error(err),
					zap.String("user_id", user.ID),
					zap.String("resource_type", resourceType),
					zap.String("resource_id", resourceID),
				)
				m.writeErrorResponse(w, http.StatusInternalServerError, "Failed to check ownership")
				return
			}

			if !allowed {
				// Check if user has general access to the resource type
				allowed, err = casbinService.CheckPermission(r.Context(), userID, resourceType, "read")
				if err != nil || !allowed {
					m.logger.Warn("Resource ownership denied",
						zap.String("user_id", user.ID),
						zap.String("username", user.Username),
						zap.String("resource_type", resourceType),
						zap.String("resource_id", resourceID),
					)
					m.writeErrorResponse(w, http.StatusForbidden, "Access denied to resource")
					return
				}
			}

			// Permission granted, proceed to next handler
			next(w, r)
		}
	}
}

// Permission represents a permission requirement
type Permission struct {
	Resource string
	Action   string
}

// NewPermission creates a new permission requirement
func NewPermission(resource, action string) Permission {
	return Permission{
		Resource: resource,
		Action:   action,
	}
}

// Common permission constants
var (
	// Cluster permissions
	ClusterRead   = NewPermission("clusters", "read")
	ClusterWrite  = NewPermission("clusters", "write")
	ClusterDelete = NewPermission("clusters", "delete")
	ClusterManage = NewPermission("clusters", "manage")

	// Operation permissions
	OperationRead   = NewPermission("operations", "read")
	OperationWrite  = NewPermission("operations", "write")
	OperationCancel = NewPermission("operations", "cancel")

	// User permissions
	UserRead   = NewPermission("users", "read")
	UserWrite  = NewPermission("users", "write")
	UserDelete = NewPermission("users", "delete")

	// System permissions
	SystemRead  = NewPermission("system", "read")
	SystemWrite = NewPermission("system", "write")

	// Wildcard permissions
	AllRead  = NewPermission("*", "read")
	AllWrite = NewPermission("*", "write")
	AllAdmin = NewPermission("*", "admin")
)

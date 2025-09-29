package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// ContextKey represents a context key type
type ContextKey string

const (
	// UserContextKey is the key for user information in context
	UserContextKey ContextKey = "user"
	// RequestIDContextKey is the key for request ID in context
	RequestIDContextKey ContextKey = "request_id"
)

// User represents the authenticated user
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	jwtManager *JWTManager
	logger     *zap.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *JWTManager, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// RequireAuth middleware that requires authentication
func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.writeErrorResponse(w, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		token, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			a.writeErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		claims, err := a.jwtManager.ValidateToken(token)
		if err != nil {
			a.logger.Warn("Invalid token", zap.Error(err))
			a.writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		user := &User{
			ID:       claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Roles:    claims.Roles,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware that requires a specific role
func (a *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*User)
			if !ok {
				a.writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			hasRole := false
			for _, userRole := range user.Roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				a.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole middleware that requires any of the specified roles
func (a *AuthMiddleware) RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*User)
			if !ok {
				a.writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range user.Roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				a.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth middleware that optionally authenticates the user
func (a *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		token, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := a.jwtManager.ValidateToken(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user := &User{
			ID:       claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Roles:    claims.Roles,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID middleware that adds a request ID to the context
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", middleware.NextRequestID())
		}

		ctx := context.WithValue(r.Context(), RequestIDContextKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORS middleware for handling cross-origin requests
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeErrorResponse writes an error response
func (a *AuthMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// GetUserFromContext extracts user information from context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserContextKey).(*User)
	return user, ok
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDContextKey).(string)
	return requestID, ok
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(role string) bool {
	for _, userRole := range u.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (u *User) HasAnyRole(roles ...string) bool {
	for _, requiredRole := range roles {
		if u.HasRole(requiredRole) {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.HasRole("admin")
}

// IsOperator checks if the user is an operator
func (u *User) IsOperator() bool {
	return u.HasAnyRole("admin", "operator")
}

// IsViewer checks if the user is a viewer
func (u *User) IsViewer() bool {
	return u.HasAnyRole("admin", "operator", "viewer")
}

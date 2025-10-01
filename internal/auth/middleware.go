package auth

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

// ContextKey represents a context key type
type ContextKey string

const (
	// UserContextKey is the key for user information in context
	UserContextKey ContextKey = "user"
)

// Middleware  provides authentication middleware
type Middleware struct {
	jwtManager *JWTManager
	logger     *zap.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *JWTManager, logger *zap.Logger) *Middleware {
	return &Middleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// RequireAuth middleware that requires authentication
func (a *Middleware) RequireAuth(next http.Handler) http.Handler {
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
			a.logger.Warn("Invalid token", zap.Error(err), zap.String("token", token[:50]+"..."))
			a.writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		user := &AuthenticatedUser{
			ID:       claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Roles:    claims.Roles,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// writeErrorResponse writes an error response
func (a *Middleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// GetUserFromContext extracts user information from context
func GetUserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	user, ok := ctx.Value(UserContextKey).(*AuthenticatedUser)
	return user, ok
}

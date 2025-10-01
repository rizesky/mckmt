package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/api"
	"github.com/rizesky/mckmt/internal/auth"
)

// AuthMethodsResponse represents the response for available auth methods
type AuthMethodsResponse struct {
	Methods []auth.AuthMethod `json:"methods"`
}

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *auth.Service
	logger      *zap.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *auth.Service, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Note: With Go 1.22 enhanced routing, these methods are no longer needed
// as the routing is handled directly in the router using pattern matching.

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login credentials"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get client information
	ipAddress := h.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Perform login
	response, err := h.authService.Login(r.Context(), &req, ipAddress, userAgent)
	if err != nil {
		h.logger.Warn("Login failed", zap.String("username", req.Username), zap.Error(err))

		// Return more specific error messages for better user experience
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "invalid credentials") ||
			strings.Contains(errorMsg, "user not found") ||
			strings.Contains(errorMsg, "invalid password") {
			h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		} else if strings.Contains(errorMsg, "account disabled") ||
			strings.Contains(errorMsg, "account locked") {
			h.writeErrorResponse(w, http.StatusForbidden, errorMsg)
		} else {
			h.writeErrorResponse(w, http.StatusUnauthorized, "Login failed")
		}
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Register handles user registration
// @Summary Register new user
// @Description Create a new user account
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "Registration data"
// @Success 201 {object} auth.RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validateRegisterRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get client information
	ipAddress := h.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Perform registration
	response, err := h.authService.Register(r.Context(), &req, ipAddress, userAgent)
	if err != nil {
		h.logger.Warn("Registration failed", zap.String("username", req.Username), zap.Error(err))

		// Check for specific error types and return appropriate status codes
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "already exists") {
			h.writeErrorResponse(w, http.StatusConflict, errorMsg)
		} else if strings.Contains(errorMsg, "password validation failed") ||
			strings.Contains(errorMsg, "password must contain") ||
			strings.Contains(errorMsg, "username") ||
			strings.Contains(errorMsg, "email") {
			// Password validation and other client-side validation errors
			h.writeErrorResponse(w, http.StatusBadRequest, errorMsg)
		} else {
			// Other errors (database, etc.)
			h.writeErrorResponse(w, http.StatusInternalServerError, "Registration failed")
		}
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// RefreshToken handles token refresh
// @Summary Refresh JWT token
// @Description Generate new access and refresh tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body auth.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} auth.RefreshTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req auth.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get client information
	ipAddress := h.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Perform token refresh
	response, err := h.authService.RefreshToken(r.Context(), &req, ipAddress, userAgent)
	if err != nil {
		h.logger.Warn("Token refresh failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Logout handles user logout
// @Summary Logout user
// @Description Logout user and invalidate session
// @Tags authentication
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get client information
	ipAddress := h.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Perform logout
	err := h.authService.Logout(r.Context(), user.ID, ipAddress, userAgent)
	if err != nil {
		h.logger.Error("Logout failed", zap.String("user_id", user.ID), zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Logout failed")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, SuccessResponse{Message: "Logged out successfully"})
}

// ChangePassword handles password change
// @Summary Change user password
// @Description Change the authenticated user's password
// @Tags authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body auth.ChangePasswordRequest true "Password change data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req auth.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validateChangePasswordRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get client information
	ipAddress := h.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Perform password change
	err := h.authService.ChangePassword(r.Context(), user.ID, &req, ipAddress, userAgent)
	if err != nil {
		h.logger.Warn("Password change failed", zap.String("user_id", user.ID), zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, SuccessResponse{Message: "Password changed successfully"})
}

// GetProfile returns the authenticated user's profile
// @Summary Get user profile
// @Description Get the authenticated user's profile information
// @Tags authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserDTO
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get full profile from service
	profile, err := h.authService.GetUserProfile(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.String("user_id", user.ID), zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get profile")
		return
	}

	// Convert domain user to DTO
	userDTO := ToUserDTO(profile)
	h.writeJSONResponse(w, http.StatusOK, userDTO)
}

// GetAuthMethods returns available authentication methods
// @Summary Get available auth methods
// @Description Returns list of available authentication methods
// @Tags authentication
// @Produce json
// @Success 200 {object} AuthMethodsResponse
// @Router /auth/methods [get]
func (h *AuthHandler) GetAuthMethods(w http.ResponseWriter, r *http.Request) {
	methods := h.authService.GetAvailableAuthMethods()
	h.writeJSONResponse(w, http.StatusOK, AuthMethodsResponse{Methods: methods})
}

// OIDCLogin initiates OIDC login flow
// @Summary Initiate OIDC login
// @Description Redirects to OIDC provider for authentication
// @Tags authentication
// @Produce json
// @Success 302 {string} string "Redirect to OIDC provider"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/oidc/login [get]
func (h *AuthHandler) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	// Get client information
	ipAddress := h.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Initiate OIDC login
	authURL, err := h.authService.OIDCInitiateLogin(r.Context(), ipAddress, userAgent)
	if err != nil {
		h.logger.Error("OIDC login initiation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to initiate OIDC login")
		return
	}

	// Redirect to OIDC provider
	http.Redirect(w, r, authURL, http.StatusFound)
}

// OIDCCallback handles OIDC callback
// @Summary Handle OIDC callback
// @Description Processes OIDC callback and returns JWT tokens
// @Tags authentication
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/oidc/callback [get]
func (h *AuthHandler) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Authorization code is required")
		return
	}

	// Get client information
	ipAddress := h.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Handle OIDC callback
	response, err := h.authService.OIDCCallback(r.Context(), code, state, ipAddress, userAgent)
	if err != nil {
		h.logger.Error("OIDC callback failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "OIDC authentication failed")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// OIDCLogout handles OIDC logout
// @Summary OIDC logout
// @Description Logs out user and redirects to OIDC provider logout
// @Tags authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/oidc/logout [post]
func (h *AuthHandler) OIDCLogout(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get client information
	ipAddress := h.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Perform OIDC logout
	logoutURL, err := h.authService.OIDCLogout(r.Context(), user.ID, ipAddress, userAgent)
	if err != nil {
		h.logger.Error("OIDC logout failed", zap.String("user_id", user.ID), zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "OIDC logout failed")
		return
	}

	// Return logout URL for client to redirect
	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message":    "Logged out successfully",
		"logout_url": logoutURL,
	})
}

// validateRegisterRequest validates registration request
func (h *AuthHandler) validateRegisterRequest(req *auth.RegisterRequest) error {
	if req.Username == "" {
		return &api.ValidationError{Field: "username", Message: "username is required"}
	}
	if len(req.Username) < 3 {
		return &api.ValidationError{Field: "username", Message: "username must be at least 3 characters long"}
	}
	if len(req.Username) > 50 {
		return &api.ValidationError{Field: "username", Message: "username must be no more than 50 characters long"}
	}
	if req.Email == "" {
		return &api.ValidationError{Field: "email", Message: "email is required"}
	}
	if req.Password == "" {
		return &api.ValidationError{Field: "password", Message: "password is required"}
	}
	return nil
}

// validateChangePasswordRequest validates password change request
func (h *AuthHandler) validateChangePasswordRequest(req *auth.ChangePasswordRequest) error {
	if req.CurrentPassword == "" {
		return &api.ValidationError{Field: "current_password", Message: "current password is required"}
	}
	if req.NewPassword == "" {
		return &api.ValidationError{Field: "new_password", Message: "new password is required"}
	}
	if req.CurrentPassword == req.NewPassword {
		return &api.ValidationError{Field: "new_password", Message: "new password must be different from current password"}
	}
	return nil
}

// getClientIP extracts the client IP address from the request
func (h *AuthHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// writeJSONResponse writes a JSON response
func (h *AuthHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

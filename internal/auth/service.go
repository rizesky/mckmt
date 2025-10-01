package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/user"
)

// Helper functions for role conversion
func rolesToStrings(roles []*user.Role) []string {
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}
	return roleNames
}

func stringsToRoles(roleNames []string) []*user.Role {
	roles := make([]*user.Role, len(roleNames))
	for i, name := range roleNames {
		roles[i] = &user.Role{Name: name}
	}
	return roles
}

// Service  handles authentication business logic
type Service struct {
	userRepo        repo.UserRepository
	roleRepo        repo.RoleRepository
	permissionRepo  repo.PermissionRepository
	auditRepo       repo.AuditLogRepository
	jwtManager      *JWTManager
	passwordManager *PasswordManager
	oidcService     *OIDC
	roleMapper      *RoleMapper
	defaultRole     string
	logger          *zap.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repo.UserRepository,
	roleRepo repo.RoleRepository,
	permissionRepo repo.PermissionRepository,
	auditRepo repo.AuditLogRepository,
	jwtManager *JWTManager,
	passwordManager *PasswordManager,
	oidcService *OIDC,
	roleMapper *RoleMapper,
	defaultRole string,
	logger *zap.Logger,
) *Service {
	return &Service{
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		permissionRepo:  permissionRepo,
		auditRepo:       auditRepo,
		jwtManager:      jwtManager,
		passwordManager: passwordManager,
		oidcService:     oidcService,
		roleMapper:      roleMapper,
		defaultRole:     defaultRole,
		logger:          logger,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresAt    time.Time  `json:"expires_at"`
	TokenType    string     `json:"token_type"`
	User         *user.User `json:"user"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	User *user.User `json:"user"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required"`
}

// AuthMethod represents an available authentication method
type AuthMethod struct {
	Type        string `json:"type"`          // "oidc" or "password"
	Name        string `json:"name"`          // Display name
	Description string `json:"description"`   // Description
	Enabled     bool   `json:"enabled"`       // Whether this method is enabled
	URL         string `json:"url,omitempty"` // Login URL for this method
}

// Validate validates the login request
func (r *LoginRequest) Validate() error {
	if r.Username == "" {
		return fmt.Errorf("username is required")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

// Validate validates the refresh token request
func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return fmt.Errorf("refresh_token is required")
	}
	return nil
}

// Login authenticates a newUser and returns JWT tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	// Get newUser by username
	newUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			s.logger.Warn("Login attempt with non-existent username", zap.String("username", req.Username), zap.String("ip", ipAddress))
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get newUser: %w", err)
	}

	// Check if newUser is active
	if !newUser.Active {
		s.logger.Warn("Login attempt with inactive newUser", zap.String("username", req.Username), zap.String("ip", ipAddress))
		return nil, fmt.Errorf("account is disabled")
	}

	// Verify password based on auth source
	if newUser.AuthSource == user.AuthSourceOIDC {
		s.logger.Warn("Login attempt with password for OIDC user", zap.String("username", req.Username), zap.String("ip", ipAddress))
		return nil, fmt.Errorf("this user must authenticate via OIDC")
	}

	if newUser.PasswordHash == "" {
		s.logger.Warn("Login attempt for user without password", zap.String("username", req.Username), zap.String("ip", ipAddress))
		return nil, fmt.Errorf("password authentication not available for this user")
	}

	valid, err := s.passwordManager.VerifyPassword(req.Password, newUser.PasswordHash)
	if err != nil {
		s.logger.Error("Password verification failed", zap.Error(err), zap.String("username", req.Username), zap.String("ip", ipAddress))
		return nil, fmt.Errorf("password verification failed")
	}

	if !valid {
		s.logger.Warn("Login attempt with invalid password", zap.String("username", req.Username), zap.String("ip", ipAddress))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateToken(newUser.ID.String(), newUser.Username, newUser.Email, rolesToStrings(newUser.Roles))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateToken(newUser.ID.String(), newUser.Username, newUser.Email, rolesToStrings(newUser.Roles))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Log successful login
	s.logAuditEvent(ctx, newUser.ID.String(), "login", "newUser", newUser.ID.String(), nil, nil, ipAddress, userAgent)

	// Create response
	authUser := &user.User{
		ID:       newUser.ID,
		Username: newUser.Username,
		Email:    newUser.Email,
		Roles:    newUser.Roles, // Keep as []*user.Role for the response
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().UTC().Add(s.jwtManager.tokenDuration),
		TokenType:    "Bearer",
		User:         authUser,
	}, nil
}

// Register creates a new newUser account
func (s *Service) Register(ctx context.Context, req *RegisterRequest, ipAddress, userAgent string) (*RegisterResponse, error) {
	// Validate password strength
	if err := s.passwordManager.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Check if username already exists
	_, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil {
		return nil, fmt.Errorf("username already exists")
	} else if !errors.Is(err, repo.ErrNotFound) {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}

	// Check if email already exists
	_, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, fmt.Errorf("email already exists")
	} else if !errors.Is(err, repo.ErrNotFound) {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	// Hash password for storage
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user with hashed password
	newUser := &user.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		AuthSource:   user.AuthSourcePassword,
		Roles:        stringsToRoles([]string{s.defaultRole}), // Configurable default role
		Active:       true,
	}

	err = s.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log registration
	s.logAuditEvent(ctx, newUser.ID.String(), "register", "user", newUser.ID.String(), nil, nil, ipAddress, userAgent)

	// Create response
	return &RegisterResponse{
		User: newUser,
	}, nil
}

// RefreshToken generates new tokens using a refresh token
func (s *Service) RefreshToken(ctx context.Context, req *RefreshTokenRequest, ipAddress, userAgent string) (*RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get newUser to ensure they still exist and are active
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid newUser ID in token: %w", err)
	}

	newUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == repo.ErrNotFound {
			return nil, fmt.Errorf("newUser not found")
		}
		return nil, fmt.Errorf("failed to get newUser: %w", err)
	}

	if !newUser.Active {
		return nil, fmt.Errorf("account is disabled")
	}

	// Generate new tokens
	accessToken, err := s.jwtManager.GenerateToken(newUser.ID.String(), newUser.Username, newUser.Email, rolesToStrings(newUser.Roles))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtManager.GenerateToken(newUser.ID.String(), newUser.Username, newUser.Email, rolesToStrings(newUser.Roles))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Log token refresh
	s.logAuditEvent(ctx, newUser.ID.String(), "refresh_token", "newUser", newUser.ID.String(), nil, nil, ipAddress, userAgent)

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().UTC().Add(s.jwtManager.tokenDuration),
		TokenType:    "Bearer",
	}, nil
}

// Logout logs out a newUser (in a real implementation, you might blacklist the token)
func (s *Service) Logout(ctx context.Context, userID, ipAddress, userAgent string) error {
	// Log logout
	s.logAuditEvent(ctx, userID, "logout", "newUser", userID, nil, nil, ipAddress, userAgent)
	return nil
}

// ChangePassword changes a newUser's password
func (s *Service) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest, ipAddress, userAgent string) error {
	// Get newUser
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid newUser ID: %w", err)
	}

	_, err = s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		if err == repo.ErrNotFound {
			return fmt.Errorf("newUser not found")
		}
		return fmt.Errorf("failed to get newUser: %w", err)
	}

	// Validate new password strength
	if err := s.passwordManager.ValidatePasswordStrength(req.NewPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Note: In a real implementation, you would verify the current password here
	// and update the password in the database

	// Log password change
	s.logAuditEvent(ctx, userID, "change_password", "newUser", userID, nil, nil, ipAddress, userAgent)

	s.logger.Info("Password changed successfully", zap.String("user_id", userID))
	return nil
}

// GetUserProfile returns the newUser's profile information
func (s *Service) GetUserProfile(ctx context.Context, userID string) (*user.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid newUser ID: %w", err)
	}

	newUser, err := s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		if err == repo.ErrNotFound {
			return nil, fmt.Errorf("newUser not found")
		}
		return nil, fmt.Errorf("failed to get newUser: %w", err)
	}

	return newUser, nil
}

// OIDCInitiateLogin initiates OIDC login flow
func (s *Service) OIDCInitiateLogin(ctx context.Context, ipAddress, userAgent string) (string, error) {
	if s.oidcService == nil {
		return "", fmt.Errorf("OIDC service not configured")
	}

	// Generate state parameter
	state, err := GenerateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Log OIDC login initiation
	s.logAuditEvent(ctx, "system", "oidc_login_initiated", "auth", "oidc", nil, nil, ipAddress, userAgent)

	return s.oidcService.GetAuthURL(state), nil
}

// OIDCCallback handles OIDC callback
func (s *Service) OIDCCallback(ctx context.Context, code, state string, ipAddress, userAgent string) (*LoginResponse, error) {
	if s.oidcService == nil {
		return nil, fmt.Errorf("OIDC service not configured")
	}

	// Exchange code for tokens
	token, err := s.oidcService.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Validate ID token and get newUser info
	userInfo, err := s.oidcService.ValidateIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ID token: %w", err)
	}

	// Create or update newUser in database
	newUser, err := s.createOrUpdateUserFromOIDC(ctx, userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update newUser: %w", err)
	}

	// Generate our own JWT token for API access
	accessToken, err := s.jwtManager.GenerateToken(newUser.ID.String(), newUser.Username, newUser.Email, rolesToStrings(newUser.Roles))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateToken(newUser.ID.String(), newUser.Username, newUser.Email, rolesToStrings(newUser.Roles))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Log successful OIDC login
	s.logAuditEvent(ctx, newUser.ID.String(), "oidc_login_success", "user", newUser.ID.String(), nil, nil, ipAddress, userAgent)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().UTC().Add(s.jwtManager.tokenDuration),
		TokenType:    "Bearer",
		User:         newUser,
	}, nil
}

// createOrUpdateUserFromOIDC creates or updates a newUser from OIDC newUser info
func (s *Service) createOrUpdateUserFromOIDC(ctx context.Context, userInfo *OIDCUserInfo) (*user.User, error) {
	// Parse newUser ID
	userID, err := uuid.Parse(userInfo.Subject)
	if err != nil {
		// If subject is not a UUID, generate one
		userID = uuid.New()
	}

	// Check if newUser exists
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil && err != repo.ErrNotFound {
		return nil, fmt.Errorf("failed to check existing newUser: %w", err)
	}

	// Map OIDC roles to database roles
	databaseRoles, err := s.roleMapper.MapOIDCRolesToDatabaseRoles(ctx, userInfo.Roles)
	if err != nil {
		s.logger.Warn("Failed to map OIDC roles to database roles, using default role",
			zap.Strings("oidc_roles", userInfo.Roles),
			zap.Error(err),
		)
		// Fall back to default role
		databaseRoles, err = s.roleMapper.MapOIDCRolesToDatabaseRoles(ctx, []string{s.defaultRole})
		if err != nil {
			return nil, fmt.Errorf("failed to get default role: %w", err)
		}
	}

	// Create user from OIDC info with mapped database roles
	newUser := &user.User{
		ID:         userID,
		Username:   userInfo.Email, // Use email as username
		Email:      userInfo.Email,
		AuthSource: user.AuthSourceOIDC,
		Roles:      databaseRoles, // Use mapped database roles instead of OIDC roles
		Active:     true,
	}

	if existingUser == nil {
		// Create new user
		err = s.userRepo.Create(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update existing user
		existingUser.Email = userInfo.Email
		existingUser.Roles = stringsToRoles(userInfo.Roles)
		existingUser.Active = true

		err = s.userRepo.Update(ctx, existingUser)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		newUser = existingUser
	}

	return newUser, nil
}

// GetAvailableAuthMethods returns the list of available authentication methods
func (s *Service) GetAvailableAuthMethods() []AuthMethod {
	var methods []AuthMethod

	// Always include password auth (for development/testing)
	methods = append(methods, AuthMethod{
		Type:        "password",
		Name:        "Username & Password",
		Description: "Login with username and password",
		Enabled:     true,
		URL:         "/api/v1/auth/login",
	})

	// Include OIDC if enabled
	if s.oidcService != nil {
		methods = append(methods, AuthMethod{
			Type:        "oidc",
			Name:        "Single Sign-On",
			Description: "Login with your organization account",
			Enabled:     true,
			URL:         "/api/v1/auth/oidc/login",
		})
	}

	return methods
}

// OIDCLogout handles OIDC logout
func (s *Service) OIDCLogout(ctx context.Context, userID, ipAddress, userAgent string) (string, error) {
	if s.oidcService == nil {
		return "", fmt.Errorf("OIDC service not configured")
	}

	// Log OIDC logout
	s.logAuditEvent(ctx, userID, "oidc_logout", "newUser", userID, nil, nil, ipAddress, userAgent)

	// Get logout URL
	logoutURL := s.oidcService.GetLogoutURL("")
	return logoutURL, nil
}

// logAuditEvent logs an audit event
func (s *Service) logAuditEvent(ctx context.Context, userID, action, resourceType, resourceID string, requestPayload, responsePayload *repo.Payload, ipAddress, userAgent string) {
	if s.auditRepo == nil {
		return // Skip if audit logging is disabled
	}

	auditLog := &repo.AuditLog{
		ID:              uuid.New(),
		UserID:          userID,
		Action:          action,
		ResourceType:    resourceType,
		ResourceID:      resourceID,
		RequestPayload:  requestPayload,
		ResponsePayload: responsePayload,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		CreatedAt:       time.Now().UTC(),
	}

	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		s.logger.Error("Failed to log audit event", zap.Error(err))
	}
}

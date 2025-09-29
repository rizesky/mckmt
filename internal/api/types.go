package api

import "time"

// UpdateClusterRequest represents a request to update a cluster
type UpdateClusterRequest struct {
	Name        string            `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description string            `json:"description,omitempty" validate:"omitempty,max=500"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// Validate validates the update cluster request
func (r *UpdateClusterRequest) Validate() error {
	if r.Name != "" && len(r.Name) > 100 {
		return &ValidationError{Field: "name", Message: "name must be less than 100 characters"}
	}
	if r.Description != "" && len(r.Description) > 500 {
		return &ValidationError{Field: "description", Message: "description must be less than 500 characters"}
	}
	return nil
}

// ApplyManifestsRequest represents a request to apply manifests
type ApplyManifestsRequest struct {
	Manifests []string `json:"manifests" validate:"required,min=1"`
	Namespace string   `json:"namespace,omitempty"`
	DryRun    bool     `json:"dry_run,omitempty"`
}

// Validate validates the apply manifests request
func (r *ApplyManifestsRequest) Validate() error {
	if len(r.Manifests) == 0 {
		return &ValidationError{Field: "manifests", Message: "at least one manifest is required"}
	}
	return nil
}

// ExecCommandRequest represents a request to execute a command
type ExecCommandRequest struct {
	Command   []string `json:"command" validate:"required,min=1"`
	Namespace string   `json:"namespace,omitempty"`
	Pod       string   `json:"pod" validate:"required"`
	Container string   `json:"container,omitempty"`
}

// Validate validates the exec command request
func (r *ExecCommandRequest) Validate() error {
	if len(r.Command) == 0 {
		return &ValidationError{Field: "command", Message: "command is required"}
	}
	if r.Pod == "" {
		return &ValidationError{Field: "pod", Message: "pod is required"}
	}
	return nil
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Validate validates the login request
func (r *LoginRequest) Validate() error {
	if r.Username == "" {
		return &ValidationError{Field: "username", Message: "username is required"}
	}
	if r.Password == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}
	return nil
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Validate validates the refresh token request
func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return &ValidationError{Field: "refresh_token", Message: "refresh_token is required"}
	}
	return nil
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// ClusterResponse represents a cluster response
type ClusterResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels"`
	Status      string            `json:"status"`
	LastSeenAt  *time.Time        `json:"last_seen_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// OperationResponse represents an operation response
type OperationResponse struct {
	ID         string                 `json:"id"`
	ClusterID  string                 `json:"cluster_id"`
	Type       string                 `json:"type"`
	Status     string                 `json:"status"`
	Payload    map[string]interface{} `json:"payload"`
	Result     map[string]interface{} `json:"result,omitempty"`
	StartedAt  *time.Time             `json:"started_at,omitempty"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

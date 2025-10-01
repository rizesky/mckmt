package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	userID := "user-123"
	username := "testuser"
	email := "test@example.com"
	roles := []string{"admin", "user"}

	token, err := jwtManager.GenerateToken(userID, username, email, roles)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_ValidateToken(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	userID := "user-123"
	username := "testuser"
	email := "test@example.com"
	roles := []string{"admin", "user"}

	// Generate token
	token, err := jwtManager.GenerateToken(userID, username, email, roles)
	require.NoError(t, err)

	// Validate token
	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, roles, claims.Roles)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
}

func TestJWTManager_ValidateToken_InvalidToken(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	// Test invalid token
	_, err := jwtManager.ValidateToken("invalid-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")

	// Test empty token
	_, err = jwtManager.ValidateToken("")
	require.Error(t, err)

	// Test malformed token
	_, err = jwtManager.ValidateToken("invalid.token.format")
	require.Error(t, err)
}

func TestJWTManager_ValidateToken_WrongSecret(t *testing.T) {
	secretKey1 := "test-secret-key-1"
	secretKey2 := "test-secret-key-2"
	tokenDuration := 1 * time.Hour

	jwtManager1 := NewJWTManager(secretKey1, tokenDuration)
	jwtManager2 := NewJWTManager(secretKey2, tokenDuration)

	userID := "user-123"
	username := "testuser"
	email := "test@example.com"
	roles := []string{"admin"}

	// Generate token with first manager
	token, err := jwtManager1.GenerateToken(userID, username, email, roles)
	require.NoError(t, err)

	// Try to validate with second manager (different secret)
	_, err = jwtManager2.ValidateToken(token)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestJWTManager_ValidateToken_ExpiredToken(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := -1 * time.Hour // Negative duration = expired
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	userID := "user-123"
	username := "testuser"
	email := "test@example.com"
	roles := []string{"admin"}

	// Generate expired token
	token, err := jwtManager.GenerateToken(userID, username, email, roles)
	require.NoError(t, err)

	// Validate expired token
	_, err = jwtManager.ValidateToken(token)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestJWTManager_RefreshToken(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	userID := "user-123"
	username := "testuser"
	email := "test@example.com"
	roles := []string{"admin", "user"}

	// Generate original token
	originalToken, err := jwtManager.GenerateToken(userID, username, email, roles)
	require.NoError(t, err)

	// Refresh token
	refreshToken, err := jwtManager.RefreshToken(originalToken)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, originalToken, refreshToken)

	// Validate refresh token
	claims, err := jwtManager.ValidateToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, roles, claims.Roles)
}

func TestJWTManager_RefreshToken_InvalidToken(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	// Test refresh with invalid token
	_, err := jwtManager.RefreshToken("invalid-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate token")
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		expectError bool
		expected    string
	}{
		{
			name:        "valid bearer token",
			header:      "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			expectError: false,
			expected:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		},
		{
			name:        "empty header",
			header:      "",
			expectError: true,
		},
		{
			name:        "no bearer prefix",
			header:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			expectError: true,
		},
		{
			name:        "wrong prefix",
			header:      "Basic eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			expectError: true,
		},
		{
			name:        "bearer with space",
			header:      "Bearer  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			expectError: false,
			expected:    " eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractTokenFromHeader(tt.header)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, token)
			}
		})
	}
}

func TestJWTManager_TokenExpiration(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Second // Use 1 second duration
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	userID := "user-123"
	username := "testuser"
	email := "test@example.com"
	roles := []string{"admin"}

	// Generate token
	token, err := jwtManager.GenerateToken(userID, username, email, roles)
	require.NoError(t, err)

	// Validate immediately (should work)
	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	// Wait for token to expire
	time.Sleep(1100 * time.Millisecond) // Wait slightly longer than token duration

	// Validate after expiration (should fail)
	_, err = jwtManager.ValidateToken(token)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestJWTManager_ClaimsStructure(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	userID := "user-123"
	username := "testuser"
	email := "test@example.com"
	roles := []string{"admin", "user"}

	// Generate token
	token, err := jwtManager.GenerateToken(userID, username, email, roles)
	require.NoError(t, err)

	// Validate token and check claims structure
	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)

	// Check custom claims
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, roles, claims.Roles)

	// Check registered claims
	assert.NotEmpty(t, claims.ID)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.NotBefore)
	assert.True(t, claims.ExpiresAt.After(claims.IssuedAt.Time))
	assert.True(t, claims.NotBefore.Before(claims.ExpiresAt.Time) || claims.NotBefore.Equal(claims.ExpiresAt.Time))
}

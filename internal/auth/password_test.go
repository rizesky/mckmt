package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordManager_HashPassword(t *testing.T) {
	pm := NewPasswordManager(nil)

	password := "TestPassword123!"
	hash, err := pm.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")
}

func TestPasswordManager_VerifyPassword(t *testing.T) {
	pm := NewPasswordManager(nil)

	password := "TestPassword123!"
	hash, err := pm.HashPassword(password)
	require.NoError(t, err)

	// Test correct password
	valid, err := pm.VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)

	// Test incorrect password
	valid, err = pm.VerifyPassword("WrongPassword", hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestPasswordManager_VerifyPassword_InvalidHash(t *testing.T) {
	pm := NewPasswordManager(nil)

	// Test invalid hash format
	valid, err := pm.VerifyPassword("password", "invalid-hash")
	require.Error(t, err)
	assert.False(t, valid)

	// Test empty hash
	valid, err = pm.VerifyPassword("password", "")
	require.Error(t, err)
	assert.False(t, valid)
}

func TestPasswordManager_ValidatePasswordStrength(t *testing.T) {
	pm := NewPasswordManager(nil)

	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid password",
			password:    "TestPassword123!",
			expectError: false,
		},
		{
			name:        "too short",
			password:    "Test1!",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
		{
			name:        "too long",
			password:    "ThisIsAVeryLongPasswordThatExceedsTheMaximumLengthLimitOf128CharactersAndShouldFailValidationBecauseItIsTooLongForTheSystemToHandleProperly123!",
			expectError: true,
			errorMsg:    "password must be no more than 128 characters long",
		},
		{
			name:        "no lowercase",
			password:    "TESTPASSWORD123!",
			expectError: true,
			errorMsg:    "password must contain at least one lowercase letter",
		},
		{
			name:        "no uppercase",
			password:    "testpassword123!",
			expectError: true,
			errorMsg:    "password must contain at least one uppercase letter",
		},
		{
			name:        "no digit",
			password:    "TestPassword!",
			expectError: true,
			errorMsg:    "password must contain at least one digit",
		},
		{
			name:        "no special character",
			password:    "TestPassword123",
			expectError: true,
			errorMsg:    "password must contain at least one special character",
		},
		{
			name:        "empty password",
			password:    "",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidatePasswordStrength(tt.password)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPasswordManager_CustomConfig(t *testing.T) {
	config := &PasswordConfig{
		Memory:      32 * 1024, // 32 MB
		Iterations:  2,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	pm := NewPasswordManager(config)
	password := "TestPassword123!"

	hash, err := pm.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	valid, err := pm.VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestPasswordManager_DifferentPasswordsProduceDifferentHashes(t *testing.T) {
	pm := NewPasswordManager(nil)

	password1 := "TestPassword123!"
	password2 := "AnotherPassword456@"

	hash1, err := pm.HashPassword(password1)
	require.NoError(t, err)

	hash2, err := pm.HashPassword(password2)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestPasswordManager_SamePasswordProducesDifferentHashes(t *testing.T) {
	pm := NewPasswordManager(nil)

	password := "TestPassword123!"

	hash1, err := pm.HashPassword(password)
	require.NoError(t, err)

	hash2, err := pm.HashPassword(password)
	require.NoError(t, err)

	// Due to random salt, same password should produce different hashes
	assert.NotEqual(t, hash1, hash2)

	// But both should verify correctly
	valid1, err := pm.VerifyPassword(password, hash1)
	require.NoError(t, err)
	assert.True(t, valid1)

	valid2, err := pm.VerifyPassword(password, hash2)
	require.NoError(t, err)
	assert.True(t, valid2)
}

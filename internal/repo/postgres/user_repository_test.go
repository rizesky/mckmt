package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	userdomain "github.com/rizesky/mckmt/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *Database {
	t.Helper()
	// For now, return nil - these tests need a proper test database setup
	// In a real implementation, this would create a test database
	t.Skip("Test database setup not implemented - requires proper test infrastructure")
	return nil
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(t *testing.T, db *Database) {
	t.Helper()
	// Cleanup logic would go here
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	testUser := &userdomain.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []*userdomain.Role{{Name: "admin"}, {Name: "user"}},
		Active:   true,
	}

	err := repo.Create(ctx, testUser)
	require.NoError(t, err)
	assert.NotZero(t, testUser.CreatedAt)
	assert.NotZero(t, testUser.UpdatedAt)

	// Verify user was created
	created, err := repo.GetByID(ctx, testUser.ID)
	require.NoError(t, err)
	assert.Equal(t, testUser.ID, created.ID)
	assert.Equal(t, testUser.Username, created.Username)
	assert.Equal(t, testUser.Email, created.Email)
	assert.Equal(t, testUser.Roles, created.Roles)
	assert.Equal(t, testUser.Active, created.Active)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	testUser := &userdomain.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []*userdomain.Role{{Name: "user"}},
		Active:   true,
	}

	// Create user
	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	// Get user by ID
	retrieved, err := repo.GetByID(ctx, testUser.ID)
	require.NoError(t, err)
	assert.Equal(t, testUser.ID, retrieved.ID)
	assert.Equal(t, testUser.Username, retrieved.Username)
	assert.Equal(t, testUser.Email, retrieved.Email)
	assert.Equal(t, testUser.Roles, retrieved.Roles)
	assert.Equal(t, testUser.Active, retrieved.Active)

	// Test non-existent user
	nonExistentID := uuid.New()
	_, err = repo.GetByID(ctx, nonExistentID)
	require.Error(t, err)
	assert.Equal(t, userdomain.ErrNotFound, err)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create test users
	users := []*userdomain.User{
		{
			ID:       uuid.New(),
			Username: "user1",
			Email:    "user1@example.com",
			Roles:    []*userdomain.Role{{Name: "user"}},
			Active:   true,
		},
		{
			ID:       uuid.New(),
			Username: "user2",
			Email:    "user2@example.com",
			Roles:    []*userdomain.Role{{Name: "admin"}},
			Active:   true,
		},
	}

	// Create users
	for _, u := range users {
		err := repo.Create(ctx, u)
		require.NoError(t, err)
	}

	// List users
	list, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
)

// userRepository implements repo.UserRepository interface
type userRepository struct {
	db *Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *Database) repo.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *repo.User) error {
	// TODO: Implement
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*repo.User, error) {
	// TODO: Implement
	return nil, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*repo.User, error) {
	// TODO: Implement
	return nil, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*repo.User, error) {
	// TODO: Implement
	return nil, nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*repo.User, error) {
	// TODO: Implement
	return nil, nil
}

func (r *userRepository) Update(ctx context.Context, user *repo.User) error {
	// TODO: Implement
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement
	return nil
}

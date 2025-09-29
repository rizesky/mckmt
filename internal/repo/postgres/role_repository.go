package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
)

// roleRepository implements repo.RoleRepository interface
type roleRepository struct {
	db *Database
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *Database) repo.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *repo.Role) error {
	// TODO: Implement
	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*repo.Role, error) {
	// TODO: Implement
	return nil, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*repo.Role, error) {
	// TODO: Implement
	return nil, nil
}

func (r *roleRepository) List(ctx context.Context, limit, offset int) ([]*repo.Role, error) {
	// TODO: Implement
	return nil, nil
}

func (r *roleRepository) Update(ctx context.Context, role *repo.Role) error {
	// TODO: Implement
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement
	return nil
}

package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/user"
)

// permissionRepository implements repo.PermissionRepository interface
type permissionRepository struct {
	db *Database
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *Database) repo.PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *user.Permission) error {
	query := `
		INSERT INTO permissions (id, name, resource, action, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.pool.Exec(ctx, query, permission.ID, permission.Name, permission.Resource, permission.Action, permission.Description, permission.CreatedAt, permission.UpdatedAt)
	return err
}

func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		WHERE id = $1
	`
	var permission user.Permission
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&permission.ID, &permission.Name, &permission.Resource, &permission.Action,
		&permission.Description, &permission.CreatedAt, &permission.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByName(ctx context.Context, name string) (*user.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		WHERE name = $1
	`
	var permission user.Permission
	err := r.db.pool.QueryRow(ctx, query, name).Scan(
		&permission.ID, &permission.Name, &permission.Resource, &permission.Action,
		&permission.Description, &permission.CreatedAt, &permission.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByResourceAction(ctx context.Context, resource, action string) (*user.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		WHERE resource = $1 AND action = $2
	`
	var permission user.Permission
	err := r.db.pool.QueryRow(ctx, query, resource, action).Scan(
		&permission.ID, &permission.Name, &permission.Resource, &permission.Action,
		&permission.Description, &permission.CreatedAt, &permission.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) List(ctx context.Context, limit, offset int) ([]*user.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		ORDER BY resource, action
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*user.Permission
	for rows.Next() {
		var permission user.Permission
		err := rows.Scan(
			&permission.ID, &permission.Name, &permission.Resource, &permission.Action,
			&permission.Description, &permission.CreatedAt, &permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}
	return permissions, nil
}

func (r *permissionRepository) ListByResource(ctx context.Context, resource string) ([]*user.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		WHERE resource = $1
		ORDER BY action
	`
	rows, err := r.db.pool.Query(ctx, query, resource)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*user.Permission
	for rows.Next() {
		var permission user.Permission
		err := rows.Scan(
			&permission.ID, &permission.Name, &permission.Resource, &permission.Action,
			&permission.Description, &permission.CreatedAt, &permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}
	return permissions, nil
}

func (r *permissionRepository) Update(ctx context.Context, permission *user.Permission) error {
	query := `
		UPDATE permissions
		SET name = $2, resource = $3, action = $4, description = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.pool.Exec(ctx, query, permission.ID, permission.Name, permission.Resource, permission.Action, permission.Description, permission.UpdatedAt)
	return err
}

func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM permissions WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id)
	return err
}

// GetUserPermissions returns all permissions for a user (through their roles)
func (r *permissionRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*user.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*user.Permission
	for rows.Next() {
		var permission user.Permission
		err := rows.Scan(
			&permission.ID, &permission.Name, &permission.Resource, &permission.Action,
			&permission.Description, &permission.CreatedAt, &permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}
	return permissions, nil
}

// GetUserPermissionsByResource returns all permissions for a user for a specific resource
func (r *permissionRepository) GetUserPermissionsByResource(ctx context.Context, userID uuid.UUID, resource string) ([]*user.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND p.resource = $2
		ORDER BY p.action
	`
	rows, err := r.db.pool.Query(ctx, query, userID, resource)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*user.Permission
	for rows.Next() {
		var permission user.Permission
		err := rows.Scan(
			&permission.ID, &permission.Name, &permission.Resource, &permission.Action,
			&permission.Description, &permission.CreatedAt, &permission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}
	return permissions, nil
}

// CheckUserPermission checks if a user has a specific permission (supports wildcards)
func (r *permissionRepository) CheckUserPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND (
			(p.resource = '*' AND p.action = '*') OR
			(p.resource = '*' AND p.action = $3) OR
			(p.resource = $2 AND p.action = '*') OR
			(p.resource = $2 AND p.action = $3)
		)
	`
	var hasPermission bool
	err := r.db.pool.QueryRow(ctx, query, userID, resource, action).Scan(&hasPermission)
	return hasPermission, err
}

// CheckUserPermissionExact checks if a user has an exact permission (no wildcard matching)
func (r *permissionRepository) CheckUserPermissionExact(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND p.resource = $2 AND p.action = $3
	`
	var hasPermission bool
	err := r.db.pool.QueryRow(ctx, query, userID, resource, action).Scan(&hasPermission)
	return hasPermission, err
}

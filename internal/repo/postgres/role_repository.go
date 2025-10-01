package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/user"
)

// roleRepository implements repo.RoleRepository interface
type roleRepository struct {
	db *Database
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *Database) repo.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *user.Role) error {
	query := `
		INSERT INTO roles (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.pool.Exec(ctx, query, role.ID, role.Name, role.Description, role.CreatedAt, role.UpdatedAt)
	return err
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.Role, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE id = $1
	`
	var role user.Role
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*user.Role, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE name = $1
	`
	var role user.Role
	err := r.db.pool.QueryRow(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context, limit, offset int) ([]*user.Role, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		ORDER BY name
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*user.Role
	for rows.Next() {
		var role user.Role
		err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *user.Role) error {
	query := `
		UPDATE roles
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.db.pool.Exec(ctx, query, role.ID, role.Name, role.Description, role.UpdatedAt)
	return err
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id)
	return err
}

// GetUserRoles returns all roles assigned to a user
func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*user.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`
	rows, err := r.db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*user.Role
	for rows.Next() {
		var role user.Role
		err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

// AssignRoleToUser assigns a role to a user
func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err := r.db.pool.Exec(ctx, query, userID, roleID, assignedBy)
	return err
}

// RemoveRoleFromUser removes a role from a user
func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.pool.Exec(ctx, query, userID, roleID)
	return err
}

// GetRolePermissions returns all permissions for a role
func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*user.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.db.pool.Query(ctx, query, roleID)
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

// AssignPermissionToRole assigns a permission to a role
func (r *roleRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID, grantedBy *uuid.UUID) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, granted_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	_, err := r.db.pool.Exec(ctx, query, roleID, permissionID, grantedBy)
	return err
}

// RemovePermissionFromRole removes a permission from a role
func (r *roleRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.db.pool.Exec(ctx, query, roleID, permissionID)
	return err
}

package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/user"
	"github.com/rizesky/mckmt/internal/utils"
)

// userRepository implements repo.UserRepository interface
type userRepository struct {
	db *Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *Database) repo.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *user.User) error {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO users (id, username, email, password_hash, auth_source, roles, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now().UTC()
	_, err := r.db.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.AuthSource,
		user.Roles,
		user.Active,
		now,
		now,
	)

	if err != nil {
		return utils.ErrCreate("user", err)
	}

	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, auth_source, roles, active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user user.User
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AuthSource,
		&user.Roles,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, repo.ErrNotFound
		}
		return nil, utils.ErrGet("user", err)
	}

	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, auth_source, roles, active, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user user.User
	err := r.db.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AuthSource,
		&user.Roles,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, repo.ErrNotFound
		}
		return nil, utils.ErrGet("user", err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, auth_source, roles, active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user user.User
	err := r.db.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AuthSource,
		&user.Roles,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, repo.ErrNotFound
		}
		return nil, utils.ErrGet("user", err)
	}

	return &user, nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, auth_source, roles, active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, utils.ErrGet("users", err)
	}
	defer rows.Close()

	users := make([]*user.User, 0)
	for rows.Next() {
		var user user.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.AuthSource,
			&user.Roles,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, utils.ErrGet("user", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.ErrGet("users", err)
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *user.User) error {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, auth_source = $5, roles = $6, active = $7, updated_at = $8
		WHERE id = $1
	`

	now := time.Now().UTC()
	result, err := r.db.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.AuthSource,
		user.Roles,
		user.Active,
		now,
	)

	if err != nil {
		return utils.ErrUpdate("user", err)
	}

	if result.RowsAffected() == 0 {
		return repo.ErrNotFound
	}

	user.UpdatedAt = now
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.pool.Exec(ctx, query, id)
	if err != nil {
		return utils.ErrUpdate("user", err)
	}

	if result.RowsAffected() == 0 {
		return repo.ErrNotFound
	}

	return nil
}

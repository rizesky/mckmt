package postgres

import (
	"context"

	"github.com/rizesky/mckmt/internal/repo"
)

// auditLogRepository implements repo.AuditLogRepository interface
type auditLogRepository struct {
	db *Database
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *Database) repo.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *repo.AuditLog) error {
	// TODO: Implement
	return nil
}

func (r *auditLogRepository) List(ctx context.Context, userID string, limit, offset int) ([]*repo.AuditLog, error) {
	// TODO: Implement
	return nil, nil
}

func (r *auditLogRepository) ListByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*repo.AuditLog, error) {
	// TODO: Implement
	return nil, nil
}

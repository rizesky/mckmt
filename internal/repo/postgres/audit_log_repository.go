package postgres

import (
	"context"
	"time"

	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/utils"
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
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, request_payload, response_payload, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.pool.Exec(ctx, query,
		log.ID,
		log.UserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.RequestPayload,
		log.ResponsePayload,
		log.IPAddress,
		log.UserAgent,
		log.CreatedAt,
	)
	return err
}

func (r *auditLogRepository) List(ctx context.Context, userID string, limit, offset int) ([]*repo.AuditLog, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, user_id, action, resource_type, resource_id, request_payload, response_payload, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*repo.AuditLog
	for rows.Next() {
		var log repo.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.RequestPayload,
			&log.ResponsePayload,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

func (r *auditLogRepository) ListByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*repo.AuditLog, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, user_id, action, resource_type, resource_id, request_payload, response_payload, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.pool.Query(ctx, query, resourceType, resourceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*repo.AuditLog
	for rows.Next() {
		var log repo.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.RequestPayload,
			&log.ResponsePayload,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

// Additional methods for comprehensive audit logging

// ListAll returns all audit logs with pagination
func (r *auditLogRepository) ListAll(ctx context.Context, limit, offset int) ([]*repo.AuditLog, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, user_id, action, resource_type, resource_id, request_payload, response_payload, ip_address, user_agent, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*repo.AuditLog
	for rows.Next() {
		var log repo.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.RequestPayload,
			&log.ResponsePayload,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

// ListByAction returns audit logs filtered by action
func (r *auditLogRepository) ListByAction(ctx context.Context, action string, limit, offset int) ([]*repo.AuditLog, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, user_id, action, resource_type, resource_id, request_payload, response_payload, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE action = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.pool.Query(ctx, query, action, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*repo.AuditLog
	for rows.Next() {
		var log repo.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.RequestPayload,
			&log.ResponsePayload,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

// ListByDateRange returns audit logs within a date range
func (r *auditLogRepository) ListByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*repo.AuditLog, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, user_id, action, resource_type, resource_id, request_payload, response_payload, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.pool.Query(ctx, query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*repo.AuditLog
	for rows.Next() {
		var log repo.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.RequestPayload,
			&log.ResponsePayload,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

// Count returns the total number of audit logs
func (r *auditLogRepository) Count(ctx context.Context) (int64, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `SELECT COUNT(*) FROM audit_logs`
	var count int64
	err := r.db.pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountByUser returns the number of audit logs for a specific user
func (r *auditLogRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	query := `SELECT COUNT(*) FROM audit_logs WHERE user_id = $1`
	var count int64
	err := r.db.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

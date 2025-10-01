package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
	"github.com/rizesky/mckmt/internal/utils"
)

// operationRepository implements repo.OperationRepository interface
type operationRepository struct {
	db *Database
}

// NewOperationRepository creates a new operation repository
func NewOperationRepository(db *Database) repo.OperationRepository {
	return &operationRepository{db: db}
}

func (r *operationRepository) Create(ctx context.Context, operation *repo.Operation) error {
	query := `
		INSERT INTO operations (id, cluster_id, type, status, payload, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	payloadJSON, err := json.Marshal(operation.Payload)
	if err != nil {
		return utils.ErrMarshal("payload", err)
	}

	now := time.Now().UTC()
	_, err = r.db.pool.Exec(ctx, query,
		operation.ID,
		operation.ClusterID,
		operation.Type,
		operation.Status,
		string(payloadJSON),
		now,
		now,
	)

	if err != nil {
		return utils.ErrCreate("operation", err)
	}

	return nil
}

func (r *operationRepository) GetByID(ctx context.Context, id uuid.UUID) (*repo.Operation, error) {
	query := `
		SELECT id, cluster_id, type, status, payload, result, started_at, finished_at, created_at, updated_at
		FROM operations
		WHERE id = $1
	`

	var operation repo.Operation
	var payloadJSON, resultJSON string

	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&operation.ID,
		&operation.ClusterID,
		&operation.Type,
		&operation.Status,
		&payloadJSON,
		&resultJSON,
		&operation.StartedAt,
		&operation.FinishedAt,
		&operation.CreatedAt,
		&operation.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}

	// Parse payload
	if err := json.Unmarshal([]byte(payloadJSON), &operation.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Parse result if exists
	if resultJSON != "" {
		var resultPayload repo.Payload
		if err := json.Unmarshal([]byte(resultJSON), &resultPayload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
		operation.Result = &resultPayload
	}

	return &operation, nil
}

func (r *operationRepository) ListByCluster(ctx context.Context, clusterID uuid.UUID, limit, offset int) ([]*repo.Operation, error) {
	query := `
		SELECT id, cluster_id, type, status, payload, result, started_at, finished_at, created_at, updated_at
		FROM operations
		WHERE cluster_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.pool.Query(ctx, query, clusterID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list operations: %w", err)
	}
	defer rows.Close()

	operations := make([]*repo.Operation, 0)
	for rows.Next() {
		var operation repo.Operation
		var payloadJSON, resultJSON string

		err := rows.Scan(
			&operation.ID,
			&operation.ClusterID,
			&operation.Type,
			&operation.Status,
			&payloadJSON,
			&resultJSON,
			&operation.StartedAt,
			&operation.FinishedAt,
			&operation.CreatedAt,
			&operation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}

		// Parse payload
		if err := json.Unmarshal([]byte(payloadJSON), &operation.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		// Parse result if exists
		if resultJSON != "" {
			var resultPayload repo.Payload
			if err := json.Unmarshal([]byte(resultJSON), &resultPayload); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}
			operation.Result = &resultPayload
		}

		operations = append(operations, &operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating operations: %w", err)
	}

	return operations, nil
}

func (r *operationRepository) Update(ctx context.Context, operation *repo.Operation) error {
	query := `
		UPDATE operations 
		SET status = $2, payload = $3, result = $4, started_at = $5, finished_at = $6, updated_at = $7
		WHERE id = $1
	`

	payloadJSON, err := json.Marshal(operation.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var resultJSON *string
	if operation.Result != nil {
		resultBytes, err := json.Marshal(*operation.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		resultStr := string(resultBytes)
		resultJSON = &resultStr
	}

	now := time.Now().UTC()
	_, err = r.db.pool.Exec(ctx, query,
		operation.ID,
		operation.Status,
		string(payloadJSON),
		resultJSON,
		operation.StartedAt,
		operation.FinishedAt,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to update operation: %w", err)
	}

	return nil
}

func (r *operationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE operations 
		SET status = $2, updated_at = now()
		WHERE id = $1
	`

	_, err := r.db.pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update operation status: %w", err)
	}

	return nil
}

func (r *operationRepository) UpdateResult(ctx context.Context, id uuid.UUID, result repo.Payload) error {
	query := `
		UPDATE operations 
		SET result = $2, updated_at = now()
		WHERE id = $1
	`

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	_, err = r.db.pool.Exec(ctx, query, id, string(resultJSON))
	if err != nil {
		return fmt.Errorf("failed to update operation result: %w", err)
	}

	return nil
}

func (r *operationRepository) SetStarted(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE operations 
		SET status = 'running', started_at = now(), updated_at = now()
		WHERE id = $1 AND status = 'queued'
	`

	result, err := r.db.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to set operation as started: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("operation not found or not in queued status")
	}

	return nil
}

func (r *operationRepository) SetFinished(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE operations 
		SET finished_at = now(), updated_at = now()
		WHERE id = $1 AND status IN ('running', 'success', 'failed', 'cancelled')
	`

	result, err := r.db.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to set operation as finished: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("operation not found or not in a finishable status")
	}

	return nil
}

func (r *operationRepository) CancelOperation(ctx context.Context, id uuid.UUID, reason string) error {
	query := `
		UPDATE operations 
		SET status = 'cancelled', 
		    result = COALESCE(result, '{}'::jsonb) || $2::jsonb,
		    updated_at = now()
		WHERE id = $1 AND status IN ('queued', 'running')
	`

	resultPayload := map[string]interface{}{
		"cancelled":    true,
		"reason":       reason,
		"cancelled_at": time.Now().UTC(),
	}

	resultJSON, err := json.Marshal(resultPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	result, err := r.db.pool.Exec(ctx, query, id, string(resultJSON))
	if err != nil {
		return fmt.Errorf("failed to cancel operation: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("operation not found or cannot be cancelled")
	}

	return nil
}

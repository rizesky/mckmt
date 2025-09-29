package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
)

// clusterRepository implements repo.ClusterRepository interface
type clusterRepository struct {
	db *Database
}

// NewClusterRepository creates a new cluster repository
func NewClusterRepository(db *Database) repo.ClusterRepository {
	return &clusterRepository{db: db}
}

func (r *clusterRepository) Create(ctx context.Context, cluster *repo.Cluster) error {
	query := `
		INSERT INTO clusters (id, name, description, labels, encrypted_credentials, status, last_seen_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.pool.Exec(ctx, query,
		cluster.ID, cluster.Name, cluster.Description, cluster.Labels, cluster.EncryptedCredentials,
		cluster.Status, cluster.LastSeenAt, cluster.CreatedAt, cluster.UpdatedAt)
	return err
}

func (r *clusterRepository) GetByID(ctx context.Context, id uuid.UUID) (*repo.Cluster, error) {
	query := `
		SELECT id, name, description, labels, encrypted_credentials, status, last_seen_at, created_at, updated_at
		FROM clusters WHERE id = $1
	`
	cluster := &repo.Cluster{}
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&cluster.ID, &cluster.Name, &cluster.Description, &cluster.Labels, &cluster.EncryptedCredentials,
		&cluster.Status, &cluster.LastSeenAt, &cluster.CreatedAt, &cluster.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (r *clusterRepository) GetByName(ctx context.Context, name string) (*repo.Cluster, error) {
	query := `
		SELECT id, name, description, labels, encrypted_credentials, status, last_seen_at, created_at, updated_at
		FROM clusters WHERE name = $1
	`
	cluster := &repo.Cluster{}
	err := r.db.pool.QueryRow(ctx, query, name).Scan(
		&cluster.ID, &cluster.Name, &cluster.Description, &cluster.Labels, &cluster.EncryptedCredentials,
		&cluster.Status, &cluster.LastSeenAt, &cluster.CreatedAt, &cluster.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (r *clusterRepository) List(ctx context.Context, limit, offset int) ([]*repo.Cluster, error) {
	query := `
		SELECT id, name, description, labels, encrypted_credentials, status, last_seen_at, created_at, updated_at
		FROM clusters ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []*repo.Cluster
	for rows.Next() {
		cluster := &repo.Cluster{}
		err := rows.Scan(
			&cluster.ID, &cluster.Name, &cluster.Description, &cluster.Labels, &cluster.EncryptedCredentials,
			&cluster.Status, &cluster.LastSeenAt, &cluster.CreatedAt, &cluster.UpdatedAt)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func (r *clusterRepository) Update(ctx context.Context, cluster *repo.Cluster) error {
	query := `
		UPDATE clusters 
		SET name = $2, description = $3, labels = $4, encrypted_credentials = $5, status = $6, 
		    last_seen_at = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.pool.Exec(ctx, query,
		cluster.ID, cluster.Name, cluster.Description, cluster.Labels, cluster.EncryptedCredentials,
		cluster.Status, cluster.LastSeenAt, cluster.UpdatedAt)
	return err
}

func (r *clusterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM clusters WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id)
	return err
}

func (r *clusterRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE clusters SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id, status, time.Now())
	return err
}

func (r *clusterRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE clusters SET last_seen_at = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id, time.Now(), time.Now())
	return err
}

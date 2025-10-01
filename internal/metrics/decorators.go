package metrics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rizesky/mckmt/internal/repo"
	"go.uber.org/zap"
)

// Repository decorators - clean separation of concerns

// WithClusterRepositoryMetrics decorates a ClusterRepository with metrics
func WithClusterRepositoryMetrics(repo repo.ClusterRepository, metrics *Metrics, logger *zap.Logger) repo.ClusterRepository {
	if metrics == nil {
		return repo // Return original if no metrics
	}
	return &ClusterRepositoryDecorator{
		repo:    repo,
		metrics: metrics,
		logger:  logger,
	}
}

// WithOperationRepositoryMetrics decorates an OperationRepository with metrics
func WithOperationRepositoryMetrics(repo repo.OperationRepository, metrics *Metrics, logger *zap.Logger) repo.OperationRepository {
	if metrics == nil {
		return repo // Return original if no metrics
	}
	return &OperationRepositoryDecorator{
		repo:    repo,
		metrics: metrics,
		logger:  logger,
	}
}

// ClusterRepositoryDecorator wraps a ClusterRepository with metrics
type ClusterRepositoryDecorator struct {
	repo    repo.ClusterRepository
	metrics *Metrics
	logger  *zap.Logger
}

func (d *ClusterRepositoryDecorator) Create(ctx context.Context, cluster *repo.Cluster) error {
	start := time.Now()
	err := d.repo.Create(ctx, cluster)

	d.metrics.DatabaseQueryDuration.WithLabelValues("create", "clusters").Observe(time.Since(start).Seconds())
	return err
}

func (d *ClusterRepositoryDecorator) GetByID(ctx context.Context, id uuid.UUID) (*repo.Cluster, error) {
	start := time.Now()
	cluster, err := d.repo.GetByID(ctx, id)

	d.metrics.DatabaseQueryDuration.WithLabelValues("get", "clusters").Observe(time.Since(start).Seconds())
	return cluster, err
}

func (d *ClusterRepositoryDecorator) GetByName(ctx context.Context, name string) (*repo.Cluster, error) {
	start := time.Now()
	cluster, err := d.repo.GetByName(ctx, name)

	d.metrics.DatabaseQueryDuration.WithLabelValues("get", "clusters").Observe(time.Since(start).Seconds())
	return cluster, err
}

func (d *ClusterRepositoryDecorator) List(ctx context.Context, limit, offset int) ([]*repo.Cluster, error) {
	start := time.Now()
	clusters, err := d.repo.List(ctx, limit, offset)

	d.metrics.DatabaseQueryDuration.WithLabelValues("list", "clusters").Observe(time.Since(start).Seconds())
	return clusters, err
}

func (d *ClusterRepositoryDecorator) Update(ctx context.Context, cluster *repo.Cluster) error {
	start := time.Now()
	err := d.repo.Update(ctx, cluster)

	d.metrics.DatabaseQueryDuration.WithLabelValues("update", "clusters").Observe(time.Since(start).Seconds())
	return err
}

func (d *ClusterRepositoryDecorator) Delete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	err := d.repo.Delete(ctx, id)

	d.metrics.DatabaseQueryDuration.WithLabelValues("delete", "clusters").Observe(time.Since(start).Seconds())
	return err
}

func (d *ClusterRepositoryDecorator) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	start := time.Now()
	err := d.repo.UpdateStatus(ctx, id, status)

	d.metrics.DatabaseQueryDuration.WithLabelValues("update_status", "clusters").Observe(time.Since(start).Seconds())
	return err
}

func (d *ClusterRepositoryDecorator) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	err := d.repo.UpdateLastSeen(ctx, id)

	d.metrics.DatabaseQueryDuration.WithLabelValues("update_last_seen", "clusters").Observe(time.Since(start).Seconds())
	return err
}

// OperationRepositoryDecorator wraps an OperationRepository with metrics
type OperationRepositoryDecorator struct {
	repo    repo.OperationRepository
	metrics *Metrics
	logger  *zap.Logger
}

func (d *OperationRepositoryDecorator) Create(ctx context.Context, operation *repo.Operation) error {
	start := time.Now()
	err := d.repo.Create(ctx, operation)

	d.metrics.DatabaseQueryDuration.WithLabelValues("create", "operations").Observe(time.Since(start).Seconds())
	return err
}

func (d *OperationRepositoryDecorator) GetByID(ctx context.Context, id uuid.UUID) (*repo.Operation, error) {
	start := time.Now()
	operation, err := d.repo.GetByID(ctx, id)

	d.metrics.DatabaseQueryDuration.WithLabelValues("get", "operations").Observe(time.Since(start).Seconds())
	return operation, err
}

func (d *OperationRepositoryDecorator) ListByCluster(ctx context.Context, clusterID uuid.UUID, limit, offset int) ([]*repo.Operation, error) {
	start := time.Now()
	operations, err := d.repo.ListByCluster(ctx, clusterID, limit, offset)

	d.metrics.DatabaseQueryDuration.WithLabelValues("list", "operations").Observe(time.Since(start).Seconds())
	return operations, err
}

func (d *OperationRepositoryDecorator) Update(ctx context.Context, operation *repo.Operation) error {
	start := time.Now()
	err := d.repo.Update(ctx, operation)

	d.metrics.DatabaseQueryDuration.WithLabelValues("update", "operations").Observe(time.Since(start).Seconds())
	return err
}

func (d *OperationRepositoryDecorator) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	start := time.Now()
	err := d.repo.UpdateStatus(ctx, id, status)

	d.metrics.DatabaseQueryDuration.WithLabelValues("update_status", "operations").Observe(time.Since(start).Seconds())
	return err
}

func (d *OperationRepositoryDecorator) UpdateResult(ctx context.Context, id uuid.UUID, result repo.Payload) error {
	start := time.Now()
	err := d.repo.UpdateResult(ctx, id, result)

	d.metrics.DatabaseQueryDuration.WithLabelValues("update_result", "operations").Observe(time.Since(start).Seconds())
	return err
}

func (d *OperationRepositoryDecorator) SetStarted(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	err := d.repo.SetStarted(ctx, id)

	d.metrics.DatabaseQueryDuration.WithLabelValues("set_started", "operations").Observe(time.Since(start).Seconds())
	return err
}

func (d *OperationRepositoryDecorator) SetFinished(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	err := d.repo.SetFinished(ctx, id)

	d.metrics.DatabaseQueryDuration.WithLabelValues("set_finished", "operations").Observe(time.Since(start).Seconds())
	return err
}

func (d *OperationRepositoryDecorator) CancelOperation(ctx context.Context, id uuid.UUID, reason string) error {
	start := time.Now()
	err := d.repo.CancelOperation(ctx, id, reason)

	d.metrics.DatabaseQueryDuration.WithLabelValues("cancel", "operations").Observe(time.Since(start).Seconds())
	return err
}

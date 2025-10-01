package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/utils"
)

// Manager handles Redis cache operations
type Manager struct {
	client *redis.Client
	logger *zap.Logger
}

// NewManager creates a new cache manager
func NewManager(client *redis.Client, logger *zap.Logger) *Manager {
	return &Manager{
		client: client,
		logger: logger,
	}
}

// Set stores a value in cache with expiration
func (m *Manager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return m.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value from cache
func (m *Manager) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := m.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (m *Manager) Delete(ctx context.Context, key string) error {
	return m.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in cache
func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	result, err := m.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return result > 0, nil
}

// SetNX sets a value only if the key doesn't exist
func (m *Manager) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return m.client.SetNX(ctx, key, data, expiration).Result()
}

// Increment increments a counter
func (m *Manager) Increment(ctx context.Context, key string) (int64, error) {
	return m.client.Incr(ctx, key).Result()
}

// IncrementBy increments a counter by a specific amount
func (m *Manager) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return m.client.IncrBy(ctx, key, value).Result()
}

// Expire sets expiration for a key
func (m *Manager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return m.client.Expire(ctx, key, expiration).Err()
}

// TTL returns the time to live for a key
func (m *Manager) TTL(ctx context.Context, key string) (time.Duration, error) {
	return m.client.TTL(ctx, key).Result()
}

// Keys returns all keys matching a pattern
func (m *Manager) Keys(ctx context.Context, pattern string) ([]string, error) {
	return m.client.Keys(ctx, pattern).Result()
}

// FlushDB flushes the current database
func (m *Manager) FlushDB(ctx context.Context) error {
	return m.client.FlushDB(ctx).Err()
}

// Ping tests the connection
func (m *Manager) Ping(ctx context.Context) error {
	return m.client.Ping(ctx).Err()
}

// Health checks the cache health
func (m *Manager) Health(ctx context.Context) error {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	if err := m.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("cache health check failed: %w", err)
	}

	return nil
}

// Cache key generators
func (m *Manager) ClusterKey(id string) string {
	return fmt.Sprintf("cluster:%s", id)
}

func (m *Manager) OperationKey(id string) string {
	return fmt.Sprintf("operation:%s", id)
}

func (m *Manager) UserKey(id string) string {
	return fmt.Sprintf("user:%s", id)
}

func (m *Manager) SessionKey(token string) string {
	return fmt.Sprintf("session:%s", token)
}

// Cluster resource cache keys
func (m *Manager) ClusterResourcesKey(clusterID, kind, namespace string) string {
	if namespace == "" {
		return fmt.Sprintf("cluster:%s:resources:%s", clusterID, kind)
	}
	return fmt.Sprintf("cluster:%s:resources:%s:%s", clusterID, kind, namespace)
}

func (m *Manager) ClusterResourceKey(clusterID, kind, namespace, name string) string {
	if namespace == "" {
		return fmt.Sprintf("cluster:%s:resource:%s:%s", clusterID, kind, name)
	}
	return fmt.Sprintf("cluster:%s:resource:%s:%s:%s", clusterID, kind, namespace, name)
}

func (m *Manager) ClusterStatusKey(clusterID string) string {
	return fmt.Sprintf("cluster:%s:status", clusterID)
}

func (m *Manager) ClusterMetricsKey(clusterID string) string {
	return fmt.Sprintf("cluster:%s:metrics", clusterID)
}

// Cluster resource caching methods
func (m *Manager) SetClusterResources(ctx context.Context, clusterID, kind, namespace string, resources interface{}, expiration time.Duration) error {
	key := m.ClusterResourcesKey(clusterID, kind, namespace)
	return m.Set(ctx, key, resources, expiration)
}

func (m *Manager) GetClusterResources(ctx context.Context, clusterID, kind, namespace string, dest interface{}) error {
	key := m.ClusterResourcesKey(clusterID, kind, namespace)
	return m.Get(ctx, key, dest)
}

func (m *Manager) SetClusterResource(ctx context.Context, clusterID, kind, namespace, name string, resource interface{}, expiration time.Duration) error {
	key := m.ClusterResourceKey(clusterID, kind, namespace, name)
	return m.Set(ctx, key, resource, expiration)
}

func (m *Manager) GetClusterResource(ctx context.Context, clusterID, kind, namespace, name string, dest interface{}) error {
	key := m.ClusterResourceKey(clusterID, kind, namespace, name)
	return m.Get(ctx, key, dest)
}

func (m *Manager) SetClusterStatus(ctx context.Context, clusterID string, status interface{}, expiration time.Duration) error {
	key := m.ClusterStatusKey(clusterID)
	return m.Set(ctx, key, status, expiration)
}

func (m *Manager) GetClusterStatus(ctx context.Context, clusterID string, dest interface{}) error {
	key := m.ClusterStatusKey(clusterID)
	return m.Get(ctx, key, dest)
}

func (m *Manager) SetClusterMetrics(ctx context.Context, clusterID string, metrics interface{}, expiration time.Duration) error {
	key := m.ClusterMetricsKey(clusterID)
	return m.Set(ctx, key, metrics, expiration)
}

func (m *Manager) GetClusterMetrics(ctx context.Context, clusterID string, dest interface{}) error {
	key := m.ClusterMetricsKey(clusterID)
	return m.Get(ctx, key, dest)
}

// Invalidate cluster resource cache
func (m *Manager) InvalidateClusterResources(ctx context.Context, clusterID string) error {
	pattern := fmt.Sprintf("cluster:%s:resources:*", clusterID)
	keys, err := m.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		return m.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (m *Manager) InvalidateClusterResource(ctx context.Context, clusterID, kind, namespace, name string) error {
	key := m.ClusterResourceKey(clusterID, kind, namespace, name)
	return m.client.Del(ctx, key).Err()
}

// Cache errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)

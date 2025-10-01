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

// CacheAdapter adapts Redis client to repo.Cache interface
type CacheAdapter struct {
	client *redis.Client
	logger *zap.Logger
}

// NewCacheAdapter creates a new cache adapter
func NewCacheAdapter(client *redis.Client, logger *zap.Logger) *CacheAdapter {
	return &CacheAdapter{
		client: client,
		logger: logger,
	}
}

// Set stores a value in cache with expiration
func (c *CacheAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value from cache
func (c *CacheAdapter) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
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
func (c *CacheAdapter) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Keys returns all keys matching pattern
func (c *CacheAdapter) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// FlushDB flushes the current database
func (c *CacheAdapter) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Ping tests the connection
func (c *CacheAdapter) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Health checks the cache health
func (c *CacheAdapter) Health(ctx context.Context) error {
	ctx, cancel := utils.WithDefaultTimeout(ctx)
	defer cancel()

	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("cache health check failed: %w", err)
	}

	return nil
}

// Cache key generators
func (c *CacheAdapter) ClusterKey(id string) string {
	return fmt.Sprintf("cluster:%s", id)
}

func (c *CacheAdapter) OperationKey(id string) string {
	return fmt.Sprintf("operation:%s", id)
}

func (c *CacheAdapter) UserKey(id string) string {
	return fmt.Sprintf("user:%s", id)
}

func (c *CacheAdapter) SessionKey(token string) string {
	return fmt.Sprintf("session:%s", token)
}

// Cluster resource cache keys
func (c *CacheAdapter) ClusterResourcesKey(clusterID, kind, namespace string) string {
	if namespace == "" {
		return fmt.Sprintf("cluster:%s:resources:%s", clusterID, kind)
	}
	return fmt.Sprintf("cluster:%s:resources:%s:%s", clusterID, kind, namespace)
}

func (c *CacheAdapter) ClusterResourceKey(clusterID, kind, namespace, name string) string {
	if namespace == "" {
		return fmt.Sprintf("cluster:%s:resource:%s:%s", clusterID, kind, name)
	}
	return fmt.Sprintf("cluster:%s:resource:%s:%s:%s", clusterID, kind, namespace, name)
}

func (c *CacheAdapter) ClusterStatusKey(clusterID string) string {
	return fmt.Sprintf("cluster:%s:status", clusterID)
}

func (c *CacheAdapter) ClusterMetricsKey(clusterID string) string {
	return fmt.Sprintf("cluster:%s:metrics", clusterID)
}

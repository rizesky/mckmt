package utils

import (
	"context"
	"time"
)

// WithDefaultTimeout creates a context with a default timeout
func WithDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}

// WithCustomTimeout creates a context with a custom timeout
func WithCustomTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

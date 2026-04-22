package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// HealthChecker wraps redis.Client to implement handler.HealthChecker
type HealthChecker struct {
	client *redis.Client
}

func NewHealthChecker(client *redis.Client) *HealthChecker {
	return &HealthChecker{client: client}
}

func (h *HealthChecker) Ping(ctx context.Context) error {
	return h.client.Ping(ctx).Err()
}

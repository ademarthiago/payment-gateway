package port

import (
	"context"
	"time"
)

// IdempotencyStore defines the contract for idempotency key storage
// Implemented by the Redis adapter
type IdempotencyStore interface {
	Exists(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}

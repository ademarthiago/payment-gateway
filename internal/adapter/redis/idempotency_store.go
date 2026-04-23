// Package redis contains Redis adapter implementations for the domain ports.
// Currently used only for idempotency key storage — keeping a short-lived record
// of processed requests so retries return the same result instead of creating duplicates.
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IdempotencyStore implements port.IdempotencyStore using Redis.
// Keys are stored with a TTL (default 24h) and expire automatically — no cleanup job needed.
type IdempotencyStore struct {
	client *redis.Client
}

// NewIdempotencyStore creates the store backed by the given Redis client.
func NewIdempotencyStore(client *redis.Client) *IdempotencyStore {
	return &IdempotencyStore{client: client}
}

// Exists checks if the idempotency key is present in Redis.
// A true result means this request was already processed — return the original response.
func (s *IdempotencyStore) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return n > 0, nil
}

// Set stores the idempotency key with a TTL. Called after a payment is successfully created.
// The value is the payment UUID so we can look it up on retry without an extra DB query.
func (s *IdempotencyStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := s.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

// Get retrieves the stored value for an idempotency key.
// Returns nil, nil when the key doesn't exist — expired keys behave the same as missing ones.
func (s *IdempotencyStore) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

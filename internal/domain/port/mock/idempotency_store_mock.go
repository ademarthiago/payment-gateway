package mock

import (
	"context"
	"time"
)

type IdempotencyStoreMock struct {
	ExistsFn func(ctx context.Context, key string) (bool, error)
	SetFn    func(ctx context.Context, key string, value []byte, ttl time.Duration) error
	GetFn    func(ctx context.Context, key string) ([]byte, error)
}

func (m *IdempotencyStoreMock) Exists(ctx context.Context, key string) (bool, error) {
	if m.ExistsFn != nil {
		return m.ExistsFn(ctx, key)
	}
	return false, nil
}

func (m *IdempotencyStoreMock) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.SetFn != nil {
		return m.SetFn(ctx, key, value, ttl)
	}
	return nil
}

func (m *IdempotencyStoreMock) Get(ctx context.Context, key string) ([]byte, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, key)
	}
	return nil, nil
}

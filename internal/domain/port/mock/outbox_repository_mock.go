package mock

import (
	"context"

	"github.com/ademarthiago/payment-gateway/internal/domain/port"
	"github.com/google/uuid"
)

type OutboxRepositoryMock struct {
	SaveFn          func(ctx context.Context, msg *port.OutboxMessage) error
	FetchPendingFn  func(ctx context.Context, limit int) ([]*port.OutboxMessage, error)
	MarkProcessedFn func(ctx context.Context, id uuid.UUID) error
	MarkFailedFn    func(ctx context.Context, id uuid.UUID, errMsg string) error
}

func (m *OutboxRepositoryMock) Save(ctx context.Context, msg *port.OutboxMessage) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, msg)
	}
	return nil
}

func (m *OutboxRepositoryMock) FetchPending(ctx context.Context, limit int) ([]*port.OutboxMessage, error) {
	if m.FetchPendingFn != nil {
		return m.FetchPendingFn(ctx, limit)
	}
	return nil, nil
}

func (m *OutboxRepositoryMock) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	if m.MarkProcessedFn != nil {
		return m.MarkProcessedFn(ctx, id)
	}
	return nil
}

func (m *OutboxRepositoryMock) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	if m.MarkFailedFn != nil {
		return m.MarkFailedFn(ctx, id, errMsg)
	}
	return nil
}

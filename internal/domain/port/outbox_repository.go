package port

import (
	"context"

	"github.com/google/uuid"
)

// OutboxMessage represents an event to be published
type OutboxMessage struct {
	ID            uuid.UUID
	AggregateID   uuid.UUID
	AggregateType string
	EventType     string
	Payload       []byte
}

// OutboxRepository defines the contract for outbox persistence
type OutboxRepository interface {
	Save(ctx context.Context, msg *OutboxMessage) error
	FetchPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	MarkProcessed(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}

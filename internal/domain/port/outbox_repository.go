// Package port defines the outbound interfaces (ports) that the domain needs from infrastructure.
// Adapters in internal/adapter/ implement these interfaces — the domain never imports them directly.
package port

import (
	"context"

	"github.com/google/uuid"
)

// OutboxMessage is the event envelope written to the outbox table alongside a domain operation.
// Persisting it in the same DB write as the payment guarantees the event is never lost,
// even if the process crashes before the event reaches a broker or channel.
type OutboxMessage struct {
	ID            uuid.UUID
	AggregateID   uuid.UUID // the payment ID this event belongs to
	AggregateType string    // always "payment" for now
	EventType     string    // e.g. "payment.created", "payment.refunded"
	Payload       []byte    // JSON-encoded event data
}

// OutboxRepository is the persistence port for the outbox table.
// The concrete implementation lives in adapter/postgres/outbox_repository.go.
type OutboxRepository interface {
	// Save writes a new pending outbox message. Called inside the same flow as the domain write.
	Save(ctx context.Context, msg *OutboxMessage) error
	// FetchPending returns up to limit unprocessed messages for the outbox worker to deliver.
	FetchPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	// MarkProcessed flags a message as delivered. Called after successful event publication.
	MarkProcessed(ctx context.Context, id uuid.UUID) error
	// MarkFailed records a delivery failure with an error message for debugging.
	// The worker will not retry automatically — a separate cleanup job or manual intervention is needed.
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}

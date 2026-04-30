// Package port defines the outbound interfaces (ports) that the domain needs from infrastructure.
// Adapters in internal/adapter/ implement these interfaces — the domain never imports them directly.
package port

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// MaxOutboxRetries is the maximum number of delivery attempts before a message is marked failed.
const MaxOutboxRetries = 5

var ErrDuplicateExternalID = errors.New("payment with this external_id already exists")

// OutboxMessage is the event envelope written to the outbox table alongside a domain operation.
// Persisting it in the same DB write as the payment guarantees the event is never lost,
// even if the process crashes before the event reaches a broker or channel.
type OutboxMessage struct {
	ID            uuid.UUID
	AggregateID   uuid.UUID // the payment ID this event belongs to
	AggregateType string    // always "payment" for now
	EventType     string    // e.g. "payment.created", "payment.refunded"
	Payload       []byte    // JSON-encoded event data
	RetryCount    int       // number of failed delivery attempts so far
	LastError     string    // last error message recorded by the worker
}

// OutboxRepository is the persistence port for the outbox table.
// The concrete implementation lives in adapter/postgres/outbox_repository.go.
type OutboxRepository interface {
	// Save writes a new pending outbox message. Called inside the same flow as the domain write.
	Save(ctx context.Context, msg *OutboxMessage) error
	// FetchPending returns up to limit unprocessed messages for the outbox worker to deliver.
	// Uses SELECT FOR UPDATE SKIP LOCKED — the caller must hold a transaction for the lock to be useful.
	FetchPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	// MarkProcessed flags a message as delivered. Called after successful event publication.
	MarkProcessed(ctx context.Context, id uuid.UUID) error
	// MarkFailed increments retry_count and records the error. Once retry_count reaches
	// MaxOutboxRetries the status transitions to 'failed'; otherwise it stays 'pending'.
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ademarthiago/payment-gateway/internal/domain/port"
)

// OutboxRepository implements port.OutboxRepository using PostgreSQL.
type OutboxRepository struct {
	pool *pgxpool.Pool
}

// NewOutboxRepository creates a repository backed by the given connection pool.
func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

// Save inserts a new outbox message with status='pending'.
// msg.Payload must already be a valid JSON byte slice.
func (r *OutboxRepository) Save(ctx context.Context, msg *port.OutboxMessage) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO payment.outbox
			(id, aggregate_id, aggregate_type, event_type, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6)`,
		msg.ID, msg.AggregateID, msg.AggregateType, msg.EventType, msg.Payload, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to save outbox message: %w", err)
	}
	return nil
}

// FetchPending returns up to limit pending messages that have not exceeded MaxOutboxRetries.
// SELECT FOR UPDATE SKIP LOCKED prevents concurrent workers from picking the same rows.
func (r *OutboxRepository) FetchPending(ctx context.Context, limit int) ([]*port.OutboxMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, aggregate_id, aggregate_type, event_type, payload, retry_count, last_error
		FROM payment.outbox
		WHERE status = 'pending'
		  AND retry_count < $2
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`,
		limit, port.MaxOutboxRetries,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending outbox messages: %w", err)
	}
	defer rows.Close()

	var messages []*port.OutboxMessage
	for rows.Next() {
		var (
			msg       port.OutboxMessage
			lastError *string
		)
		if err := rows.Scan(
			&msg.ID, &msg.AggregateID, &msg.AggregateType, &msg.EventType,
			&msg.Payload, &msg.RetryCount, &lastError,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outbox row: %w", err)
		}
		if lastError != nil {
			msg.LastError = *lastError
		}
		messages = append(messages, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("outbox row iteration error: %w", err)
	}
	return messages, nil
}

// MarkProcessed sets status='processed' and records the processed timestamp.
func (r *OutboxRepository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE payment.outbox
		SET status = 'processed', processed_at = $1
		WHERE id = $2`,
		time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to mark outbox message as processed: %w", err)
	}
	return nil
}

// MarkFailed increments retry_count and records the error message.
// When retry_count reaches MaxOutboxRetries the status transitions to 'failed';
// otherwise it stays 'pending' so the worker retries on the next poll.
func (r *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE payment.outbox
		SET
			retry_count = retry_count + 1,
			last_error  = $1,
			status      = CASE WHEN retry_count + 1 >= $2 THEN 'failed' ELSE 'pending' END
		WHERE id = $3`,
		errMsg, port.MaxOutboxRetries, id,
	)
	if err != nil {
		return fmt.Errorf("failed to mark outbox message as failed: %w", err)
	}
	return nil
}

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ademarthiago/payment-gateway/internal/domain/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

func (r *OutboxRepository) Save(ctx context.Context, msg *port.OutboxMessage) error {
	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO payment.outbox
			(id, aggregate_id, aggregate_type, event_type, payload, status, created_at)
		VALUES ($1,$2,$3,$4,$5,'pending',$6)`,
		msg.ID, msg.AggregateID, msg.AggregateType, msg.EventType, payload, time.Now().UTC(),
	)
	return err
}

func (r *OutboxRepository) FetchPending(ctx context.Context, limit int) ([]*port.OutboxMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, aggregate_id, aggregate_type, event_type, payload
		FROM payment.outbox
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending outbox: %w", err)
	}
	defer rows.Close()

	var messages []*port.OutboxMessage
	for rows.Next() {
		var (
			msg         port.OutboxMessage
			payloadJSON []byte
		)
		if err := rows.Scan(&msg.ID, &msg.AggregateID, &msg.AggregateType, &msg.EventType, &payloadJSON); err != nil {
			return nil, fmt.Errorf("failed to scan outbox row: %w", err)
		}
		msg.Payload = payloadJSON
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE payment.outbox SET status='processed', processed_at=$1 WHERE id=$2`,
		time.Now().UTC(), id,
	)
	return err
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE payment.outbox
		SET status='failed', last_error=$1, attempts=attempts+1
		WHERE id=$2`,
		errMsg, id,
	)
	return err
}

// Package event contains the in-process event infrastructure.
// It has two parts: a channel-based publisher for immediate delivery,
// and the OutboxWorker that guarantees delivery by polling the database.
package event

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ademarthiago/payment-gateway/internal/domain/port"
)

// defaultOutboxBatchSize is the maximum number of pending messages fetched per tick.
const DefaultOutboxBatchSize = 100

// OutboxWorker polls the outbox table and publishes pending events.
// It's the durability safety net — if the channel publish fails or the process crashes
// after writing to the DB but before the event was delivered, this worker picks it up.
// Events may be delivered more than once (at-least-once), so downstream handlers must be idempotent.
type OutboxWorker struct {
	outboxRepo port.OutboxRepository
	publisher  port.EventPublisher
	interval   time.Duration
	batchSize  int
}

// NewOutboxWorker creates the worker with a configurable polling interval and batch size.
// The interval is read from OUTBOX_WORKER_INTERVAL_SECONDS env at startup.
func NewOutboxWorker(
	outboxRepo port.OutboxRepository,
	publisher port.EventPublisher,
	interval time.Duration,
	batchSize int,
) *OutboxWorker {
	if batchSize <= 0 {
		batchSize = DefaultOutboxBatchSize
	}
	return &OutboxWorker{
		outboxRepo: outboxRepo,
		publisher:  publisher,
		interval:   interval,
		batchSize:  batchSize,
	}
}

// Start runs the polling loop in the calling goroutine.
// It stops cleanly when ctx is cancelled (e.g. on SIGTERM during graceful shutdown).
func (w *OutboxWorker) Start(ctx context.Context) {
	log.Info().Dur("interval", w.interval).Int("batch_size", w.batchSize).Msg("outbox worker started")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("outbox worker stopped")
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *OutboxWorker) process(ctx context.Context) {
	messages, err := w.outboxRepo.FetchPending(ctx, w.batchSize)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch pending outbox messages")
		return
	}

	for _, msg := range messages {
		if err := w.publisher.Publish(ctx, port.Event{
			Type:    msg.EventType,
			Payload: msg.Payload,
		}); err != nil {
			log.Error().
				Err(err).
				Str("event_type", msg.EventType).
				Str("aggregate_id", msg.AggregateID.String()).
				Msg("failed to publish outbox message")

			if err := w.outboxRepo.MarkFailed(ctx, msg.ID, err.Error()); err != nil {
				log.Error().
					Err(err).
					Str("message_id", msg.ID.String()).
					Msg("failed to mark outbox message as failed")
			}
			continue
		}

		if err := w.outboxRepo.MarkProcessed(ctx, msg.ID); err != nil {
			log.Error().Err(err).Str("message_id", msg.ID.String()).Msg("failed to mark outbox message as processed")
		}

		log.Debug().
			Str("event_type", msg.EventType).
			Str("message_id", msg.ID.String()).
			Msg("outbox message processed")
	}
}

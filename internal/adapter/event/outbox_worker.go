package event

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ademarthiago/payment-gateway/internal/domain/port"
)

// OutboxWorker polls the outbox table and publishes pending events
// This guarantees at-least-once delivery even if channel publish fails
type OutboxWorker struct {
	outboxRepo port.OutboxRepository
	publisher  port.EventPublisher
	interval   time.Duration
}

func NewOutboxWorker(
	outboxRepo port.OutboxRepository,
	publisher port.EventPublisher,
	interval time.Duration,
) *OutboxWorker {
	return &OutboxWorker{
		outboxRepo: outboxRepo,
		publisher:  publisher,
		interval:   interval,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	log.Info().Dur("interval", w.interval).Msg("outbox worker started")

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
	messages, err := w.outboxRepo.FetchPending(ctx, 100)
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

			_ = w.outboxRepo.MarkFailed(ctx, msg.ID, err.Error())
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

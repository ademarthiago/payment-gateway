package event

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/ademarthiago/payment-gateway/internal/domain/port"
)

// ChannelPublisher publishes events via Go channels (in-process)
// The Outbox Pattern guarantees durability if the channel publish fails
type ChannelPublisher struct {
	ch chan<- port.Event
}

func NewChannelPublisher(ch chan<- port.Event) *ChannelPublisher {
	return &ChannelPublisher{ch: ch}
}

func (p *ChannelPublisher) Publish(ctx context.Context, event port.Event) error {
	select {
	case p.ch <- event:
		log.Debug().Str("event_type", event.Type).Msg("event published to channel")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("publish cancelled: %w", ctx.Err())
	default:
		// Channel full — outbox will handle delivery
		log.Warn().Str("event_type", event.Type).Msg("channel full, outbox will retry")
		return nil
	}
}

package event

import (
	"context"

	"github.com/ademarthiago/payment-gateway/internal/domain/port"
	"github.com/rs/zerolog/log"
)

// Dispatcher reads events from channel and routes to registered handlers
type Dispatcher struct {
	ch       <-chan port.Event
	handlers map[string][]port.EventHandler
}

func NewDispatcher(ch <-chan port.Event) *Dispatcher {
	return &Dispatcher{
		ch:       ch,
		handlers: make(map[string][]port.EventHandler),
	}
}

func (d *Dispatcher) Register(eventType string, handler port.EventHandler) {
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

func (d *Dispatcher) Start(ctx context.Context) {
	log.Info().Msg("event dispatcher started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("event dispatcher stopped")
			return
		case event := <-d.ch:
			d.dispatch(ctx, event)
		}
	}
}

func (d *Dispatcher) dispatch(ctx context.Context, event port.Event) {
	handlers, ok := d.handlers[event.Type]
	if !ok {
		log.Debug().Str("event_type", event.Type).Msg("no handlers registered")
		return
	}

	for _, h := range handlers {
		if err := h.Handle(ctx, event); err != nil {
			log.Error().
				Err(err).
				Str("event_type", event.Type).
				Msg("handler failed")
		}
	}
}

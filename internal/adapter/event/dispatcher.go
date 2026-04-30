package event

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/ademarthiago/payment-gateway/internal/domain/port"
)

// Dispatcher reads events from channel and routes to registered handlers.
// Register must be called before Start. It is not safe for concurrent use.
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

// Register adds a handler for the given event type.
// Silently ignores duplicate registrations — the same handler will never run twice for the same event.
// Must be called before Start.
func (d *Dispatcher) Register(eventType string, handler port.EventHandler) {
	for _, existing := range d.handlers[eventType] {
		if existing == handler {
			return
		}
	}
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

func (d *Dispatcher) Start(ctx context.Context) {
	log.Info().Msg("event dispatcher started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("event dispatcher stopped")
			return
		case event, ok := <-d.ch:
			if !ok {
				log.Info().Msg("event channel closed, stopping dispatcher")
				return
			}
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

	var wg sync.WaitGroup
	for _, h := range handlers {
		h := h
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Handle(ctx, event); err != nil {
				log.Error().
					Err(err).
					Str("event_type", event.Type).
					Msg("handler failed")
			}
		}()
	}
	wg.Wait()
}

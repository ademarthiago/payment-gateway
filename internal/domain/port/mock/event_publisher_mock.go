package mock

import (
	"context"

	"github.com/ademarthiago/payment-gateway/internal/domain/port"
)

type EventPublisherMock struct {
	PublishFn func(ctx context.Context, event port.Event) error
	Events    []port.Event
}

func (m *EventPublisherMock) Publish(ctx context.Context, event port.Event) error {
	m.Events = append(m.Events, event)
	if m.PublishFn != nil {
		return m.PublishFn(ctx, event)
	}
	return nil
}

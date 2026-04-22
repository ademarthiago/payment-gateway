package port

import "context"

// Event represents a domain event to be published
type Event struct {
	Type    string
	Payload []byte
}

// EventPublisher defines the contract for publishing domain events
// Implemented by the channel-based adapter
type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}

// EventHandler defines the contract for handling domain events
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
}

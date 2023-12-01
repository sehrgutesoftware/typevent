// Package typevent provides type safe event channels.
package typevent

import (
	"context"
)

// Channel is an event channel that can emit and subscribe to events of a specific type.
//
// See [typevent.redis.NewChannel] for an example implementation.
type Channel[E Event] interface {
	// Emit emits an event of type E on the channel.
	Emit(E) error
	// Subscribe registers a handler for events of type E on the channel.
	Subscribe(ctx context.Context, handler Handler[E]) (Subscription, error)
}

// Event is an event that can be emitted on an event channel.
type Event any

// Handler is a function that handles an event of a specific type.
type Handler[E Event] func(context.Context, E) error

// Subscription is a subscription to an event channel.
type Subscription interface {
	// Close unsubscribes from the channel.
	Close() error
}

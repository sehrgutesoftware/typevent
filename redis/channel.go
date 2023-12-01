package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/sehrgutesoftware/typevent"
)

// channel is a redis backed implementation of [typevent.Channel].
type channel[E typevent.Event] struct {
	*Config
	event string
}

// NewChannel returns a new [typevent.Channel] backed by redis Pub/Sub as the underlying event bus.
//
// The channel will emit and subscribe to events of type E. The `name` param is a unique identifier
// for the event. All channels created with the same `name` must use the same type `E`. The package
// does intentionally not enforce this constraint, as it would require the use of reflection.
//
// Redis Pub/Sub is used as the underlying event bus. The events emitted on the channel are passed
// to all channels subscribed on the same `name` on the same redis server, regardless of the DB
// they're connected to – see [https://redis.io/docs/interact/pubsub/#database--scoping].
func NewChannel[E typevent.Event](conf *Config, event string) typevent.Channel[E] {
	return &channel[E]{
		Config: conf,
		event:  event,
	}
}

// Emit emits an event of type E on the channel.
func (c *channel[E]) Emit(event E) error {
	encoded, err := c.codec.Marshal(event)
	if err != nil {
		return err
	}

	return c.client.Publish(context.Background(), c.prefix(c.event), encoded).Err()
}

// Subscribe registers a handler for events of type E on the channel.
func (c *channel[E]) Subscribe(ctx context.Context, handler typevent.Handler[E]) (typevent.Subscription, error) {
	sub := c.client.Subscribe(ctx, c.prefix(c.event))

	// The following call is necessary to make sure the subscription is established.
	_, err := sub.Receive(ctx)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	go c.listen(ctx, sub, handler)

	return &subscription{sub: sub, cancel: cancel}, nil
}

// listen is the goroutine that listens for events on the redis channel.
func (c *channel[E]) listen(ctx context.Context, sub *redis.PubSub, handler typevent.Handler[E]) {
	defer sub.Close()

	for {
		select {
		case msg := <-sub.Channel():
			var event E
			err := c.codec.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				continue
			}
			go handler(ctx, event)
		case <-ctx.Done():
			return
		}
	}
}

type subscription struct {
	sub    *redis.PubSub
	cancel context.CancelFunc
}

// Close unsubscribes from the channel.
func (s *subscription) Close() error {
	s.cancel()
	return s.sub.Close()
}

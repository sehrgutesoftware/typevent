# typevent

typevent provides type-safe event messaging channels for Go. They can be used to implement Pub/Sub schemes without the need for type assertions, streamlining the application code that uses the event channels.

At the core it consists of a generic `Channel` interface with the following methods:

```go
type Channel[E Event] interface {
	// Emit emits an event of type E on the channel.
	Emit(E) error
	// Subscribe registers a handler for events of type E on the channel.
	Subscribe(ctx context.Context, handler Handler[E]) (Subscription, error)
}
```

The package currently provides one implementation of the interface, using [Redis Pub/Sub](https://redis.io/docs/interact/pubsub/) as the backing distribution system. Usage can look as follows:

```go
import (
	"context"
	"fmt"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/sehrgutesoftware/typevent/redis"
)

func ExampleNewChannel() {
    type event string

	// Create a new channel using redis Pub/Sub as the underlying event bus.
	client := redisclient.NewClient(&redisclient.Options{Addr: "localhost:6379"})

    // conf holds the redis client used by the channel
	conf := redis.NewConfig(client)

    // This is where we create the channel that can be used to emit and subscribe to events
	channel := redis.NewChannel[event](conf, "CHANNEL_NAME")

	// Register a subscriber for the channel.
	sub, _ := channel.Subscribe(context.Background(), func(ctx context.Context, ev event) error {
		fmt.Printf("subscriber says: %s\n", ev)
		return nil
	})
	defer sub.Close()

	// Emit an event on the channel.
	channel.Emit("Hello World!")
}

```

## Development
### Run Tests
```sh
go test ./...
```

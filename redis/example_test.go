package redis_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/alicebob/miniredis/v2"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/sehrgutesoftware/typevent/redis"
)

func ExampleNewChannel() {
	server, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer server.Close()

	type event string

	// Create a new channel using redis Pub/Sub as the underlying event bus.
	client := redisclient.NewClient(&redisclient.Options{Addr: server.Addr()})
	conf := redis.NewConfig(client)
	channel := redis.NewChannel[event](conf, "CHANNEL_NAME")

	// The WaitGroup is necessary to make the example test pass. In a real application,
	// you would likely just let the handler run until its context is canceled.
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Register a subscriber for the channel.
	sub, _ := channel.Subscribe(context.Background(), func(ctx context.Context, ev event) error {
		defer wg.Done()
		fmt.Printf("subscriber says: %s\n", ev)
		return nil
	})
	defer sub.Close()

	// Emit an event on the channel.
	channel.Emit("Hello World!")

	wg.Wait() // again, just there to statisfy the example test
	// Output: subscriber says: Hello World!
}

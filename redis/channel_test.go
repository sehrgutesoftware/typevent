package redis_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redismock/v9"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/sehrgutesoftware/typevent/redis"
)

type evType struct {
	ExportedKey string
	hiddenKey   string
}

func TestItDistributesEventsFromRedisToAllSubscribers(t *testing.T) {
	// Setup a redis server.
	s := miniredis.RunT(t)

	// Create a new channel.
	codec := &redis.JSONCodec{}
	db := redisclient.NewClient(&redisclient.Options{Addr: s.Addr()})
	config := redis.NewConfig(db, redis.WithCodec(codec), redis.WithKeyPrefix("test:"))
	channel := redis.NewChannel[evType](config, "CHANNEL_NAME")

	// Set up some syncing for the event handlers to know when they're done.
	// This is necessary because the event handlers are called in a goroutine.
	wg := sync.WaitGroup{}
	wg.Add(2)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Add a subscriber to the channel.
	timesReceivedA := 0
	receivedByA := evType{}
	sub, err := channel.Subscribe(ctx, func(ctx context.Context, e evType) error {
		receivedByA = e
		timesReceivedA++
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error returned from Subscribe(): %s", err)
	}
	defer sub.Close()

	// Add another subscriber to the channel.
	timesReceivedB := 0
	receivedByB := evType{}
	sub, err = channel.Subscribe(ctx, func(ctx context.Context, e evType) error {
		receivedByB = e
		timesReceivedB++
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error returned from Subscribe(): %s", err)
	}
	defer sub.Close()

	// Emit an event on the redis Pub/Sub channel.
	emitted := evType{
		ExportedKey: "exported",
		hiddenKey:   "hidden",
	}

	encoded, _ := codec.Marshal(emitted)
	err = db.Publish(context.Background(), "test:CHANNEL_NAME", encoded).Err()
	if err != nil {
		t.Fatalf("unexpected error returned from Publish(): %s", err)
	}

	// Wait for the waitgroup or the context timeout.
	wgDone := make(chan bool)
	go func() {
		wg.Wait()
		wgDone <- true
	}()
	select {
	case <-wgDone:
		break
	case <-ctx.Done():
		break
	}

	// Subscriber A was called correctly?
	if timesReceivedA != 1 {
		t.Errorf("expected handler A to be called once, got called %d times", timesReceivedA)
	}
	if receivedByA.ExportedKey != emitted.ExportedKey {
		t.Errorf("expected handler A to receive the submitted event, got %v", receivedByA)
	}

	// Subscriber B was called correctly?
	if timesReceivedB != 1 {
		t.Errorf("expected handler B to be called once, got called %d times", timesReceivedB)
	}
	if receivedByB.ExportedKey != emitted.ExportedKey {
		t.Errorf("expected handler B to receive the submitted event, got %v", receivedByB)
	}

	// Unexported fields are not serialized.
	if receivedByA.hiddenKey != "" {
		t.Errorf("expected handler A to receive the submitted event with unexported fields removed, got %v", receivedByB)
	}
	if receivedByB.hiddenKey != "" {
		t.Errorf("expected handler B to receive the submitted event with unexported fields removed, got %v", receivedByB)
	}
}

func TestItPublishesEventsToRedis(t *testing.T) {
	client, mock := redismock.NewClientMock()

	// Create a new channel.
	codec := &redis.JSONCodec{}
	config := redis.NewConfig(client, redis.WithCodec(codec), redis.WithKeyPrefix("test:"))
	channel := redis.NewChannel[evType](config, "CHANNEL_NAME")

	// Event to be emitted.
	emitted := evType{
		ExportedKey: "exported",
		hiddenKey:   "hidden",
	}
	encoded, _ := codec.Marshal(emitted)

	// Set up the mock expectations.
	mock.ExpectPublish("test:CHANNEL_NAME", encoded).SetVal(1)

	// Emit the event.
	err := channel.Emit(emitted)
	if err != nil {
		t.Fatalf("unexpected error returned from Emit(): %s", err)
	}

	// Check the expectations.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

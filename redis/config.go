package redis

import (
	redis "github.com/redis/go-redis/v9"
)

// config contains the configuration for a redis channel.
type config struct {
	// client is the redis client used by the channel.
	client *redis.Client
	// codec is the encoding used to serialize and deserialize events.
	codec Codec
	// keyPrefix is the prefix for all redis keys used by the channel.
	keyPrefix string
}

// NewConfig returns a new [config] for a redis channel.
func NewConfig(client *redis.Client, opts ...ConfigOption) *config {
	c := &config{
		client: client,
		codec:  defaultCodec,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// prefix prefixes the given prefix with the channel's prefix prefix.
func (c *config) prefix(key string) string {
	return c.keyPrefix + key
}

// ConfigOption is a configuration option for the redis channel.
type ConfigOption func(*config)

// WithCodec sets the codec used to serialize and deserialize events.
func WithCodec(codec Codec) ConfigOption {
	return func(c *config) {
		c.codec = codec
	}
}

// WithKeyPrefix sets the prefix for all redis keys used by the channel.
func WithKeyPrefix(prefix string) ConfigOption {
	return func(c *config) {
		c.keyPrefix = prefix
	}
}

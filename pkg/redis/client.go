package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client wraps Redis operations used by reusable components.
type Client struct {
	client *goredis.Client
}

// New creates a Redis client.
func New(config Config) *Client {
	return &Client{
		client: goredis.NewClient(&goredis.Options{
			Addr:     config.Address,
			Username: config.Username,
			Password: config.Password,
			DB:       config.Database,
		}),
	}
}

// Close closes the Redis client.
func (client *Client) Close() error {
	return client.client.Close()
}

// Delete removes a Redis key.
func (client *Client) Delete(ctx context.Context, key string) error {
	return client.client.Del(ctx, key).Err()
}

// Expire updates the expiration duration for a Redis key.
func (client *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return client.client.Expire(ctx, key, ttl).Err()
}

// Find reads a Redis key and reports whether it exists.
func (client *Client) Find(ctx context.Context, key string) ([]byte, bool, error) {
	value, err := client.client.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	return value, true, nil
}

// Set writes a Redis key with an optional expiration duration.
func (client *Client) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return client.client.Set(ctx, key, value, ttl).Err()
}

// Take reads and deletes a Redis key atomically.
func (client *Client) Take(ctx context.Context, key string) ([]byte, bool, error) {
	value, err := client.client.GetDel(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	return value, true, nil
}

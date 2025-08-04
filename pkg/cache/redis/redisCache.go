package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zhavkk/order-service/pkg/cache"
)

type Client struct {
	client *redis.Client
	log    *slog.Logger
}

func NewClient(client *redis.Client, logger *slog.Logger) (*Client, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Error("redis ping failed", slog.Any("error", err))
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}
	logger.Info("redis client initialized successfully")
	return &Client{
		client: client,
		log:    logger,
	}, nil
}

func (c *Client) Get(ctx context.Context, key string, destination any) error {
	const op = "rediscache.Client.Get"
	start := time.Now()

	val, err := c.client.Get(ctx, key).Result()

	dur := time.Since(start)

	c.log.Info(op, "key", slog.String("key", key), "duration", dur)
	if errors.Is(err, redis.Nil) {
		c.log.Warn("cache miss", slog.String("key", key))
		return cache.ErrCacheMiss
	}

	if err != nil {
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	return json.Unmarshal([]byte(val), destination)

}

func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	const op = "rediscache.Client.Set"
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	start := time.Now()
	err = c.client.Set(ctx, key, data, ttl).Err()

	dur := time.Since(start)
	c.log.Info(op, "key", slog.String("key", key), "duration", dur, "ttl", ttl)

	return err
}

func (c *Client) Delete(ctx context.Context, key string) error {
	const op = "rediscache.Client.Delete"
	start := time.Now()
	err := c.client.Del(ctx, key).Err()

	dur := time.Since(start)
	c.log.Info(op, "key", slog.String("key", key), "duration", dur)
	return err
}

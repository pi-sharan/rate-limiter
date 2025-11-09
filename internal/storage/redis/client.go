package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps go-redis with localized settings.
type Client struct {
	conn *redis.Client
}

// Config provides Redis connection details.
type Config struct {
	Addr     string
	Username string
	Password string
	DB       int
}

// New creates a Redis client.
func New(cfg Config) (*Client, error) {
	conn := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := ping(context.Background(), conn); err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func ping(ctx context.Context, conn *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := conn.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// Raw exposes the underlying redis.Client when needed.
func (c *Client) Raw() *redis.Client {
	return c.conn
}

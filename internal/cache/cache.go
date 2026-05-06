package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func InitClientRedis(ctx context.Context, addr string) (*redis.Client, error) {
	client := redis.NewClient(
		&redis.Options{
			Addr: addr,
		})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{
		client: client,
	}
}
func makeKey(targetID int64) string {
	return fmt.Sprintf("target:%d:status", targetID)
}

func (c *Cache) SaveStatus(ctx context.Context, targetID int64, status string) error {
	key := makeKey(targetID)
	return c.client.Set(ctx, key, status, 0).Err()
}

func (c *Cache) GetLastStatus(ctx context.Context, id int64) (string, error) {
	key := makeKey(id)
	status, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return status, nil
}

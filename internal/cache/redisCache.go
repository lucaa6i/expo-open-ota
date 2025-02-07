package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client   *redis.Client
	host     string
	password string
	port     string
}

func NewRedisCache(host, password, port string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := client.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	return &RedisCache{client: client}
}

func (c *RedisCache) Get(key string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return ""
	} else if err != nil {
		return ""
	}
	return val
}

func (c *RedisCache) Set(key string, value string, ttl *int) error {
	expiration := time.Duration(0)
	if ttl != nil {
		expiration = time.Duration(*ttl) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *RedisCache) Delete(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c.client.Del(ctx, key)
}

func (c *RedisCache) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.client.FlushDB(ctx).Err()
}

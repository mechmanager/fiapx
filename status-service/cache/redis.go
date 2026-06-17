// Package cache implementa o cache de status de vídeos usando Redis.
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implementa domain.StatusCache usando Redis.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache cria o cliente Redis a partir da URL fornecida.
func NewRedisCache(url string) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	return &RedisCache{client: client}, nil
}

// Get retorna o valor armazenado para a chave. Retorna "" e nil em caso de cache miss.
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set armazena o valor com o TTL informado.
func (c *RedisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type redisCache struct {
	client *redis.Client
	ctx    context.Context
}

type RedisCacheInterface interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (interface{}, error)
	Delete(key string) error
	HGet(key string, field string) (interface{}, error)
	HGetAll(key string) (map[string]string, error)
	HSet(key string, field string, value interface{}) error
	HDel(key string, field string) error
}

func NewRedisCache(address string) RedisCacheInterface {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("redis ping err: %v", err)
	}

	return &redisCache{
		client: client,
		ctx:    ctx,
	}
}

func (r *redisCache) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

func (r *redisCache) Get(key string) (interface{}, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *redisCache) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *redisCache) HSet(key string, field string, value interface{}) error {
	return r.client.HSet(r.ctx, key, field, value).Err()
}

func (r *redisCache) HGet(key string, field string) (interface{}, error) {
	return r.client.HGet(r.ctx, key, field).Result()
}

func (r *redisCache) HGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(r.ctx, key).Result()
}

func (r *redisCache) HDel(key string, field string) error {
	return r.client.HDel(r.ctx, key, field).Err()
}

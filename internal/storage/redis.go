package storage

import (
	"context"
	"errors"
	"strconv"
	"time"

	"watchtower/internal/checker"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string, password string) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
	return &RedisStore{client: client}
}

func (r *RedisStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisStore) Client() *redis.Client {
	return r.client
}

func (r *RedisStore) IncrFailures(ctx context.Context, target string) (int64, error) {
	key := "watchtower:failures:" + target
	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 24*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incrCmd.Val(), nil
}

func (r *RedisStore) ResetFailures(ctx context.Context, target string) error {
	key := "watchtower:failures:" + target
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStore) GetFailures(ctx context.Context, target string) (int64, error) {
	key := "watchtower:failures:" + target
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

func (r *RedisStore) SetCooldown(ctx context.Context, key string, ttl time.Duration) error {
	redisKey := "watchtower:cooldown:" + key
	return r.client.Set(ctx, redisKey, "1", ttl).Err()
}

func (r *RedisStore) IsOnCooldown(ctx context.Context, key string) (bool, error) {
	redisKey := "watchtower:cooldown:" + key
	exists, err := r.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *RedisStore) SetLastStatus(ctx context.Context, target string, status checker.Status) error {
	key := "watchtower:status:" + target
	return r.client.Set(ctx, key, string(status), 7*24*time.Hour).Err()
}

func (r *RedisStore) GetLastStatus(ctx context.Context, target string) (checker.Status, error) {
	key := "watchtower:status:" + target
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return checker.Status(val), nil
}

func (r *RedisStore) Close() error {
	return r.client.Close()
}


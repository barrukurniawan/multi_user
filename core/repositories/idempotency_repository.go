package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/barrukurniawan/multi_user/core/models"
	"github.com/redis/go-redis/v9"
)

type IdempotencyRepository interface {
	Find(ctx context.Context, key string) (*models.IdempotencyKey, error)
	Reserve(ctx context.Context, key, userID string, ttl time.Duration) (bool, error)
	SaveResponse(ctx context.Context, key, userID string, response []byte, statusCode int) error
	Release(ctx context.Context, key string) error
}

type idempotencyRepository struct {
	Redis *redis.Client
}

func NewIdempotencyRepository(rdb *redis.Client) IdempotencyRepository {
	return &idempotencyRepository{Redis: rdb}
}

func redisKey(key string) string {
	return "idempotency:" + key
}

func (r *idempotencyRepository) Find(ctx context.Context, key string) (*models.IdempotencyKey, error) {
	data, err := r.Redis.Get(ctx, redisKey(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var k models.IdempotencyKey
	if err := json.Unmarshal(data, &k); err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *idempotencyRepository) Reserve(ctx context.Context, key, userID string, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(models.IdempotencyKey{UserID: userID})
	if err != nil {
		return false, err
	}
	return r.Redis.SetNX(ctx, redisKey(key), data, ttl).Result()
}

func (r *idempotencyRepository) SaveResponse(ctx context.Context, key, userID string, response []byte, statusCode int) error {
	data, err := json.Marshal(models.IdempotencyKey{UserID: userID, StatusCode: statusCode, Response: response})
	if err != nil {
		return err
	}
	return r.Redis.SetArgs(ctx, redisKey(key), data, redis.SetArgs{KeepTTL: true}).Err()
}

func (r *idempotencyRepository) Release(ctx context.Context, key string) error {
	return r.Redis.Del(ctx, redisKey(key)).Err()
}

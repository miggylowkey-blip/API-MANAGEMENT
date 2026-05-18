package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client *redis.Client
	rps    int
	window time.Duration
}

func NewRedisLimiter(client *redis.Client, rps int) *RedisLimiter {
	if rps <= 0 {
		rps = 5

	}
	return &RedisLimiter{
		client: client,
		rps:    rps,
		window: time.Second,
	}
}

func (r *RedisLimiter) Allow(key string) bool {
	ctx := context.Background()
	rediskey := "ratelimit" + key

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, rediskey)
	pipe.Expire(ctx, rediskey, r.window)
	_, err := pipe.Exec(ctx)
	if err != nil {

		return true
	}
	return incr.Val() <= int64(r.rps)
}

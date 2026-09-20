package dao

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

const TIMEOUT = 2 * time.Second

type RedisGroup struct{}

func (redisGroup *RedisGroup) SetKey(key string, data interface{}, timeout time.Duration) error {
	if timeout <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	if err := global.LNF_RDB.Set(ctx, key, data, timeout).Err(); err != nil {
		return err
	}

	return nil
}

func (redisGroup *RedisGroup) GetValueInt64(key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	val, err := global.LNF_RDB.Get(ctx, key).Int64()
	if err != nil {
		return -1, err
	}
	return val, nil
}

func (redisGroup *RedisGroup) GetValueString(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	val, err := global.LNF_RDB.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

func (redisGroup *RedisGroup) KeyExists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	val, err := global.LNF_RDB.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val != 0, nil
}

func (redisGroup *RedisGroup) DelKey(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	if err := global.LNF_RDB.Del(ctx, key).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	return nil
}

func (redisGroup *RedisGroup) INCR(key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	value, err := global.LNF_RDB.Incr(ctx, key).Result()
	if err != nil {
		return -1, err
	}
	return value, nil
}

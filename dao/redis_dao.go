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

func (redisGroup *RedisGroup) PipeSetKey(kv map[string]interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	if len(kv) == 0 {
		return nil
	}
	pipe := global.LNF_RDB.Pipeline()
	for key, value := range kv {
		pipe.Set(ctx, key, value, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (redisGroup *RedisGroup) PipeGetString(keys []string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	if len(keys) == 0 {
		return nil, nil
	}
	pipe := global.LNF_RDB.Pipeline()
	for _, key := range keys {
		pipe.Get(ctx, key)
	}
	result, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	resMap := make(map[string]interface{}, len(result))
	for i, cmd := range result {
		if i >= len(keys) {
			break
		}
		stringCmd, ok := cmd.(*redis.StringCmd)
		if !ok {
			continue
		}
		val, err := stringCmd.Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				resMap[keys[i]] = nil
			}
			return nil, err
		}
		resMap[keys[i]] = val
	}

	return resMap, nil
}

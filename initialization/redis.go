package initialization

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

func InitRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     global.LNF_CONFIG.Redis.Host + ":" + strconv.Itoa(global.LNF_CONFIG.Redis.Port),
		Password: global.LNF_CONFIG.Redis.Password,
		DB:       global.LNF_CONFIG.Redis.Database,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

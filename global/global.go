package global

import (
	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/unicornfairy864/LNF-SERVER/config"
	"gorm.io/gorm"
)

var (
	// 项目配置
	LNF_VP     *viper.Viper
	LNF_CONFIG config.Config

	// 数据库
	LNF_DB  *gorm.DB
	LNF_RDB *redis.Client

	// Http Connection
	LNF_Resty *resty.Client
)

package global

import (
	"github.com/spf13/viper"
	"github.com/unicornfairy864/LNF-SERVER/config"
	"github.com/unicornfairy864/LNF-SERVER/router"
	"github.com/unicornfairy864/LNF-SERVER/service"
	"gorm.io/gorm"
)

var (
	// 项目配置
	LNF_VP *viper.Viper
	LNF_CONFIG config.Config
	
	// 数据库
	LNF_DB *gorm.DB

	// 路由
	LNF_ROUTER router.Router
	// 控制层
	// LNF_CONTROLLER controller.Controller
	// 业务层
	LNF_SERVICE service.Service
)
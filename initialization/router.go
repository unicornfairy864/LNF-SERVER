package initialization

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/global"
	"github.com/unicornfairy864/LNF-SERVER/router"
)

func InitRouter() (r *gin.Engine) {
	// 初始化路由
	r = gin.Default()
	r.Group(global.LNF_CONFIG.Server.RouterPrefix)
	{
		// Basic
		router.BasicRouter.User.UserRouter(r)
		// Advanced
	}
	return
}
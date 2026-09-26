package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type UploadRouter struct{}

func (u *UploadRouter) CreateRouter(api *gin.RouterGroup) {
	// Private：上传必须登录（限流排期见 agent.md/image-upload-plan.md 后续迭代）
	private := api.Group("/upload")
	private.Use(middleware.JWTAuthMiddleware())
	{
		private.POST("/image", handler.UploadHandler.UploadImageHandler)
	}
}

package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type AnnouncementRouter struct{}

func (a *AnnouncementRouter) CreateRouter(api *gin.RouterGroup) {
	userGroup := api.Group("")

	// 公开：列表 + 详情（静态 "" 与参数 /:id 同层共存，路径与契约一致，无尾斜杠）
	public := userGroup.Group("/announcement")
	{
		public.GET("", handler.AnnouncementHandler.GetHandler)
		public.GET("/:id", handler.AnnouncementHandler.GetDetailHandler)
	}

	// 管理（JWT + role=2）：四个端点路径保持不变，GET 列表改 query 传参
	admin := userGroup.Group("/admin/announcement")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.SystemAdminAuthMiddleware())
	{
		admin.POST("/create", handler.AnnouncementHandler.CreateHandler)
		admin.POST("/update", handler.AnnouncementHandler.UpdatedHandler)
		admin.DELETE("/:id", handler.AnnouncementHandler.DeleteHandler)
		admin.GET("", handler.AnnouncementHandler.AuthGetHandler)
	}
}

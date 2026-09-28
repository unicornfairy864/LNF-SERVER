package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type AnnouncementRouter struct{}

func (a *AnnouncementRouter) CreateRouter(api *gin.RouterGroup) {
	userGroup := api.Group("")

	public := userGroup.Group("/announcement")
	{
		public.GET("/", handler.AnnouncementHandler.GetHandler)
	}

	admin := userGroup.Group("/admin/announcement")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.SystemAdminAuthMiddleware())
	{
		admin.POST("/create", handler.AnnouncementHandler.CreateHandler)
		admin.POST("/update", handler.AnnouncementHandler.UpdatedHandler)
		admin.DELETE("/:id", handler.AnnouncementHandler.DeleteHandler)
		admin.GET("/", handler.AnnouncementHandler.AuthGetHandler)
	}
}

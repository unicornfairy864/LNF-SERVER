package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type CommentRouter struct{}

func (l *CommentRouter) CreateRouter(api *gin.RouterGroup) {
	userGroup := api.Group("")
	//public
	public := userGroup.Group("")
	{
		public.GET("/item/:item_id/comments", handler.CommentHandler.GetListHandler)
		public.GET("/item/:item_id/comments/replies", handler.CommentHandler.GetChildHandler)
	}
	//private
	private := userGroup.Group("")
	private.Use(middleware.JWTAuthMiddleware())
	{
		private.POST("/item/:item_id/comments/create", handler.CommentHandler.CreateHandler)
	}
	//admin
	admin := userGroup.Group("")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.ServiceAdminAuthMiddleware())
	{
		admin.PATCH("/item/:item_id/comments/update", handler.CommentHandler.UpdateHandler)
	}
}

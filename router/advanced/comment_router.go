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
		public.GET("/item/:itemID/comments", handler.CommentHandler.GetListHandler)
		public.GET("/item/:itemID/comments/replies", handler.CommentHandler.GetChildHandler)
	}
	//private
	private := userGroup.Group("")
	private.Use(middleware.JWTAuthMiddleware())
	{
		private.POST("/item/:itemID/comments/create", handler.CommentHandler.CreateHandler)
	}
	//admin
	admin := userGroup.Group("")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.ServiceAdminAuthMiddleware())
	{
		admin.PATCH("/item/:itemID/comments/update", handler.CommentHandler.UpdateHandler)
	}
}

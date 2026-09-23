package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type TagRouter struct{}

func (t *TagRouter) CreateRouter(api *gin.RouterGroup) {
	userGroup := api.Group("")
	// Public
	public := userGroup.Group("/tag")
	{
		public.GET("/list", handler.TagHandler.ListTagHandler)
	}
	// Private
	private := userGroup.Group("/tag")
	private.Use(middleware.JWTAuthMiddleware())
	{
	}
	// Admin
	admin := userGroup.Group("/tag")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.ServiceAdminAuthMiddleware())
	{
		admin.POST("/create", handler.TagHandler.CreateTagHandler)
		admin.POST("/update", handler.TagHandler.UpdateTagHandler)
		admin.DELETE("/:tagID", handler.TagHandler.DeleteTagHandler)
	}
}

package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type LocationRouter struct{}

func (l *LocationRouter) CreateRouter(api *gin.RouterGroup) {
	const ItemPrefix = "/item/{itemID}"

	userGroup := api.Group("")
	// Public
	public := userGroup.Group("")
	{
		public.POST(ItemPrefix + "/location")
	}
	// Private
	private := userGroup.Group("")
	private.Use(middleware.JWTAuthMiddleware())
	{
	}
	// Admin
	admin := userGroup.Group("")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.ServiceAdminAuthMiddleware())
	{
		admin.POST("/create", handler.LocationHandler.CreateLocationHandler)
	}
}

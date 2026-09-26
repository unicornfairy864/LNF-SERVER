package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type LocationRouter struct{}

func (l *LocationRouter) CreateRouter(api *gin.RouterGroup) {
	// gin 仅支持 :param 语法；参数名必须与 item 路由保持一致
	const ItemPrefix = "/item/:itemID"

	userGroup := api.Group("")
	// Public
	public := userGroup.Group("")
	{
		public.GET("/location/list", handler.LocationHandler.ListLocationHandler)
		public.GET(ItemPrefix + "/locations", handler.LocationHandler.GetItemLocationsHandler)
	}
	// Private
	private := userGroup.Group("")
	private.Use(middleware.JWTAuthMiddleware())
	{
	}
	// Admin
	admin := userGroup.Group("/location")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.ServiceAdminAuthMiddleware())
	{
		admin.POST("/create", handler.LocationHandler.CreateLocationHandler)
		admin.POST("/update", handler.LocationHandler.UpdateLocationHandler)
		admin.DELETE("/:locationID", handler.LocationHandler.DeleteLocationHandler)
	}
}

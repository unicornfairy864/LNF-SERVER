package advanced

import (
	"github.com/gin-gonic/gin"
	handlerAdvanced "github.com/unicornfairy864/LNF-SERVER/handler/advanced"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type AdminStatsRouter struct{}

func (r *AdminStatsRouter) CreateRouter(api *gin.RouterGroup) {
	admin := api.Group("/admin/stats")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.ServiceAdminAuthMiddleware())
	admin.GET("/overview", (&handlerAdvanced.AdminStatsHandlerGroup{}).Overview)
	admin.GET("/distribution", (&handlerAdvanced.AdminStatsHandlerGroup{}).Distribution)
	admin.GET("/funnel", (&handlerAdvanced.AdminStatsHandlerGroup{}).Funnel)
	admin.GET("/trend", (&handlerAdvanced.AdminStatsHandlerGroup{}).Trend)
	admin.GET("/locations", (&handlerAdvanced.AdminStatsHandlerGroup{}).Locations)
	admin.GET("/time-heatmap", (&handlerAdvanced.AdminStatsHandlerGroup{}).TimeHeatmap)
	admin.GET("/items/stagnant", (&handlerAdvanced.AdminStatsHandlerGroup{}).Stagnant)
	admin.GET("/items/high-view", (&handlerAdvanced.AdminStatsHandlerGroup{}).HighView)
	admin.GET("/return-duration", (&handlerAdvanced.AdminStatsHandlerGroup{}).ReturnDuration)
}

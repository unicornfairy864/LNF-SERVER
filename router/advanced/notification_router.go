package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type NotificationRouter struct{}

// CreateRouter 通知路由（用户侧 + 管理侧发送）
func (r *NotificationRouter) CreateRouter(Router *gin.RouterGroup) {
	notification := Router.Group("")
	// Private：用户侧通知（JWT 登录态）
	private := notification.Group("/notifications")
	private.Use(middleware.JWTAuthMiddleware())
	{
		private.GET("", handler.NotificationHandler.List)
		private.GET("unread-count", handler.NotificationHandler.UnreadCount)
		private.GET(":id", handler.NotificationHandler.GetAndRead)
		private.PUT("read", handler.NotificationHandler.BatchMarkRead)
		private.DELETE("", handler.NotificationHandler.BatchDelete)
	}
	// Admin：管理侧发送（JWT + 服务管理员 role≥1，路径与 handler swagger 注解一致）
	admin := notification.Group("/admin/notifications")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.ServiceAdminAuthMiddleware())
	{
		admin.POST("", handler.NotificationHandler.Send)
	}
}

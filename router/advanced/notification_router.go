package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
)

type NotificationRouter struct{}

// InitNotificationRouter 用户侧通知路由
func (r *NotificationRouter) CreateRouter(Router *gin.RouterGroup) {
	notification := Router.Group("notifications")
	//private
	private := notification.Group("")
	{
		private.GET("", handler.NotificationHandler.List)
		private.GET("unread-count", handler.NotificationHandler.UnreadCount)
		private.GET(":id", handler.NotificationHandler.GetAndRead)
		private.PUT("read", handler.NotificationHandler.BatchMarkRead)
		private.DELETE("", handler.NotificationHandler.BatchDelete)

	}
	//admin
	notification.POST("", handler.NotificationHandler.Send)
}

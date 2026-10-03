package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type AgentRouter struct{}

// CreateRouter Agent 对话助手路由（普通用户 role=0，全部需要登录）
func (a *AgentRouter) CreateRouter(api *gin.RouterGroup) {
	group := api.Group("/agent")
	group.Use(middleware.JWTAuthMiddleware())
	{
		group.POST("/chat", handler.AgentHandler.AgentChatHandler)
		group.POST("/match", handler.AgentHandler.AgentMatchHandler)
		group.POST("/extract", handler.AgentHandler.AgentExtractHandler)
		group.POST("/session/close", handler.AgentHandler.AgentSessionCloseHandler)
	}
}

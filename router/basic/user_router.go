package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type UserRouter struct{}

func (userRouter *UserRouter) CreateRouter(api *gin.RouterGroup) {
	// public
	public := api.Group("/user")
	{
		public.POST("/create", handler.UserHandler.CreateUserHandler)
		public.POST("/login", handler.UserHandler.LoginHandler)
	}
	// private
	private := api.Use(middleware.JWTAuthMiddleware())
	{
		public := api.Group("/user")
		{
			public.POST("/logout", handler.UserHandler.LogoutHandler)
		}
		// 测试 JWT 状态
		private.GET("/jwt-test", func(c *gin.Context) {
			response.TestWithData(c, struct {
				JwtID      string `json:"jwtID"`
				JwtRole    string `json:"jwtRole"`
				TokenFresh bool   `json:"tokenFresh"`
			}{
				JwtID:      c.GetString("jwt:id"),
				JwtRole:    c.GetString("jwt:role"),
				TokenFresh: false,
			})
		})
	}
}

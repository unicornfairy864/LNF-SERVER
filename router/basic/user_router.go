package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type UserRouter struct{}

func (userRouter *UserRouter) CreateRouter(api *gin.RouterGroup) {
	userGroup := api.Group("")
	// Public
	public := userGroup.Group("/user")
	{
		public.POST("/create", handler.UserHandler.CreateUserHandler)
		public.POST("/login", handler.UserHandler.LoginHandler)

		public.POST("/batch", handler.UserHandler.BatchHandler)
	}
	// Private
	private := userGroup.Group("/user")
	private.Use(middleware.JWTAuthMiddleware())
	{
		private.POST("/logout", handler.UserHandler.LogoutHandler)
		private.POST("/update", handler.UserHandler.UpdateHandler)
		public.GET("/me", handler.UserHandler.GetMeHandler)
		public.POST("/me", handler.UserHandler.GetMeHandler)

		private.POST("/qq/get-code", handler.UserHandler.QQGetCodeHandler)
		private.POST("/qq/bind", handler.UserHandler.QQBindHandler)

		// 测试 JWT 状态
		private.GET("/jwt-test", func(c *gin.Context) {
			response.TestWithData(c, struct {
				JwtID      int64 `json:"jwtID"`
				JwtRole    int8  `json:"jwtRole"`
				TokenFresh bool  `json:"tokenFresh"`
			}{
				JwtID:      c.GetInt64("jwt:id"),
				JwtRole:    c.GetInt8("jwt:role"),
				TokenFresh: false,
			})
		})
	}

	// Admin
	admin := userGroup.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.SystemAdminAuthMiddleware())
	{
		admin.POST("/change-role", handler.UserHandler.ChangeUserRoleHandler)
		admin.POST("/change-status", handler.UserHandler.ChangeUserStatusHandler)
		admin.POST("/add-credit", handler.UserHandler.AddUserCreditHandler)
	}
}

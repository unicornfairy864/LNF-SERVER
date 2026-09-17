package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
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
	// private := api.Use()
	// {
	// 	private.POST("/",)
	// }
}
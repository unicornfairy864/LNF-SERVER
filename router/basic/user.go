package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type UserRouter struct{}

func (userRouter *UserRouter) CreateRouter(api *gin.RouterGroup) {
	// public
	public := api.Group("/user") 
	{
		public.POST("/create", service.UserService.Create)
	}
	// private
	// private := api.Use()
	// {
	// 	private.POST("/",)
	// }
}
package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type UserRouter struct{}

func (userRouter *UserRouter) UserRouter(r *gin.Engine) {
	r.Group("/user") 
	{
		r.POST("/create", service.UserService.Create)
	}
}
package basic

import (
	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type UserHandlerGroup struct{}

func (userHandler *UserHandlerGroup) CreateUserHandler(c *gin.Context) {
	req := model.CreateUserRequest{}
	c.ShouldBindBodyWithJSON(&req)
	user, errCode := service.UserService.Create(&req)
	if (user != nil) {
		response.SuccessWithData(c, user)
		return
	}
	response.FailWithCode(c, errCode)
}
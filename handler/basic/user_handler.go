package basic

import (
	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type UserHandlerGroup struct{}

// Create 创建用户
// @Summary      创建用户
// @Description  创建新用户。用户名和昵称长度 2-32，密码长度 8-20，用户名必须唯一。
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request  body      model.CreateUserRequest  true  "创建用户请求体"
// @Success      200      {object}  response.CommonResponse{data=model.UserResponse}
// @Router       /api/v1/user [post]
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

func (userHandler *UserHandlerGroup) LoginHandler(c *gin.Context) {
	req := model.LoginRequest{}
	c.ShouldBindBodyWithJSON(&req)
	user, token, errCode := service.UserService.Login(&req)
	if (user != nil) {
		c.Header("Authorization", *token)
		response.SuccessWithData(c, user)
		return
	}
	response.FailWithCode(c, errCode)
}
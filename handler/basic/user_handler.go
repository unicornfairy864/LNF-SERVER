package basic

import (
	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type UserHandlerGroup struct{}

// CreateUserHandler 创建用户
// @Summary      创建用户
// @Description  创建新用户。用户名和昵称长度 2-32，密码长度 8-20，用户名必须唯一。
// @Tags         user
// @Accept       JSON
// @Produce      JSON
// @Param        request  body      model.CreateUserRequest  true  "创建用户请求体"
// @Success      200      {object}  response.CommonResponse{data=model.UserResponse}
// @Router       /api/v1/user/create [post]
func (userHandler *UserHandlerGroup) CreateUserHandler(c *gin.Context) {
	req := model.CreateUserRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	user, errCode := service.UserService.Create(&req)
	if user != nil {
		response.SuccessWithData(c, user)
		return
	}
	response.FailWithCode(c, errCode)
}

// LoginHandler  用户登录
// @Summary      用户登录
// @Description  用户使用用户名和密码登录，验证通过后返回用户信息并在响应体 Header 设置 Authorization : Token
// @Tags         user
// @Accept       JSON
// @Produce      JSON
// @Param        request  body      model.LoginRequest  true  "用户登录请求体"
// @Success      200      {object}  response.CommonResponse{data=model.UserResponse}
// @Router       /api/v1/user/login [post]
func (userHandler *UserHandlerGroup) LoginHandler(c *gin.Context) {
	req := model.LoginRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	user, token, errCode := service.UserService.Login(&req)
	if user != nil {
		c.Header("Authorization", "Bearer "+*token)
		response.SuccessWithData(c, user)
		return
	}
	response.FailWithCode(c, errCode)
}

// LogoutHandler  用户登出
// @Summary      用户登出
// @Description  选择全部登出或仅当前会话登出
// @Tags         user
// @Accept       JSON
// @Produce      JSON
// @Param        request  body      model.LogoutRequest  true  "用户登出请求体"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/logout [post]
func (userHandler *UserHandlerGroup) LogoutHandler(c *gin.Context) {
	req := model.LogoutRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	var code response.Code
	if req.LogoutAll == 1 {
		code = service.UserService.LogoutAll(c.GetInt64("jwt:id"))
	} else {
		code = service.UserService.Logout(c.GetString("jwt:jti"), c.GetTime("jwt:expired_at"))
	}
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
	}
	response.Success(c)
}

// UpdateHandler  用户更改信息
// @Summary      用户更改信息
// @Description  用户更改昵称、真名、性别、头像
// @Tags         user
// @Accept       JSON
// @Produce      JSON
// @Param        request  body      model.UpdateUserRequest  true  "用户登出请求体"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/update [post]
func (userHandler *UserHandlerGroup) UpdateHandler(c *gin.Context) {
	req := model.UpdateUserRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); (err != nil || req == model.UpdateUserRequest{}) {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.UserService.Update(c.GetInt64("jwt:id"), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// QQGetCodeHandler  用户申请获取qq验证码
// @Summary      用户申请获取qq验证码
// @Description  后端生成验证码并让陈松发送
// @Tags         user
// @Accept       JSON
// @Produce      JSON
// @Param        request  body      model.QQGetCodeRequest  true  "获取qq验证码请求"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/update [post]
func (userHandler *UserHandlerGroup) QQGetCodeHandler(c *gin.Context) {
	req := model.QQGetCodeRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.UserService.QQGetCode(req.QQ, c.GetString("jwt:nickname"), c.GetString("jwt:jti"))
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
	}
	response.Success(c)
}

// QQBindHandler  用户绑定qq
// @Summary      用户通过验证码绑定qq
// @Description  用户在统一jti会话中验证验证码，最大次数不超过?次/?时间
// @Tags         user
// @Accept       JSON
// @Produce      JSON
// @Param        request  body      model.QQBindRequest  true  "绑定QQ请求"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/update [post]
func (userHandler *UserHandlerGroup) QQBindHandler(c *gin.Context) {
	req := model.QQBindRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	response.Success(c)
}

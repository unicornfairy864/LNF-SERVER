package basic

import (
	"strconv"

	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
	"github.com/unicornfairy864/LNF-SERVER/utils"
)

type UserHandlerGroup struct{}

// CreateUserHandler 创建用户
// @Summary      创建用户
// @Description  创建新用户。用户名和昵称长度 2-32，密码长度 8-20，用户名必须唯一。
// @Tags         user
// @Accept       json
// @Produce      json
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
// @Accept       json
// @Produce      json
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

// GetListHandler  根据id获取用户列表
// @Summary      根据id获取用户列表
// @Description  接收id数组，返回PublicUserResponse数组 <br /> 注意：结果可能不是输入列表的顺序且会过滤无效id
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request  body      model.GetListRequest  true  "获取用户列表请求体"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/get-list [post]
func (userHandler *UserHandlerGroup) GetListHandler(c *gin.Context) {
	req := model.GetListRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	if len(req.IDs) == 0 {
		response.SuccessWithData(c, []model.PublicUserResponse{})
		return
	}
	pubRes, code := service.UserService.GetListService(req.IDs)
	utils.LogJson(pubRes)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, pubRes)
}

// LogoutHandler  用户登出
// @Summary      用户登出
// @Description  选择全部登出或仅当前会话登出
// @Tags         user
// @Accept       json
// @Produce      json
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
		return
	}
	response.Success(c)
}

// UpdateHandler  用户更改信息
// @Summary      用户更改信息
// @Description  用户更改昵称、真名、性别、头像
// @Tags         user
// @Accept       json
// @Produce      json
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
// @Accept       json
// @Produce      json
// @Param        request  body      model.QQGetCodeRequest  true  "获取qq验证码请求"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/qq/get-code [post]
func (userHandler *UserHandlerGroup) QQGetCodeHandler(c *gin.Context) {
	req := model.QQGetCodeRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil ||
		!(req.QQ >= 10000 && req.QQ <= 99999999999) {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.UserService.QQGetCode(strconv.FormatInt(req.QQ, 10), c.GetString("jwt:nickname"), c.GetString("jwt:jti"))
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// QQBindHandler  用户绑定qq
// @Summary      用户通过验证码绑定qq
// @Description  用户在统一jti会话中验证验证码，最大次数不超过5次/3分钟 <br />接收json为number，但后端实际操作统一用string
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request  body      model.QQBindRequest  true  "绑定QQ请求"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/qq/bind [post]
func (userHandler *UserHandlerGroup) QQBindHandler(c *gin.Context) {
	req := model.QQBindRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil ||
		!(req.QQ >= 10000 && req.QQ <= 99999999999 || !(req.Code >= 100000 && req.Code <= 999999)) {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.UserService.QQBind(c.GetInt64("jwt:id"), c.GetString("jwt:jti"), strconv.FormatInt(req.QQ, 10), strconv.FormatInt(req.Code, 10))
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// ChangeUserRoleHandler  系统管理员改变用户角色
// @Summary      系统管理员改变用户角色
// @Description  Role为2的用户改变任意用户Role为0/1/2
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request  body      model.ChangeUserRoleRequest  true  "变更用户角色请求"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/admin-change-role [post]
func (userHandler *UserHandlerGroup) ChangeUserRoleHandler(c *gin.Context) {
	req := model.ChangeUserRoleRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.UserService.ChangeUserRoleService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// ChangeUserStatusHandler  系统管理员改变用户状态(禁用)
// @Summary      系统管理员改变用户状态(禁用)
// @Description  Role为2的用户改变任意用户Status为0(禁用)/1(正常)
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request  body      model.ChangeUserStatusRequest  true  "变更用户角色请求"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/admin-change-status [post]
func (userHandler *UserHandlerGroup) ChangeUserStatusHandler(c *gin.Context) {
	req := model.ChangeUserStatusRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.UserService.ChangeUserStatusRequest(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// AddUserCreditHandler  系统管理员改变用户积分
// @Summary      系统管理员改变用户积分
// @Description  Role为2的用户改变任意用户积分，可选为SET或ADD (is_delta)
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request  body      model.AddUserCreditRequest  true  "变更用户角色请求"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/user/admin-add-credit [post]
func (userHandler *UserHandlerGroup) AddUserCreditHandler(c *gin.Context) {
	req := model.AddUserCreditRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.UserService.ChangeUserCreditRequest(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

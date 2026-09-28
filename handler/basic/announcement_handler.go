package basic

import (
	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type AnnouncementHandlerGroup struct{}

// AnnouncementcreateHandler  创建公告
// @Summary      创建公告
// @Description  系统管理员创建公告（草稿）
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.LoginRequest  true  "创建公告请求体"
// @Success      200      {object}  response.CommonResponse{data=model.UserResponse}
// @Router       /api/v1/admin/announcement/create [post]
func (a *AnnouncementHandlerGroup) CreateHandler(c *gin.Context) {
	rep := model.AnnouncementUpdateRequest{}
	if err := c.ShouldBindJSON(&rep); err != nil {
		response.FailWithCode(c, response.CodeParamError)
	}
	err := service.AnnouncementService.CreateAnnouncement(&rep)
	if err != response.CodeSuccess {
		response.FailWithCode(c, err)
	}
	response.Success(c)
}

// AnnouncementUpdatedHandler  更新公告
// @Summary      更新公告
// @Description  系统管理员保存，发布，删除，下架公告(更新阅读数还未完成，请先忽略)
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.LoginRequest  true  "更新公告请求体"
// @Success      200      {object}  response.CommonResponse{data=model.UserResponse}
// @Router       /api/v1/admin/announcement/update [post]
func (a *AnnouncementHandlerGroup) UpdatedHandler(c *gin.Context) {
	rep := model.AnnouncementUpdateRequest{}
	if err := c.ShouldBindJSON(&rep); err != nil {
		response.FailWithCode(c, response.CodeParamError)
	}
	err := service.AnnouncementService.UpdatedAnnouncement(&rep)
	if err != response.CodeSuccess {
		response.FailWithCode(c, err)
	}
	response.Success(c)
}

// AnnouncementGetHandler  获取公告
// @Summary      获取公告
// @Description  非系统管理员获取公告，从最新的公告开始读取一定条数，允许跳过一定条数，获取全部直接输一个过大值，注意：返回条数不一定等于请求条数
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.LoginRequest  true  "获取公告请求体"
// @Success      200      {object}  response.CommonResponse{data=model.UserResponse}
// @Router       /api/v1/announcement [post]
func (a *AnnouncementHandlerGroup) GetHandler(c *gin.Context) {
	req := model.AnnouncementGetRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
	}
	req.Auth = false
	ans := service.AnnouncementService.GetAnnouncements(&req)
	if ans == nil {
		response.FailWithCode(c, response.CodeAnnouncementNotFound)
		return
	}
	response.SuccessWithData(c, ans)
	return
}

// AnnouncementAuthGetHandler  获取公告
// @Summary      获取公告
// @Description  系统管理员获取公告，从最新的公告开始读取一定条数，允许跳过一定条数，获取全部直接输一个过大值，注意：返回条数不一定等于请求条数
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.LoginRequest  true  "获取公告请求体"
// @Success      200      {object}  response.CommonResponse{data=model.UserResponse}
// @Router       /api/v1/superadmin/announcement [post]
func (a *AnnouncementHandlerGroup) AuthGetHandler(c *gin.Context) {
	req := model.AnnouncementGetRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
	}
	ans := service.AnnouncementService.GetAnnouncements(&req)
	if ans == nil {
		response.FailWithCode(c, response.CodeAnnouncementNotFound)
		return
	}
	response.SuccessWithData(c, ans)
	return
}

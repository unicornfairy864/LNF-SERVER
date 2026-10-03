package basic

import (
	"strconv"

	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type AnnouncementHandlerGroup struct{}

// CreateHandler  创建公告
// @Summary      创建公告
// @Description  系统管理员创建公告（草稿）
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.AnnouncementUpdateRequest  true  "创建/更新公告请求体"
// @Success      200      {object}  response.CommonResponse{}
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
// @Description  系统管理员保存，发布，下架公告(更新阅读数还未完成，请先忽略)
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.AnnouncementUpdateRequest  true  "创建/更新公告请求体"
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

// AnnouncementDeleteHandler  删除公告
// @Summary      删除公告
// @Description  系统管理员删除公告
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        id      path      int64  true  "公告ID"
// @Success      200      {object}  response.CommonResponse{data=model.UserResponse}
// @Router       /api/v1/admin/announcement/{id} [delete]
func (a *AnnouncementHandlerGroup) DeleteHandler(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	errcode := service.AnnouncementService.DeleteAnnouncement(id)
	if errcode != response.CodeSuccess {
		response.FailWithCode(c, errcode)
	}
	response.Success(c)
}

// GetHandler  获取公告
// @Summary      获取公告
// @Description  非系统管理员获取公告，从给定id（默认最新）从最新的公告开始读取一定条数，允许跳过一定条数，获取全部直接输一个过大值
// @Description  分页查询第一次StartedId带0，之后请带上上一次返回的最小（最老）id-1，
// @Description  ignore依旧可以使用，但是请尽量不要使用不然可能出现重复返回，并且影响性能
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.AnnouncementGetRequest  true  "获取公告请求体"
// @Success      200      {object}  response.CommonResponse{data=[]model.Announcement}
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
}

// AuthGetHandler  获取公告
// @Summary      获取公告
// @Description  系统管理员获取公告，从给定id（默认最新）的公告开始读取一定条数，允许跳过一定条数，获取全部直接输一个过大值
// @Description  分页查询第一次StartedId带0，之后请带上上一次返回的最小（最老）id-1，
// @Description  ignore依旧可以使用，但是请尽量不要使用不然可能出现重复返回，并且影响性能
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.AnnouncementGetRequest  true  "获取公告请求体"
// @Success      200      {object}  response.CommonResponse{data=[]model.Announcement}
// @Router       /api/v1/admin/announcement [post]
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
}

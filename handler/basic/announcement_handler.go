package basic

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type AnnouncementHandlerGroup struct{}

// parseAnnouncementID 解析路径中的公告 ID（非正整数 → 1）
func parseAnnouncementID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.FailWithCode(c, response.CodeParamError)
		return 0, false
	}
	return id, true
}

// GetHandler 公开分页获取已发布公告列表
// @Summary      公开分页获取已发布公告列表
// @Description  仅返回已发布（status=1）且未删除的公告，按 id 倒序（新→旧）；page 缺省为 1，page_size 缺省为 10、上限 100，越界报参数错误 1；data = {total, page, page_size, announcements}
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        page       query   int   false  "页码，默认1"
// @Param        page_size  query   int   false  "每页数量，默认10，最大100"
// @Success      200  {object}  response.CommonResponse{data=model.AnnouncementListResponse}
// @Router       /api/v1/announcement [get]
func (a *AnnouncementHandlerGroup) GetHandler(c *gin.Context) {
	req := model.AnnouncementListQuery{}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := service.AnnouncementService.GetPublishedList(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// GetDetailHandler 公开获取公告详情
// @Summary      公开获取公告详情
// @Description  返回公告详情（仅已发布可见），每次访问浏览量 +1 且返回值包含本次 +1；公告不存在/已下架/已删除返回 90001，id 非正整数返回 1
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        id   path   int  true  "公告ID"
// @Success      200  {object}  response.CommonResponse{data=model.AnnouncementResponse}
// @Router       /api/v1/announcement/{id} [get]
func (a *AnnouncementHandlerGroup) GetDetailHandler(c *gin.Context) {
	id, ok := parseAnnouncementID(c)
	if !ok {
		return
	}
	res, code := service.AnnouncementService.GetDetail(id)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// CreateHandler 创建公告（创建即发布）
// @Summary      创建公告
// @Description  系统管理员创建公告，创建即发布：status=1、published_at=当前时间、admin_id 取自 JWT（不信任请求体）；标题空白或超100字节、内容空白、type 不在 0-3、is_top 不在 0/1 返回 90003；缺字段/JSON 类型错返回 1；成功 data 为 {}
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.CreateAnnouncementRequest  true  "创建公告请求体"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/admin/announcement/create [post]
func (a *AnnouncementHandlerGroup) CreateHandler(c *gin.Context) {
	req := model.CreateAnnouncementRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.AnnouncementService.Create(c.GetInt64(middleware.ContextID), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// UpdatedHandler 增量更新公告
// @Summary      更新公告
// @Description  系统管理员增量更新：id 必传，其余字段传了才更新（指针语义），view_count/published_at/created_at 不可更新；status 仅允许 1已发布/2已下架（传 0/3 返回 90003），published_at 保持首次发布时间不变；公告不存在返回 90001
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        request  body      model.AnnouncementUpdateRequest  true  "更新公告请求体"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/admin/announcement/update [post]
func (a *AnnouncementHandlerGroup) UpdatedHandler(c *gin.Context) {
	req := model.AnnouncementUpdateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.AnnouncementService.Update(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// DeleteHandler 删除公告
// @Summary      删除公告
// @Description  系统管理员软删除公告（is_deleted=1），删除后公开/管理列表与详情均不可见；公告不存在返回 90001，id 非正整数返回 1
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        id      path      int  true  "公告ID"
// @Success      200     {object}  response.CommonResponse{}
// @Router       /api/v1/admin/announcement/{id} [delete]
func (a *AnnouncementHandlerGroup) DeleteHandler(c *gin.Context) {
	id, ok := parseAnnouncementID(c)
	if !ok {
		return
	}
	code := service.AnnouncementService.Delete(id)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// AuthGetHandler 管理端分页获取公告列表
// @Summary      管理端分页获取公告列表
// @Description  管理员视角：含已发布与已下架（status=0 历史废弃行不可见），按 id 倒序；status 可选筛选 1已发布/2已下架；page/page_size 行为同公开列表
// @Tags         announcement
// @Accept       json
// @Produce      json
// @Param        page       query   int   false  "页码，默认1"
// @Param        page_size  query   int   false  "每页数量，默认10，最大100"
// @Param        status     query   int   false  "状态筛选: 1已发布 2已下架，默认全部"
// @Success      200  {object}  response.CommonResponse{data=model.AnnouncementListResponse}
// @Router       /api/v1/admin/announcement [get]

func (a *AnnouncementHandlerGroup) AuthGetHandler(c *gin.Context) {
	req := model.AnnouncementListQuery{}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := service.AnnouncementService.GetAdminList(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

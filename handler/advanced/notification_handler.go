package advanced

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type NotificationHandlerGroup struct{}

// List 通知列表
// @Tags Notification
// @Summary 获取通知列表
// @Description 获取当前用户的通知列表，支持类型/已读状态/管理员筛选与分页
// @Description 分页可能出现重复返回问题，出现问题优先注意这里，解决方法参照公告模块
// @Accept json
// @Produce json
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Param type query int false "类型: 0系统通知 1物品匹配 2认领申请 3认领结果 4评论回复 5积分变动 6商品兑换"
// @Param is_read query int false "是否已读: 0未读 1已读"
// @Param admin_id query int false "发布管理员ID"
// @Success 200 {object} response.CommonResponse{data=[]model.NotificationItem}
// @Failure 400 {object} response.CommonResponse
// @Router /api/v1/notifications [get]
func (h *NotificationHandlerGroup) List(c *gin.Context) {
	// 2026-10-01 修复：原误用 "ContextID"，中间件实际写入 jwt:id（middleware.ContextID），导致恒取 0
	userID := c.GetInt64(middleware.ContextID)
	var req model.NotificationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	list, code := service.NotificationService.List(userID, &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, list)
}

// UnreadCount 未读数
// @Tags Notification
// @Summary 获取未读通知数量
// @Description 获取当前用户的未读通知数量，走 Redis 缓存（5 分钟过期）
// @Accept json
// @Produce json
// @Success 200 {object} response.CommonResponse{data=int64}
// @Failure 400 {object} response.CommonResponse
// @Router /api/v1/notifications/unread-count [get]
func (h *NotificationHandlerGroup) UnreadCount(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextID)
	count, code := service.NotificationService.UnreadCount(userID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, count)
}

// GetAndRead 详情自动标记已读
// @Tags Notification
// @Summary 获取通知详情
// @Description 获取通知详情，若未读则自动标记为已读
// @Accept json
// @Produce json
// @Param id path int true "通知ID"
// @Success 200 {object} response.CommonResponse{data=model.Notification}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /api/v1/notifications/{id} [get]
func (h *NotificationHandlerGroup) GetAndRead(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextID)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	n, code := service.NotificationService.GetAndRead(id, userID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, n)
}

// BatchMarkRead 批量已读
// @Tags Notification
// @Summary 批量标记通知已读
// @Description 批量将指定通知标记为已读，不限制条数，不返回条数
// @Accept json
// @Produce json
// @Param body body model.NotificationIDsRequest true "通知ID列表"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /api/v1/notifications/read [put]
func (h *NotificationHandlerGroup) BatchMarkRead(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextID)
	var req model.NotificationIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.NotificationService.BatchMarkRead(req.IDs, userID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// BatchDelete 批量删除
// @Tags Notification
// @Summary 批量删除通知
// @Description 批量软删除指定通知，跳过自己发给自己的记录，只返回返回码
// @Accept json
// @Produce json
// @Param body body model.NotificationIDsRequest true "通知ID列表"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /api/v1/notifications [delete]
func (h *NotificationHandlerGroup) BatchDelete(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextID)
	var req model.NotificationIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.NotificationService.BatchDelete(req.IDs, userID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// Send 管理侧发送通知
// @Tags Notification
// @Summary 发送通知（管理侧）
// @Description 发送通知，支持发给全体用户或指定用户，异步执行，admin_id 从登录态提取
// @Accept json
// @Produce json
// @Param body body model.NotificationSendRequest true "发送通知请求"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /api/v1/admin/notifications [post]
func (h *NotificationHandlerGroup) Send(c *gin.Context) {
	adminID := c.GetInt64(middleware.ContextID)
	var req model.NotificationSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.NotificationService.Send(&req, adminID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

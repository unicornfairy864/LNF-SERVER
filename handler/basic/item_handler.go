package basic

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type ItemHandlerGroup struct{}

// parseItemID 解析路径中的物品 ID
func parseItemID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("itemID"), 10, 64)
	if err != nil || id <= 0 {
		response.FailWithCode(c, response.CodeParamError)
		return 0, false
	}
	return id, true
}

// ListItemHandler 公开分页查询物品列表
// @Summary      公开分页查询物品列表
// @Description  按类型/状态/地点/标签/关键词筛选物品，分页返回 ItemResponse 数组<br />未指定 status 时默认只返回已发布（status=0）的未删除物品
// @Tags         item
// @Accept       json
// @Produce      json
// @Param        type        query   int     false  "类型: 0丢失 1拾到"
// @Param        status      query   int     false  "状态: 0已发布 1已认领 2已关闭，默认0"
// @Param        location_id query   int     false  "地点ID"
// @Param        tag_id      query   int     false  "标签ID"
// @Param        keyword     query   string  false  "标题/描述关键词"
// @Param        page        query   int     false  "页码，默认1"
// @Param        page_size   query   int     false  "每页数量，默认10，最大100"
// @Success      200  {object}  response.CommonResponse{data=model.ItemListResponse}
// @Router       /api/v1/item/list [get]
func (itemHandler *ItemHandlerGroup) ListItemHandler(c *gin.Context) {
	req := model.ListItemQuery{}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := service.ItemService.ListPublicService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// GetItemHandler 公开获取物品详情
// @Summary      公开获取物品详情
// @Description  返回物品详情（含地点链、标签、图片），浏览量 +1；物品不存在或已删除返回 20001
// @Tags         item
// @Accept       json
// @Produce      json
// @Param        itemID  path  int  true  "物品ID"
// @Success      200  {object}  response.CommonResponse{data=model.ItemResponse}
// @Router       /api/v1/item/{itemID} [get]
func (itemHandler *ItemHandlerGroup) GetItemHandler(c *gin.Context) {
	itemID, ok := parseItemID(c)
	if !ok {
		return
	}
	res, code := service.ItemService.GetDetailService(itemID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// CreateItemHandler 登录用户创建物品信息
// @Summary      登录用户创建物品信息
// @Description  创建失物（type=0）或招领（type=1）信息，创建即发布（status=0）；<br />user_id 取自 JWT；可携带 tag_ids 关联标签；location_id 指向不存在地点返回 20010，标签不存在返回 40001
// @Tags         item
// @Accept       json
// @Produce      json
// @Param        request  body  model.CreateItemRequest  true  "创建物品请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/item/create [post]
func (itemHandler *ItemHandlerGroup) CreateItemHandler(c *gin.Context) {
	req := model.CreateItemRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.ItemService.CreateService(c.GetInt64(middleware.ContextID), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// UpdateItemHandler 登录用户增量更新自己的物品
// @Summary      登录用户增量更新自己的物品
// @Description  仅发布者本人可改；给定 id 和非空字段做增量更新，非整体替换；<br />携带 tag_ids（可为空数组）时整体替换标签关联；非本人返回 20005，物品不存在返回 20001
// @Tags         item
// @Accept       json
// @Produce      json
// @Param        request  body  model.UpdateItemRequest  true  "更新物品请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/item/update [post]
func (itemHandler *ItemHandlerGroup) UpdateItemHandler(c *gin.Context) {
	req := model.UpdateItemRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.ItemService.UpdateService(c.GetInt64(middleware.ContextID), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// DeleteItemHandler 登录用户删除自己的物品
// @Summary      登录用户删除自己的物品
// @Description  仅发布者本人可删；items 表为逻辑删除，删除后 is_deleted=1，不再出现在任何列表/详情中
// @Tags         item
// @Accept       json
// @Produce      json
// @Param        request  body  model.DeleteItemRequest  true  "删除物品请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/item/delete [post]
func (itemHandler *ItemHandlerGroup) DeleteItemHandler(c *gin.Context) {
	req := model.DeleteItemRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.ItemService.DeleteService(c.GetInt64(middleware.ContextID), req.ID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// ListMyItemHandler 登录用户查询自己的发布记录
// @Summary      登录用户查询自己的发布记录
// @Description  查询当前用户发布的物品，默认返回全部状态；筛选与分页参数同公开列表接口
// @Tags         item
// @Accept       json
// @Produce      json
// @Param        type        query   int     false  "类型: 0丢失 1拾到"
// @Param        status      query   int     false  "状态: 0已发布 1已认领 2已关闭，默认全部"
// @Param        keyword     query   string  false  "标题/描述关键词"
// @Param        page        query   int     false  "页码，默认1"
// @Param        page_size   query   int     false  "每页数量，默认10，最大50"
// @Success      200  {object}  response.CommonResponse{data=model.ItemListResponse}
// @Router       /api/v1/item/mine [get]
func (itemHandler *ItemHandlerGroup) ListMyItemHandler(c *gin.Context) {
	req := model.ListItemQuery{}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := service.ItemService.ListMyService(c.GetInt64(middleware.ContextID), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// SetItemImagesHandler 登录用户覆盖式设置自己物品的图片
// @Summary      登录用户覆盖式设置自己物品的图片
// @Description  仅发布者本人可操作；整体替换该物品的图片列表，最多3张，sort_order 取值1-3且不可重复；<br />传空数组 images=[] 清空图片；超过3张返回 20013，参数非法返回 20014
// @Tags         item
// @Accept       json
// @Produce      json
// @Param        itemID  path    int  true  "物品ID"
// @Param        request body    model.SetItemImagesRequest  true  "覆盖式设置图片请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/item/{itemID}/images [post]
func (itemHandler *ItemHandlerGroup) SetItemImagesHandler(c *gin.Context) {
	itemID, ok := parseItemID(c)
	if !ok {
		return
	}
	req := model.SetItemImagesRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.ItemService.SetImagesService(c.GetInt64(middleware.ContextID), itemID, &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

package advanced

import (
	"strconv"

	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type LocationHandler struct{}

// ListLocationHandler 公开查询地点列表
// @Summary      公开查询地点列表
// @Description  按 parent_id / level 筛选地点（均可选，缺省返回全部），按 sort_order 升序；<br />传 parent_id=0 返回根节点列表，用于级联选择器
// @Tags         location
// @Accept       json
// @Produce      json
// @Param        parent_id  query  int  false  "父地点ID，0表示根节点"
// @Param        level      query  int  false  "树深度"
// @Success      200  {object}  response.CommonResponse{data=[]model.Location}
// @Router       /api/v1/location/list [get]
func (l *LocationHandler) ListLocationHandler(c *gin.Context) {
	req := model.ListLocationRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := service.LocationService.ListService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// GetItemLocationsHandler 公开查询物品地点链
// @Summary      公开查询物品地点链
// @Description  返回物品所在地点从根到叶的完整链路（含自身）；物品无地点时返回空数组；物品不存在返回 20001
// @Tags         location
// @Accept       json
// @Produce      json
// @Param        itemID  path  int  true  "物品ID"
// @Success      200  {object}  response.CommonResponse{data=[]model.Location}
// @Router       /api/v1/item/{itemID}/locations [get]
func (l *LocationHandler) GetItemLocationsHandler(c *gin.Context) {
	itemID, err := strconv.ParseInt(c.Param("itemID"), 10, 64)
	if err != nil || itemID <= 0 {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := service.LocationService.GetItemLocationsService(itemID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// CreateLocationHandler  管理员创建地点
// @Summary      管理员创建地点
// @Description  serviceAdmin 及以上角色可创建；根据父元素ID创建地点，parent_id=0 表示根节点（level=1），否则 level=父级level+1；<br />详细地址 address 可选，sort_order 为递增排序条件（允许0）；父级不存在返回 80001，同父级下重名返回 80002
// @Tags         location
// @Accept       json
// @Produce      json
// @Param        request  body      model.CreateLocationRequest  true  "创建地点请求体"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/location/create [post]
func (l *LocationHandler) CreateLocationHandler(c *gin.Context) {
	req := model.CreateLocationRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.LocationService.CreateService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// UpdateLocationHandler 管理员更新地点
// @Summary      管理员更新地点
// @Description  给定 id 和非空字段做增量更新，非整体替换；修改 parent_id 时校验父级存在（80001）、非自身非后代（1），<br />并在同一事务中重算整棵子树 level；地点不存在返回 80001，同父级下重名返回 80002
// @Tags         location
// @Accept       json
// @Produce      json
// @Param        request  body      model.UpdateLocationRequest  true  "更新地点请求体"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/location/update [post]
func (l *LocationHandler) UpdateLocationHandler(c *gin.Context) {
	req := model.UpdateLocationRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.LocationService.UpdateService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// DeleteLocationHandler 管理员删除地点
// @Summary      管理员删除地点
// @Description  硬删除（locations 表无 is_deleted）；存在子地点返回 80005，被物品 location_id 引用返回 80004，地点不存在返回 80001
// @Tags         location
// @Accept       json
// @Produce      json
// @Param        locationID  path  int  true  "地点ID"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/location/{locationID} [delete]
func (l *LocationHandler) DeleteLocationHandler(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("locationID"), 10, 64)
	if err != nil || locationID <= 0 {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.LocationService.DeleteService(locationID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

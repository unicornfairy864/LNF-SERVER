package advanced

import (
	"strconv"

	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type TagHandler struct{}

// ListTagHandler 公开查询全部标签
// @Summary      公开查询全部标签
// @Description  按 sort_order 升序返回全部标签，用于前端筛选器
// @Tags         tag
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.CommonResponse{data=[]model.Tag}
// @Router       /api/v1/tag/list [get]
func (t *TagHandler) ListTagHandler(c *gin.Context) {
	res, code := service.TagService.ListService()
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// CreateTagHandler 管理员创建标签
// @Summary      管理员创建标签
// @Description  serviceAdmin 及以上角色可创建；标签名必填且唯一（40002），长度 1-50（40004）；<br />color 可选（十六进制颜色），sort_order 为排序序号（默认0）
// @Tags         tag
// @Accept       json
// @Produce      json
// @Param        request  body  model.CreateTagRequest  true  "创建标签请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/tag/create [post]
func (t *TagHandler) CreateTagHandler(c *gin.Context) {
	req := model.CreateTagRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.TagService.CreateService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// UpdateTagHandler 管理员更新标签
// @Summary      管理员更新标签
// @Description  给定 id 和非空字段做增量更新，非整体替换；标签不存在返回 40001，改名与其他标签重复返回 40002
// @Tags         tag
// @Accept       json
// @Produce      json
// @Param        request  body  model.UpdateTagRequest  true  "更新标签请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/tag/update [post]
func (t *TagHandler) UpdateTagHandler(c *gin.Context) {
	req := model.UpdateTagRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.TagService.UpdateService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// DeleteTagHandler 管理员删除标签
// @Summary      管理员删除标签
// @Description  硬删除（tags 表无 is_deleted）；标签不存在返回 40001，仍被物品关联时返回 40005
// @Tags         tag
// @Accept       json
// @Produce      json
// @Param        tagID  path  int  true  "标签ID"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/tag/{tagID} [delete]
func (t *TagHandler) DeleteTagHandler(c *gin.Context) {
	tagID, err := strconv.ParseInt(c.Param("tagID"), 10, 64)
	if err != nil || tagID <= 0 {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.TagService.DeleteService(tagID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

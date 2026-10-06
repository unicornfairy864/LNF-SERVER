package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type CommentHandlerGroup struct{}

// CreateHandler 登录用户创建评论或回复
// @Summary      创建评论/回复
// @Description  需登录（JWT）；请求体 item_id、user_id、content 均必填（user_id 传非 0 值即可，服务端会用登录态覆盖）；<br />item_id 以请求体为准，路径参数不参与业务校验；parent_id 不传创建根评论，传入必须为已存在评论 id（否则返回 110001），新评论并入父评论所在评论树；<br />参数缺失/JSON 类型错误返回 1，数据库错误返回 6；成功 data 为 {}
// @Tags         comment
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer JWT"
// @Param        itemID       path    int     true  "物品ID（路由占位，实际以请求体 item_id 为准）"
// @Param        request        body    model.CreateCommentRequest  true  "创建评论请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/item/{itemID}/comments/create [post]
func (c *CommentHandlerGroup) CreateHandler(g *gin.Context) {
	req := model.CreateCommentRequest{}
	if err := g.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	req.UserID = g.GetInt64(middleware.ContextID)
	err := service.CommentService.CreateComment(&req)
	if err != response.CodeSuccess {
		response.FailWithCode(g, err)
		return
	}
	response.Success(g)
}

// UpdateHandler 管理员批量更新评论状态（含子树）
// @Summary      管理员更新评论状态
// @Description  需 serviceAdmin 及以上角色（role>=1）登录；请求体 id、status 均必填，缺失或 JSON 类型错误返回 1；<br />按 id 定位评论，对该评论及其全部后代（同一评论树内）批量更新 status；status 取值：0待审核 1正常 2已隐藏；<br />评论不存在返回 110001，数据库错误返回 6；成功 data 为 {}
// @Tags         comment
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer JWT"
// @Param        itemID       path    int     true  "物品ID（路由占位，当前实现不参与业务校验）"
// @Param        request        body    model.UpdateCommentRequest  true  "更新评论请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/item/{itemID}/comments/update [patch]
func (c *CommentHandlerGroup) UpdateHandler(g *gin.Context) {
	req := model.UpdateCommentRequest{}
	if err := g.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	err := service.CommentService.UpdateComment(&req)
	if err != response.CodeSuccess {
		response.FailWithCode(g, err)
		return
	}
	response.Success(g)
}

// GetListHandler 公开获取物品评论列表
// @Summary      公开获取物品评论列表
// @Description  无需登录；仅返回 status=1（正常）的评论（根评论与回复混排），按 id 降序；<br />分页为游标式：以 started_id 为起点、按 limit 分批向下读取，直到凑满 limit 条或数据耗尽；total 为本次返回条数（非全量总数）；<br />物品不存在或已删除返回 20001，数据库错误返回 6；成功 data 为 {total, comment_dtos}
// @Tags         comment
// @Accept       json
// @Produce      json
// @Param        itemID    path   int  true   "物品ID"
// @Param        started_id  query  int  false  "分页游标：返回 id <= started_id 的评论；首次建议传大于全部评论 id 的正数，翻页传本批最后一条 id - 1（缺省 0 可能重复首批）"
// @Param        limit       query  int  false  "本次返回条数上限，缺省 0（此时至多返回 1 条），请传正整数"
// @Success      200  {object}  response.CommonResponse{data=model.CommentsListDTO}
// @Router       /api/v1/item/{itemID}/comments [get]
func (c *CommentHandlerGroup) GetListHandler(g *gin.Context) {
	qreq := model.CommentGetlistRequestQuery{}
	preq := model.CommentGetListRequestParam{}
	if err := g.ShouldBindQuery(&qreq); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	if err := g.ShouldBindUri(&preq); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	list, err := service.CommentService.GetList(&qreq, &preq)
	if err != response.CodeSuccess {
		response.FailWithCode(g, err)
		return
	}
	response.SuccessWithData(g, list)
}

// GetChildHandler 公开获取评论回复子树
// @Summary      公开获取评论回复子树
// @Description  无需登录；以 query id 定位目标评论（评论不存在返回 110001），返回其回复子树（不含目标评论本身），仅含 status=1（正常）的评论；<br />depth 为向下展开层数：缺省 0 不展开（返回空列表），1 返回一级回复，以此类推（不超过子树实际深度）；<br />max_count 为返回节点数上限（缺省 0 时按 100 处理），超出时自动收窄 depth，仍无法满足返回 1；数据库错误返回 6；total 为本次返回条数
// @Tags         comment
// @Accept       json
// @Produce      json
// @Param        itemID    path   int  true   "物品ID（路由占位，当前实现不参与校验）"
// @Param        id         query  int  true   "目标评论ID"
// @Param        depth      query  int  false  "向下展开层数，缺省0（0=返回空，1=一级回复）"
// @Param        max_count  query  int  false  "返回节点数上限，缺省0（按100处理）"
// @Success      200  {object}  response.CommonResponse{data=model.CommentsListDTO}
// @Router       /api/v1/item/{itemID}/comments/replies [get]
func (c *CommentHandlerGroup) GetChildHandler(g *gin.Context) {
	qreq := model.GetChildrenRequest{}
	if err := g.ShouldBindQuery(&qreq); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	list, err := service.CommentService.GetChildComments(&qreq)
	if err != response.CodeSuccess {
		response.FailWithCode(g, err)
		return
	}
	response.SuccessWithData(g, list)
}

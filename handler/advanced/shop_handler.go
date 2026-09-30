package advanced

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type ShopHandlerGroup struct{}

// parseGoodsID 解析路径中的商品 ID（参数名 goodsID 与 shop_router 注册一致）
func parseGoodsID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("goodsID"), 10, 64)
	if err != nil || id <= 0 {
		response.FailWithCode(c, response.CodeParamError)
		return 0, false
	}
	return id, true
}

// ListGoodsHandler 公开分页查询商品列表
// @Summary      公开分页查询商品列表
// @Description  仅返回未下架（is_deleted=0）商品，sort_order 升序、创建时间降序；<br />支持名称关键词模糊匹配与积分区间筛选：闭区间含边界，只传其一为单边筛选，<br />min_price/max_price 传负数或 min>max 返回参数错误
// @Tags         shop
// @Accept       json
// @Produce      json
// @Param        keyword    query  string  false  "商品名称关键词（LIKE 模糊匹配）"
// @Param        min_price  query  int     false  "积分下限（闭区间，含边界），0 或不传表示不限"
// @Param        max_price  query  int     false  "积分上限（闭区间，含边界），0 或不传表示不限"
// @Param        page       query  int     false  "页码，默认1"
// @Param        page_size  query  int     false  "每页数量，默认10，最大100"
// @Success      200  {object}  response.CommonResponse{data=model.GoodListResponse}
// @Router       /api/v1/shop/goods/list [get]
func (shopHandler *ShopHandlerGroup) ListGoodsHandler(c *gin.Context) {
	req := model.ListGoodQuery{}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := service.ShopService.ListService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// GetGoodsHandler 公开获取商品详情
// @Summary      公开获取商品详情
// @Description  返回未下架商品详情；商品不存在或已下架返回 11001
// @Tags         shop
// @Accept       json
// @Produce      json
// @Param        goodsID  path  int  true  "商品ID"
// @Success      200  {object}  response.CommonResponse{data=model.Good}
// @Router       /api/v1/shop/goods/{goodsID} [get]
func (shopHandler *ShopHandlerGroup) GetGoodsHandler(c *gin.Context) {
	goodsID, ok := parseGoodsID(c)
	if !ok {
		return
	}
	res, code := service.ShopService.GetDetailService(goodsID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// CreateGoodsHandler 管理员创建商品
// @Summary      管理员创建商品
// @Description  创建积分商城商品；name 必填且 ≤100 字符，重名（未下架集合内）返回 11003；<br />price 取值 1~1000000，非法返回 11004；stock/sort_order 缺省 0
// @Tags         shop
// @Accept       json
// @Produce      json
// @Param        request  body  model.CreateGoodRequest  true  "创建商品请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/shop/goods/create [post]
func (shopHandler *ShopHandlerGroup) CreateGoodsHandler(c *gin.Context) {
	req := model.CreateGoodRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.ShopService.CreateService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// UpdateGoodsHandler 管理员增量更新商品
// @Summary      管理员增量更新商品
// @Description  按 id 增量更新给定字段（指针语义，非整体替换）；软删商品返回 11001；<br />重名返回 11003，price 非法返回 11004
// @Tags         shop
// @Accept       json
// @Produce      json
// @Param        request  body  model.UpdateGoodRequest  true  "更新商品请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/shop/goods/update [post]
func (shopHandler *ShopHandlerGroup) UpdateGoodsHandler(c *gin.Context) {
	req := model.UpdateGoodRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.ShopService.UpdateService(&req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// DeleteGoodsHandler 管理员下架商品
// @Summary      管理员下架商品（软删）
// @Description  软删除（is_deleted=1）即下架，商品不再出现在公开列表/详情/兑换中；<br />商品不存在返回 11001；历史订单为快照设计，不受影响
// @Tags         shop
// @Accept       json
// @Produce      json
// @Param        request  body  model.DeleteGoodRequest  true  "下架商品请求体"
// @Success      200  {object}  response.CommonResponse{}
// @Router       /api/v1/shop/goods/delete [post]
func (shopHandler *ShopHandlerGroup) DeleteGoodsHandler(c *gin.Context) {
	req := model.DeleteGoodRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	code := service.ShopService.DeleteService(req.ID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.Success(c)
}

// RedeemGoodsHandler 登录用户积分兑换商品
// @Summary      积分兑换商品
// @Description  单事务完成「条件扣库存（stock>0 防超卖）+ 扣积分（行锁，不足回滚）+ 写订单快照」；<br />仅绑定 QQ（5~11 位）的用户可兑换，未绑定返回 11005；<br />商品不存在/已下架返回 11001，库存不足返回 11002，积分不足返回 50001，用户不存在/禁用返回 10006；<br />兑换成功后异步发送群消息（失败不影响兑换结果）；返回订单号与兑换后剩余积分
// @Tags         shop
// @Produce      json
// @Param        goodsID  path  int  true  "商品ID"
// @Success      200  {object}  response.CommonResponse{data=model.RedeemGoodsResponse}
// @Router       /api/v1/shop/goods/{goodsID}/redeem [post]
func (shopHandler *ShopHandlerGroup) RedeemGoodsHandler(c *gin.Context) {
	goodsID, ok := parseGoodsID(c)
	if !ok {
		return
	}
	res, code := service.ShopService.RedeemGoodsService(c.GetInt64(middleware.ContextID), goodsID)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

// ListMyOrdersHandler 登录用户查询自己的兑换记录
// @Summary      查询我的兑换记录
// @Description  分页返回当前用户的兑换订单（created_at 降序）；订单为快照设计，商品改名/下架不影响历史记录
// @Tags         shop
// @Accept       json
// @Produce      json
// @Param        page       query  int  false  "页码，默认1"
// @Param        page_size  query  int  false  "每页数量，默认10，最大100"
// @Success      200  {object}  response.CommonResponse{data=model.OrderListResponse}
// @Router       /api/v1/shop/orders [get]
func (shopHandler *ShopHandlerGroup) ListMyOrdersHandler(c *gin.Context) {
	req := model.OrderListQuery{}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	res, code := service.ShopService.ListMyOrdersService(c.GetInt64(middleware.ContextID), &req)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

package basic

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/agent/orchestrator"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

// SimilarItemHandler 详情页「相似帖子」推荐
// @Summary      获取相似帖子推荐
// @Description  根据当前物品的类型/标签/地点/时间，返回相似度最高的其他帖子（同类型优先，不足用相反类型补齐）<br />纯 SQL 召回（location 全链命中 + 标签命中 + ngram 全文相关度），**不调用 LLM、无额外成本**<br />公开接口，无需登录；物品不存在或已删除返回 20001
// @Tags         item
// @Accept       json
// @Produce      json
// @Param        itemID  path   int  true   "物品ID"
// @Param        limit   query  int  false  "返回条数，默认5，最大10"
// @Success      200  {object}  response.CommonResponse{data=[]model.ItemResponse}
// @Router       /api/v1/item/{itemID}/similar [get]
func (itemHandler *ItemHandlerGroup) SimilarItemHandler(c *gin.Context) {
	itemID, ok := parseItemID(c)
	if !ok {
		return
	}
	limit := 5
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	res, code := orchestrator.Service.SimilarItems(itemID, limit)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}

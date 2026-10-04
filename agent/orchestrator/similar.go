package orchestrator

import (
	"log"

	"github.com/unicornfairy864/LNF-SERVER/agent/internal/vocab"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

// SimilarItems 详情页「相似帖子」推荐（批次 4：纯 SQL，无 LLM，成本 0）
//
// 规则（沿用召回口径，见 agent.md/agent.md §7）：
//   - 候选：is_deleted=0 且 status ∈ {0 已发布, 1 已认领}
//   - 计分：location 全链命中(0/1，忽略楼号) + 标签命中数；标签≥2 走严格门槛，否则模糊模式
//   - 排序：ngram 全文相关度 → 基础分 → lost_found_time 倒序
//   - 类型：**同类型优先**（先同 type 召回），不足时用相反 type 补齐
func (s *ServiceGroup) SimilarItems(itemID int64, limit int) ([]model.ItemResponse, response.Code) {
	item := dao.ItemDao.GetItemByID(itemID)
	if item.ID == 0 {
		return nil, response.CodeItemNotFound
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 10 {
		limit = 10
	}

	// 源物品的标签与地点（地点做全链展开、忽略楼号）
	tagIDs := make([]int64, 0, 6)
	tagNames := make([]string, 0, 6)
	for _, tg := range dao.ItemTagDao.GetTagsByItemID(itemID) {
		tagIDs = append(tagIDs, tg.ID)
		tagNames = append(tagNames, tg.Name)
	}
	var locationIDs []int64
	if item.LocationID != nil {
		locationIDs = vocab.Shared().LocationScopeIDs([]int64{*item.LocationID})
	}
	fuzzy := len(tagIDs) < 2

	out := make([]model.ItemResponse, 0, limit)
	seen := map[int64]bool{itemID: true}
	// 第一轮：同类型；第二轮：相反类型补齐
	for _, typ := range []int8{item.Type, oppositeType(item.Type)} {
		if len(out) >= limit {
			break
		}
		rows, err := dao.AgentRecallDao.RecallItems(&dao.AgentRecallParams{
			Type:        typ,
			Statuses:    []int8{0, 1},
			LocationIDs: locationIDs,
			TagIDs:      tagIDs,
			MinScore:    2,
			Fuzzy:       fuzzy,
			Keywords:    tagNames,
			Limit:       (limit - len(out)) + 4,
		})
		if err != nil {
			log.Printf("[agent] 相似推荐召回失败 item_id=%d type=%d: %v", itemID, typ, err)
			break
		}
		for _, row := range rows {
			if len(out) >= limit {
				break
			}
			if seen[row.ID] {
				continue
			}
			seen[row.ID] = true
			out = append(out, *buildItemResponse(&row.Item))
		}
	}
	return out, response.CodeSuccess
}

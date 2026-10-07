package dao

import (
	"strings"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	modeladv "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
)

// AgentRecallGroup Agent 召回专用 DAO（纯 SQL 计分，独立文件避免改动既有 item_dao.go）
type AgentRecallGroup struct{}

// AgentRecallParams 召回参数（全部由 orchestrator 完成校验与归一化后传入）
type AgentRecallParams struct {
	Type        int8       // 主结果类型：与用户物品相反（用户失物→招领帖；用户拾物→失物帖）
	Statuses    []int8     // 参与召回的状态（0 已发布 / 1 已认领）
	LocationIDs []int64    // 检索用地点集合（全链展开、已忽略楼号）
	TagIDs      []int64    // 标签集合（仅用于计分/排序，不设门槛）
	Keywords    []string   // 全文检索关键词（ngram FULLTEXT，BOOLEAN MODE）
	TimeFrom    *time.Time // items.lost_found_time 下界（含）
	TimeTo      *time.Time // items.lost_found_time 上界（含）
	Limit       int
}

// AgentRecallRow 召回结果行（基础分 + 全文相关度）
type AgentRecallRow struct {
	model.Item
	BaseScore int     `gorm:"column:base_score"`
	FtScore   float64 `gorm:"column:ft_score"`
}

// AgentItemTagRow 物品-标签批量查询行
type AgentItemTagRow struct {
	ItemID    int64   `gorm:"column:item_id"`
	ID        int64   `gorm:"column:id"`
	Name      string  `gorm:"column:name"`
	Color     *string `gorm:"column:color"`
	SortOrder int     `gorm:"column:sort_order"`
}

// TagsByItemIDs 批量查询多个物品的标签（一次查询，避免 N+1；用于候选渲染）
func (g *AgentRecallGroup) TagsByItemIDs(itemIDs []int64) (map[int64][]modeladv.Tag, error) {
	out := make(map[int64][]modeladv.Tag, len(itemIDs))
	if len(itemIDs) == 0 {
		return out, nil
	}
	var rows []AgentItemTagRow
	err := global.LNF_DB.Table("item_tags").
		Select("item_tags.item_id AS item_id, tags.id AS id, tags.name AS name, tags.color AS color, tags.sort_order AS sort_order").
		Joins("JOIN tags ON tags.id = item_tags.tag_id").
		Where("item_tags.item_id IN ?", itemIDs).
		Order("tags.sort_order ASC, tags.id ASC").
		Scan(&rows).Error
	if err != nil {
		return out, err
	}
	for _, r := range rows {
		out[r.ItemID] = append(out[r.ItemID], modeladv.Tag{
			ID:        r.ID,
			Name:      r.Name,
			Color:     r.Color,
			SortOrder: r.SortOrder,
		})
	}
	return out, nil
}

// agentRecallDefaultLimit 未指定 limit 时的兜底
const agentRecallDefaultLimit = 50

// RecallItems 召回候选：标签/地点计分与 ngram 全文相关度仅用于排序，不设「契合度 ≥ N」门槛
// （2026-10-07 用户裁定：删除原严格模式 `score ≥ agent_match_min_score`；候选只需
// 「标签/地点计分 > 0」或「全文命中」至少其一，相关性交由 LLM 精排判定）
// 全部条件参数化，无字符串拼接进 SQL 值
func (g *AgentRecallGroup) RecallItems(p *AgentRecallParams) ([]AgentRecallRow, error) {
	scoreExpr, scoreArgs := agentScoreExpr(p.LocationIDs, p.TagIDs)
	ftExpr, ftArgs := agentFtExpr(p.Keywords)

	limit := p.Limit
	if limit <= 0 {
		limit = agentRecallDefaultLimit
	}

	var sql strings.Builder
	args := make([]interface{}, 0, 16)
	sql.WriteString("SELECT items.*, (")
	sql.WriteString(scoreExpr)
	sql.WriteString(") AS base_score, (")
	sql.WriteString(ftExpr)
	sql.WriteString(") AS ft_score FROM items WHERE items.is_deleted = 0")
	args = append(args, scoreArgs...)
	args = append(args, ftArgs...)

	sql.WriteString(" AND items.status IN ?")
	args = append(args, p.Statuses)
	sql.WriteString(" AND items.type = ?")
	args = append(args, p.Type)
	if p.TimeFrom != nil {
		sql.WriteString(" AND items.lost_found_time >= ?")
		args = append(args, *p.TimeFrom)
	}
	if p.TimeTo != nil {
		sql.WriteString(" AND items.lost_found_time <= ?")
		args = append(args, *p.TimeTo)
	}
	// 信号门槛：至少命中「标签/地点计分」或「全文检索」之一（原「模糊模式」口径；
	// 2026-10-07 用户裁定删除 score 硬性门槛后，所有查询统一走此口径）
	sql.WriteString(" AND ((")
	sql.WriteString(scoreExpr)
	sql.WriteString(") > 0 OR (")
	sql.WriteString(ftExpr)
	sql.WriteString(") > 0)")
	args = append(args, scoreArgs...)
	args = append(args, ftArgs...)

	// 排序：全文相关度 → 基础分 → 时间新→旧
	sql.WriteString(" ORDER BY ft_score DESC, base_score DESC, items.lost_found_time DESC, items.id DESC LIMIT ?")
	args = append(args, limit)

	var rows []AgentRecallRow
	err := global.LNF_DB.Raw(sql.String(), args...).Scan(&rows).Error
	return rows, err
}

// agentScoreExpr 计分表达式：location 命中(0/1) + 标签命中数
func agentScoreExpr(locationIDs, tagIDs []int64) (string, []interface{}) {
	parts := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if len(locationIDs) > 0 {
		parts = append(parts, "(CASE WHEN items.location_id IN ? THEN 1 ELSE 0 END)")
		args = append(args, locationIDs)
	} else {
		parts = append(parts, "0")
	}
	if len(tagIDs) > 0 {
		parts = append(parts, "(SELECT COUNT(*) FROM item_tags it WHERE it.item_id = items.id AND it.tag_id IN ?)")
		args = append(args, tagIDs)
	} else {
		parts = append(parts, "0")
	}
	return strings.Join(parts, " + "), args
}

// agentFtExpr 全文相关度表达式（无关键词时恒 0，不触发索引）
func agentFtExpr(keywords []string) (string, []interface{}) {
	q := agentBooleanQuery(keywords)
	if q == "" {
		return "0", nil
	}
	return "MATCH(items.title, items.description) AGAINST (? IN BOOLEAN MODE)", []interface{}{q}
}

// agentBooleanQuery 关键词 → BOOLEAN 查询串（词间 OR 语义：命中越多分越高）
// 去掉 BOOLEAN MODE 操作符，避免语法错误与注入
var agentBooleanReplacer = strings.NewReplacer(
	"+", " ", "-", " ", ">", " ", "<", " ",
	"(", " ", ")", " ", "~", " ", "*", " ",
	"\"", " ", "@", " ", "\\", " ",
)

func agentBooleanQuery(keywords []string) string {
	words := make([]string, 0, len(keywords))
	for _, k := range keywords {
		k = strings.TrimSpace(agentBooleanReplacer.Replace(k))
		if k == "" {
			continue
		}
		words = append(words, k)
		if len(words) >= 8 {
			break
		}
	}
	return strings.Join(words, " ")
}

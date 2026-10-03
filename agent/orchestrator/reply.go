package orchestrator

import (
	"fmt"
	"strings"

	"github.com/unicornfairy864/LNF-SERVER/agent/internal/schema"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
)

// ==================== 通用小工具 ====================

// truncateRunes 按字符数截断（不切坏 UTF-8）
func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// directionText 精排场景说明（注入 prompt）
func directionText(itemType int8) string {
	if itemType == 1 {
		return "用户捡到了物品，正在寻找失主"
	}
	return "用户丢失了物品，正在寻找线索"
}

// draftTagNames 抽取结果中的标签名（品类 + 颜色 + 特征）
func draftTagNames(it *schema.ExtractItem) []string {
	names := make([]string, 0, 6)
	if it.ItemTag != nil {
		names = append(names, *it.ItemTag)
	}
	if it.ColorTag != nil {
		names = append(names, *it.ColorTag)
	}
	names = append(names, it.FeatureTags...)
	return names
}

// locationText 地点展示（链路 + 细节）
func locationText(it *schema.ExtractItem, locSet schema.LocationSet) string {
	if it.LocationID == nil {
		return ""
	}
	name := locSet.Name(*it.LocationID)
	if it.LocationDetail != nil {
		if d := strings.TrimSpace(*it.LocationDetail); d != "" {
			name += " " + d
		}
	}
	return name
}

// timeRangeText 时间区间展示
func timeRangeText(it *schema.ExtractItem) string {
	from, to := it.Times()
	switch {
	case from != nil && to != nil:
		return from.Format("2006-01-02 15:04") + " ~ " + to.Format("2006-01-02 15:04")
	case from != nil:
		return from.Format("2006-01-02 15:04") + " 之后"
	case to != nil:
		return to.Format("2006-01-02 15:04") + " 之前"
	}
	return ""
}

// missingLabels 缺失字段 → 中文提示
func missingLabels(fields []string) []string {
	label := map[string]string{
		"item":            "物品名称",
		"location":        "地点",
		"location_detail": "具体位置",
		"time":            "时间",
		"contact":         "联系方式",
		"color":           "颜色",
		"features":        "特征",
	}
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if s, ok := label[f]; ok {
			out = append(out, s)
		}
	}
	return out
}

// questionList 追问列表（缺失项提示 + LLM 追问，最多 2 条）
func questionList(d *model.AgentDraft) []string {
	if d == nil {
		return []string{}
	}
	out := make([]string, 0, 2)
	if labels := missingLabels(d.MissingFields); len(labels) > 0 {
		out = append(out, "还缺："+strings.Join(labels, "、"))
	}
	if f := strings.TrimSpace(d.Followup); f != "" {
		out = append(out, f)
	}
	return out
}

// typeLabel 类型中文名
func typeLabel(itemType int8) string {
	if itemType == 1 {
		return "招领"
	}
	return "失物"
}

// ==================== 回复文案（事实由代码渲染） ====================

// renderConfirmAsk 唯一一次的「补充信息 + 确认」提问
func renderConfirmAsk(d *model.AgentDraft) string {
	if d == nil {
		return renderChitchat()
	}
	var b strings.Builder
	b.WriteString("我帮你初步写好了")
	b.WriteString(typeLabel(d.Type))
	b.WriteString("信息：「")
	b.WriteString(d.Title)
	b.WriteString("」")
	if d.Description != "" && d.Description != d.Title {
		b.WriteString("（")
		b.WriteString(truncateRunes(d.Description, 60))
		b.WriteString("）")
	}
	if d.LocationName != "" {
		b.WriteString("，地点：")
		b.WriteString(d.LocationName)
	}
	if labels := missingLabels(d.MissingFields); len(labels) > 0 {
		b.WriteString("。还缺")
		b.WriteString(strings.Join(labels, "、"))
	}
	b.WriteString("。回复补充信息我会合并后发布；回复「确认」也会发布；回复「取消」则放弃；其他内容我不会发布。")
	return b.String()
}

// renderCreated 建帖成功
func renderCreated(d *model.AgentDraft, itemID int64) string {
	title := "物品信息"
	if d != nil && d.Title != "" {
		title = d.Title
	}
	if itemID > 0 {
		return fmt.Sprintf("已发布%s：「%s」（编号 #%d），可在「我的发布」查看。", typeLabel(draftType(d)), title, itemID)
	}
	return fmt.Sprintf("已发布%s：「%s」，可在「我的发布」查看。", typeLabel(draftType(d)), title)
}

func draftType(d *model.AgentDraft) int8 {
	if d == nil {
		return 0
	}
	return d.Type
}

// renderMatched 匹配到候选：**条数由代码渲染**（模型自报数量不可靠，已在 schema 层丢弃含数字的摘要）
func renderMatched(summary string, n int) string {
	base := fmt.Sprintf("为你找到 %d 条可能相关的帖子，请核对下面的列表。", n)
	if s := strings.TrimSpace(summary); s != "" {
		return base + " " + s
	}
	return base
}

// renderNoMatch 无匹配
func renderNoMatch() string {
	return "暂时没有找到匹配的帖子。你可以把情况说给我，我帮你登记成失物帖，有线索时再来核对。"
}

// renderNoExplicitConfirm 确认轮收到无关内容（严格模式：不发布）
func renderNoExplicitConfirm() string {
	return "没有收到明确的补充信息或确认，本次未发布任何信息。需要发布时把情况再告诉我一次即可。"
}

// renderCancelled 用户放弃
func renderCancelled() string {
	return "已取消本次操作。需要时再叫我。"
}

// renderChitchat 闲聊/无关语境
func renderChitchat() string {
	return "我是失物招领助手，可以：1）按你的描述匹配已有的失物/招领帖子；2）一句话登记失物或招领信息。直接把情况告诉我就行～"
}

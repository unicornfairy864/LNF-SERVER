package orchestrator

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/agent"
	"github.com/unicornfairy864/LNF-SERVER/agent/internal/prompt"
	"github.com/unicornfairy864/LNF-SERVER/agent/internal/schema"
	"github.com/unicornfairy864/LNF-SERVER/agent/internal/vocab"
	"github.com/unicornfairy864/LNF-SERVER/config"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	modeladv "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

// CreateItemFn 建帖函数注入点：由 initialization 在启动时绑定为
// service.ItemService.CreateService。
// 说明：本包不直接 import service，否则 service → chensong → chensong/internal/service
// 会与本包形成循环依赖，批次 3 的 QQ 侧将无法复用本包。
var CreateItemFn func(userID int64, req *model.CreateItemRequest) response.Code

// ==================== 抽取（LLM #1：判类 + 信息抽取） ====================

// extractOutcome 抽取结果 + 词表快照（保证校验与后续渲染使用同一份词表）
type extractOutcome struct {
	res    *schema.ExtractResult
	tagSet schema.TagSet
	locSet schema.LocationSet
}

// doExtract 调用 LLM 完成「语境判定 + 意图识别 + 信息抽取」，并做白名单/枚举校验
func doExtract(text string, images []string) (*extractOutcome, response.Code) {
	v := vocab.Shared()
	tagSet, err := v.TagSet()
	if err != nil {
		log.Printf("[agent] 词表(标签)加载失败: %v", err)
		return nil, response.CodeAgentLLMFailed
	}
	locSet, err := v.LocationSet()
	if err != nil {
		log.Printf("[agent] 词表(地点)加载失败: %v", err)
		return nil, response.CodeAgentLLMFailed
	}
	itemNames, colorNames, featNames, err := v.GroupedTagNames()
	if err != nil {
		log.Printf("[agent] 标签词表分组失败: %v", err)
		return nil, response.CodeAgentLLMFailed
	}
	locLines, err := v.LocationEnumLines()
	if err != nil {
		log.Printf("[agent] 地点词表加载失败: %v", err)
		return nil, response.CodeAgentLLMFailed
	}

	sys := prompt.ExtractSystem(itemNames, colorNames, featNames, locLines, time.Now())
	user := prompt.ExtractUser(text, len(images))

	var raw string
	if len(images) > 0 {
		raw, err = agent.Client.RequestJSONVision(sys, user, images)
	} else {
		raw, err = agent.Client.RequestJSON(sys, user)
	}
	if err != nil {
		log.Printf("[agent] 抽取调用失败: %v", err)
		return nil, response.CodeAgentLLMFailed
	}
	res, err := schema.ValidateExtract(raw, tagSet, locSet)
	if err != nil {
		log.Printf("[agent] 抽取结果解析失败: %v raw=%.200s", err, raw)
		return nil, response.CodeAgentLLMFailed
	}
	return &extractOutcome{res: res, tagSet: tagSet, locSet: locSet}, response.CodeSuccess
}

// buildDraft 把抽取结果转成可发布草稿（type：失物=0，招领=1）
func buildDraft(intent string, it *schema.ExtractItem, ex *schema.ExtractResult, locSet schema.LocationSet) *model.AgentDraft {
	if it == nil {
		return nil
	}
	d := &model.AgentDraft{
		Type:           0,
		Title:          strings.TrimSpace(it.Title),
		Description:    strings.TrimSpace(it.Description),
		LocationID:     it.LocationID,
		LocationDetail: it.LocationDetail,
		Contact:        it.Contact,
		MissingFields:  ex.MissingFields,
		Followup:       ex.FollowupQuestion,
	}
	if intent == schema.IntentCreateFound {
		d.Type = 1
	}
	if d.Description == "" {
		d.Description = d.Title
	}
	if d.Title == "" {
		d.Title = truncateRunes(d.Description, 30)
	}
	// 标签 ID 与名称
	ids := make([]int64, 0, 6)
	names := make([]string, 0, 6)
	if it.ItemTagID != nil {
		ids = append(ids, *it.ItemTagID)
	}
	if it.ItemTag != nil {
		names = append(names, *it.ItemTag)
	}
	if it.ColorTagID != nil {
		ids = append(ids, *it.ColorTagID)
	}
	if it.ColorTag != nil {
		names = append(names, *it.ColorTag)
	}
	ids = append(ids, it.FeatureTagIDs...)
	names = append(names, it.FeatureTags...)
	d.TagIDs = ids
	d.TagNames = names
	if it.LocationID != nil {
		d.LocationName = locSet.Name(*it.LocationID)
	}
	if from, to := it.Times(); from != nil || to != nil {
		if from != nil {
			d.TimeFrom = from.Format(time.RFC3339)
			d.LostFoundTime = d.TimeFrom
		}
		if to != nil {
			d.TimeTo = to.Format(time.RFC3339)
			if d.LostFoundTime == "" {
				d.LostFoundTime = d.TimeTo
			}
		}
	}
	return d
}

// ==================== 召回 + 精排 ====================

// matchOutcome 匹配结果
type matchOutcome struct {
	Verdict  string
	Summary  string
	Question string
	Matches  []model.AgentMatchBrief
	Similar  []model.AgentMatchBrief
	Matched  bool
}

// collectTagIDs 汇总抽取出的标签 ID（品类 + 颜色 + 特征）
func collectTagIDs(it *schema.ExtractItem) []int64 {
	ids := make([]int64, 0, 6)
	if it.ItemTagID != nil {
		ids = append(ids, *it.ItemTagID)
	}
	if it.ColorTagID != nil {
		ids = append(ids, *it.ColorTagID)
	}
	ids = append(ids, it.FeatureTagIDs...)
	return ids
}

// oppositeType 主结果类型：用户失物(0) → 招领帖(1)；用户拾物(1) → 失物帖(0)
func oppositeType(t int8) int8 {
	if t == 1 {
		return 0
	}
	return 1
}

// matchTimeWindow 时间窗（用户定稿：向前容差 beforeDays，总跨度不超过 windowDays）
func matchTimeWindow(it *schema.ExtractItem, itemType int8, st config.AgentSettings) (from, to *time.Time) {
	now := time.Now()
	windowStart := now.AddDate(0, 0, -st.MatchTimeWindowDays)
	fromT, toT := it.Times()
	base := fromT
	if itemType == 1 && toT != nil {
		base = toT // 用户拾物：以拾到时间（上界）为基准
	}
	if base == nil {
		return &windowStart, &now
	}
	if itemType == 0 {
		// 候选=招领帖：拾到时间不早于丢失时间（放宽 beforeDays），上界 now
		s := base.AddDate(0, 0, -st.MatchTimeBeforeDays)
		if s.Before(windowStart) {
			s = windowStart
		}
		return &s, &now
	}
	// 候选=失物帖：丢失时间不晚于拾到时间（放宽 beforeDays），下界 windowStart
	e := base.AddDate(0, 0, st.MatchTimeBeforeDays)
	if e.After(now) {
		e = now
	}
	return &windowStart, &e
}

// runMatch 召回 + 精排（主结果=相反 type；相似区=同 type 最多 2 条）
func runMatch(it *schema.ExtractItem, itemType int8, text string, locSet schema.LocationSet, st config.AgentSettings, topN int) (*matchOutcome, response.Code) {
	tagIDs := collectTagIDs(it)
	// 用户定稿：tag 粒度不足/数量 < 门槛 → 模糊模式（不设分数门槛）
	fuzzy := len(tagIDs) < st.MatchMinScore
	var locationIDs []int64
	if it.LocationID != nil {
		locationIDs = vocab.Shared().LocationScopeIDs([]int64{*it.LocationID})
	}
	timeFrom, timeTo := matchTimeWindow(it, itemType, st)

	rows, err := dao.AgentRecallDao.RecallItems(&dao.AgentRecallParams{
		Type:        oppositeType(itemType),
		Statuses:    []int8{0, 1},
		LocationIDs: locationIDs,
		TagIDs:      tagIDs,
		MinScore:    st.MatchMinScore,
		Fuzzy:       fuzzy,
		Keywords:    it.Keywords,
		TimeFrom:    timeFrom,
		TimeTo:      timeTo,
		Limit:       st.TopK,
	})
	if err != nil {
		log.Printf("[agent] 召回失败: %v", err)
		return nil, response.CodeDatabaseError
	}
	out := &matchOutcome{Verdict: schema.VerdictNoMatch, Matches: []model.AgentMatchBrief{}, Similar: []model.AgentMatchBrief{}}
	if len(rows) == 0 {
		return out, response.CodeSuccess
	}

	rr, code := doRerank(rows, it, itemType, text, locSet, st)
	if code != response.CodeSuccess {
		return nil, code
	}
	out.Verdict = rr.Verdict
	out.Summary = rr.Summary
	if rr.NeedMoreInfo != nil {
		out.Question = rr.NeedMoreInfo.Question
	}
	if rr.Verdict != schema.VerdictNoMatch {
		n := topN
		if n <= 0 {
			n = 3
		}
		if n > 10 {
			n = 10
		}
		out.Matches = buildMatchBriefs(rr.TopN(n), rows)
		out.Matched = len(out.Matches) > 0
	}
	// 相似区（同 type，最多 2 条，不参与精排）
	out.Similar = recallSimilar(it, itemType, st, rows)
	return out, response.CodeSuccess
}

// doRerank 调用 LLM 精排（候选为空时不调用）
func doRerank(rows []dao.AgentRecallRow, it *schema.ExtractItem, itemType int8, text string, locSet schema.LocationSet, st config.AgentSettings) (*schema.RerankResult, response.Code) {
	ids := make([]int64, 0, len(rows))
	for i := range rows {
		ids = append(ids, rows[i].ID)
	}
	tagMap, err := dao.AgentRecallDao.TagsByItemIDs(ids)
	if err != nil {
		log.Printf("[agent] 候选标签批量查询失败: %v", err)
		tagMap = map[int64][]modeladv.Tag{}
	}
	cands := make([]prompt.RerankCandidate, 0, len(rows))
	for i := range rows {
		r := rows[i]
		tagNames := make([]string, 0, 4)
		for _, t := range tagMap[r.ID] {
			tagNames = append(tagNames, t.Name)
		}
		locName := ""
		if r.LocationID != nil {
			locName = locSet.Name(*r.LocationID)
		}
		cands = append(cands, prompt.RerankCandidate{
			ID:          r.ID,
			Type:        r.Type,
			Status:      r.Status,
			Title:       r.Title,
			Description: r.Description,
			Tags:        tagNames,
			Location:    locName,
			Time:        r.LostFoundTime,
			BaseScore:   r.BaseScore,
		})
	}
	q := prompt.RerankQuery{
		Direction: directionText(itemType),
		Text:      truncateRunes(text, 200),
		Tags:      draftTagNames(it),
		Location:  locationText(it, locSet),
		TimeRange: timeRangeText(it),
	}
	sys := prompt.RerankSystem(st.StrongThreshold, st.AmbiguousThreshold)
	user := prompt.BuildRerankUser(q, cands)
	raw, err := agent.Client.RequestJSON(sys, user)
	if err != nil {
		log.Printf("[agent] 精排调用失败: %v", err)
		return nil, response.CodeAgentLLMFailed
	}
	idSet := make(map[int64]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	rr, err := schema.ValidateRerank(raw, idSet, st.StrongThreshold, st.AmbiguousThreshold)
	if err != nil {
		log.Printf("[agent] 精排结果解析失败: %v raw=%.200s", err, raw)
		return nil, response.CodeAgentLLMFailed
	}
	return rr, response.CodeSuccess
}

// buildMatchBriefs 把精排结果映射为对外结构（Item 复用既有 ItemResponse）
func buildMatchBriefs(ranked []schema.RerankItem, rows []dao.AgentRecallRow) []model.AgentMatchBrief {
	byID := make(map[int64]*dao.AgentRecallRow, len(rows))
	for i := range rows {
		byID[rows[i].ID] = &rows[i]
	}
	out := make([]model.AgentMatchBrief, 0, len(ranked))
	for _, r := range ranked {
		row, ok := byID[r.ItemID]
		if !ok {
			continue
		}
		out = append(out, model.AgentMatchBrief{
			ItemID:  r.ItemID,
			Score:   r.Score,
			Reasons: r.Reasons,
			Risk:    r.Risk,
			Item:    *buildItemResponse(&row.Item),
		})
	}
	return out
}

// recallSimilar 相似区（同 type，最多 2 条，纯 SQL 计分，不调用 LLM）
func recallSimilar(it *schema.ExtractItem, itemType int8, st config.AgentSettings, mainRows []dao.AgentRecallRow) []model.AgentMatchBrief {
	exclude := make(map[int64]struct{}, len(mainRows))
	for i := range mainRows {
		exclude[mainRows[i].ID] = struct{}{}
	}
	tagIDs := collectTagIDs(it)
	var locationIDs []int64
	if it.LocationID != nil {
		locationIDs = vocab.Shared().LocationScopeIDs([]int64{*it.LocationID})
	}
	now := time.Now()
	windowStart := now.AddDate(0, 0, -st.MatchTimeWindowDays)
	rows, err := dao.AgentRecallDao.RecallItems(&dao.AgentRecallParams{
		Type:        itemType,
		Statuses:    []int8{0, 1},
		LocationIDs: locationIDs,
		TagIDs:      tagIDs,
		MinScore:    st.MatchMinScore,
		Fuzzy:       len(tagIDs) < st.MatchMinScore,
		Keywords:    it.Keywords,
		TimeFrom:    &windowStart,
		TimeTo:      &now,
		Limit:       8,
	})
	if err != nil {
		return []model.AgentMatchBrief{}
	}
	out := make([]model.AgentMatchBrief, 0, 2)
	for i := range rows {
		if len(out) >= 2 {
			break
		}
		if _, skip := exclude[rows[i].ID]; skip {
			continue
		}
		out = append(out, model.AgentMatchBrief{
			ItemID: rows[i].ID,
			Score:  float64(rows[i].BaseScore),
			Item:   *buildItemResponse(&rows[i].Item),
		})
	}
	return out
}

// buildItemResponse 组装 ItemResponse（与 service/basic.buildResponse 同构：
// 地点链 + 图片 + 标签；此处本地实现以避免改动既有文件）
func buildItemResponse(item *model.Item) *model.ItemResponse {
	resp := &model.ItemResponse{
		ID:             item.ID,
		UserID:         item.UserID,
		Title:          item.Title,
		Description:    item.Description,
		Type:           item.Type,
		Status:         item.Status,
		LocationDetail: item.LocationDetail,
		Images:         make([]model.ItemImage, 0),
		Tags:           make([]modeladv.Tag, 0),
		Locations:      make([]modeladv.Location, 0),
		LostFoundTime:  item.LostFoundTime,
		Contact:        item.Contact,
		CreditReward:   item.CreditReward,
		ViewCount:      item.ViewCount,
		ClaimUserID:    item.ClaimUserID,
		ClaimTime:      item.ClaimTime,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
	if item.LocationID != nil && *item.LocationID > 0 {
		resp.Locations = dao.LocationDao.GetLocationChain(*item.LocationID)
	}
	resp.Images = dao.ItemImageDao.GetImagesByItemID(item.ID)
	resp.Tags = dao.ItemTagDao.GetTagsByItemID(item.ID)
	return resp
}

// ==================== 建帖（复用 ItemService.CreateService 的全部校验） ====================

// createItemFromDraft 按草稿建帖；返回新物品 ID（0 = 已创建但未能定位 ID，此时不返回该字段）
// fallbackContact 为联系方式兜底（QQ 侧=发送者 QQ 号）；草稿中用户明确给出的 contact 优先
func createItemFromDraft(userID int64, d *model.AgentDraft, st config.AgentSettings, fallbackContact string) (int64, response.Code) {
	if d == nil {
		return 0, response.CodeAgentStageConflict
	}
	title := strings.TrimSpace(d.Title)
	desc := strings.TrimSpace(d.Description)
	if title == "" && desc == "" {
		return 0, response.CodeAgentInputInvalid
	}
	if title == "" {
		title = truncateRunes(desc, 30)
	}
	if desc == "" {
		desc = title
	}
	itemType := d.Type
	if itemType != 0 && itemType != 1 {
		itemType = 0
	}
	lostFoundTime := time.Now()
	if d.LostFoundTime != "" {
		if t, err := time.Parse(time.RFC3339, d.LostFoundTime); err == nil {
			lostFoundTime = t
		}
	}
	// 地点：缺省用「其他地点」（agent_default_location_id，DB 中为 140）
	locationID := d.LocationID
	if locationID == nil || *locationID <= 0 {
		def := st.DefaultLocationID
		locationID = &def
	}
	// 详细信息：只写链路表达不了的内容；没有则「暂无」（用户定稿）
	detail := ""
	if d.LocationDetail != nil {
		detail = strings.TrimSpace(*d.LocationDetail)
	}
	if detail == "" {
		detail = "暂无"
	}
	req := &model.CreateItemRequest{
		Title:          title,
		Description:    desc,
		Type:           &itemType,
		LocationID:     locationID,
		LocationDetail: &detail,
		LostFoundTime:  lostFoundTime,
		CreditReward:   0,
		TagIDs:         d.TagIDs,
	}
	// 联系方式：优先用户在对话中明确给出的（草稿 contact）；否则用调用方兜底（QQ 侧=发送者 QQ）
	// 开关 openai.agent_fill_contact 控制是否写入；缺省开启，不编造
	contact := d.Contact
	if contact == nil || strings.TrimSpace(*contact) == "" {
		if fc := strings.TrimSpace(fallbackContact); fc != "" {
			contact = &fc
		}
	}
	if st.FillContact && contact != nil {
		if c := strings.TrimSpace(*contact); c != "" {
			req.Contact = &c
		}
	}
	if CreateItemFn == nil {
		log.Printf("[agent] 建帖函数未注入，拒绝创建物品（检查 initialization 的绑定）")
		return 0, response.CodeServerError
	}
	if code := CreateItemFn(userID, req); code != response.CodeSuccess {
		return 0, code
	}
	return locateCreatedItem(userID, title), response.CodeSuccess
}

// locateCreatedItem 定位刚创建的物品 ID（CreateService 不返回 ID；
// 取该用户最近 5 条中标题一致且 3 分钟内创建的记录，避免影响既有接口签名）
func locateCreatedItem(userID int64, title string) int64 {
	items, _, err := dao.ItemDao.GetItemPage(&model.ListItemQuery{Page: 1, PageSize: 5}, userID, nil)
	if err != nil {
		return 0
	}
	for i := range items {
		if items[i].Title == title && time.Since(items[i].CreatedAt) < 3*time.Minute {
			return items[i].ID
		}
	}
	if len(items) > 0 && time.Since(items[0].CreatedAt) < 3*time.Minute {
		return items[0].ID
	}
	return 0
}

// ==================== 补充信息合并（确认轮，仅 1 次） ====================

// doMerge 判断用户本轮回复是「确认 / 取消 / 补充信息 / 无关」，并抽取 patch
func doMerge(draft *model.AgentDraft, reply string) (*schema.MergeResult, response.Code) {
	v := vocab.Shared()
	tagSet, err := v.TagSet()
	if err != nil {
		return nil, response.CodeAgentLLMFailed
	}
	locSet, err := v.LocationSet()
	if err != nil {
		return nil, response.CodeAgentLLMFailed
	}
	itemNames, colorNames, featNames, err := v.GroupedTagNames()
	if err != nil {
		return nil, response.CodeAgentLLMFailed
	}
	locLines, err := v.LocationEnumLines()
	if err != nil {
		return nil, response.CodeAgentLLMFailed
	}
	b, err := json.Marshal(draft)
	if err != nil {
		return nil, response.CodeAgentLLMFailed
	}
	sys := prompt.MergeSystem(itemNames, colorNames, featNames, locLines, time.Now())
	user := prompt.BuildMergeUser(string(b), reply)
	raw, err := agent.Client.RequestJSON(sys, user)
	if err != nil {
		log.Printf("[agent] 合并调用失败: %v", err)
		return nil, response.CodeAgentLLMFailed
	}
	mr, err := schema.ValidateMerge(raw, tagSet, locSet)
	if err != nil {
		log.Printf("[agent] 合并结果解析失败: %v raw=%.200s", err, raw)
		return nil, response.CodeAgentLLMFailed
	}
	return mr, response.CodeSuccess
}

// applyMergePatch 把 patch 增量合并进草稿
func applyMergePatch(d *model.AgentDraft, p *schema.MergePatch) {
	if d == nil || p == nil {
		return
	}
	if p.Title != nil {
		d.Title = *p.Title
	}
	if p.Description != nil {
		d.Description = *p.Description
	}
	if p.ItemTagID != nil || p.ColorTagID != nil || p.FeatureTagIDs != nil {
		ids := make([]int64, 0, 6)
		names := make([]string, 0, 6)
		if p.ItemTagID != nil && p.ItemTag != nil {
			ids = append(ids, *p.ItemTagID)
			names = append(names, *p.ItemTag)
		}
		if p.ColorTagID != nil && p.ColorTag != nil {
			ids = append(ids, *p.ColorTagID)
			names = append(names, *p.ColorTag)
		}
		if p.FeatureTagIDs != nil {
			ids = append(ids, *p.FeatureTagIDs...)
			if p.FeatureTags != nil {
				names = append(names, *p.FeatureTags...)
			}
		}
		d.TagIDs = ids
		d.TagNames = names
	}
	if p.LocationID != nil {
		d.LocationID = p.LocationID
		d.LocationName = ""
	}
	if p.LocationDetail != nil {
		d.LocationDetail = p.LocationDetail
	}
	if p.Contact != nil {
		if strings.TrimSpace(*p.Contact) == "" {
			d.Contact = nil
		} else {
			d.Contact = p.Contact
		}
	}
	if p.TimeFrom != nil {
		d.TimeFrom = *p.TimeFrom
		d.LostFoundTime = *p.TimeFrom
	}
	if p.TimeTo != nil {
		d.TimeTo = *p.TimeTo
	}
}

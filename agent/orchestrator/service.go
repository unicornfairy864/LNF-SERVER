package orchestrator

import (
	"log"
	"strings"

	"github.com/unicornfairy864/LNF-SERVER/agent/internal/schema"
	"github.com/unicornfairy864/LNF-SERVER/agent/internal/vocab"
	"github.com/unicornfairy864/LNF-SERVER/config"
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

// ServiceGroup Agent 编排服务：API 与 QQBOT 共用同一套业务链路
// （QQ 侧批次 3 通过同一入口调用，仅前置过滤与文案渲染不同）
type ServiceGroup struct{}

// Service 全局编排入口
var Service ServiceGroup

// settings 读取归一化后的 Agent 配置
func settings() config.AgentSettings { return global.LNF_CONFIG.OpenAI.AgentValues() }

// ==================== 对外接口 ====================

// Extract 无状态抽取（只读）：文本/图片 → 意图 + 草稿，供前端「智能填充」使用
func (s *ServiceGroup) Extract(userID int64, req *model.AgentExtractRequest) (*model.AgentExtractResponse, response.Code) {
	st := settings()
	if !st.Enabled {
		return nil, response.CodeAgentNotAvailable
	}
	if !allowRate(userID, st.RateLimitPerMinute) {
		return nil, response.CodeAgentRateLimited
	}
	if !allowRateTotal(st.RateLimitTotalMin) {
		return nil, response.CodeAgentRateLimited
	}
	text, images, code := prepareInput(req.Text, req.ImageURLs, st)
	if code != response.CodeSuccess {
		return nil, code
	}
	out, code := doExtract(text, images)
	if code != response.CodeSuccess {
		return nil, code
	}
	resp := &model.AgentExtractResponse{
		Intent:        out.res.Intent,
		IsLNFContext:  out.res.IsLNFContext,
		MissingFields: nonNilStrings(out.res.MissingFields),
	}
	if it := out.res.FirstItem(); it != nil {
		resp.Draft = buildDraft(out.res.Intent, it, out.res, out.locSet)
	}
	resp.Questions = questionList(resp.Draft)
	return resp, response.CodeSuccess
}

// Match 无状态匹配（只读）：文本/图片 → 意图 + 候选帖（主结果 + 相似区）
func (s *ServiceGroup) Match(userID int64, req *model.AgentMatchRequest) (*model.AgentMatchResponse, response.Code) {
	st := settings()
	if !st.Enabled {
		return nil, response.CodeAgentNotAvailable
	}
	if !allowRate(userID, st.RateLimitPerMinute) {
		return nil, response.CodeAgentRateLimited
	}
	if !allowRateTotal(st.RateLimitTotalMin) {
		return nil, response.CodeAgentRateLimited
	}
	text, images, code := prepareInput(req.Text, req.ImageURLs, st)
	if code != response.CodeSuccess {
		return nil, code
	}
	out, code := doExtract(text, images)
	if code != response.CodeSuccess {
		return nil, code
	}
	resp := &model.AgentMatchResponse{
		Intent:  out.res.Intent,
		Matches: []model.AgentMatchBrief{},
		Similar: []model.AgentMatchBrief{},
		Verdict: schema.VerdictNoMatch,
	}
	it := out.res.FirstItem()
	if it == nil || (!schema.IsCreateIntent(out.res.Intent) && out.res.Intent != schema.IntentMatch) {
		return resp, response.CodeSuccess
	}
	intent := out.res.Intent
	draftType := int8(0)
	if intent == schema.IntentCreateFound {
		draftType = 1
	}
	draft := buildDraft(intent, it, out.res, out.locSet)
	resp.Entities = draft
	if draft != nil {
		draftType = draft.Type
	}
	mo, code := runMatch(it, draftType, text, out.locSet, st, req.TopN)
	if code != response.CodeSuccess {
		return nil, code
	}
	resp.Matches = mo.Matches
	resp.Similar = mo.Similar
	resp.Summary = mo.Summary
	resp.Verdict = mo.Verdict
	return resp, response.CodeSuccess
}

// Chat 会话主链路：匹配 / 建帖草稿 / 唯一一次确认轮 / 建帖
func (s *ServiceGroup) Chat(userID int64, req *model.AgentChatRequest) (*model.AgentChatResponse, response.Code) {
	st := settings()
	if !st.Enabled {
		return nil, response.CodeAgentNotAvailable
	}
	if !allowRate(userID, st.RateLimitPerMinute) {
		return nil, response.CodeAgentRateLimited
	}
	if !allowRateTotal(st.RateLimitTotalMin) {
		return nil, response.CodeAgentRateLimited
	}
	text, images, code := prepareInput(req.Text, req.ImageURLs, st)
	if code != response.CodeSuccess {
		return nil, code
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action == "" {
		action = model.AgentActionAuto
	}

	// 会话加载（严格模式，用户 2026-10-03 定稿）：
	//   · 不带 session_id → 一律新建会话并立即覆盖旧会话（旧 session_id 失效）
	//   · 带 session_id 但不匹配/不存在 → 返回 120001，且**不改动已存会话**
	sess, has := loadSession(userID)
	log.Printf("[agent] chat 入口 user_id=%d action=%s req_session=%q has_session=%v stage=%s rounds=%d",
		userID, action, req.SessionID, has, sessionStage(sess), sessionRounds(sess))
	if req.SessionID != "" {
		if !has || sess.SessionID != req.SessionID {
			log.Printf("[agent] 会话不匹配（不改动已存会话）user_id=%d req_session=%q stored=%q", userID, req.SessionID, sessionIDOf(sess))
			return nil, response.CodeAgentSessionNotFound
		}
		if sess.Rounds >= agentMaxRounds {
			log.Printf("[agent] 会话轮次用尽，重建 user_id=%d session=%s rounds=%d", userID, sess.SessionID, sess.Rounds)
			sess = &sessionState{SessionID: newSessionID()}
			saveSession(userID, sess, st.SessionTTL)
		}
	} else {
		sess = &sessionState{SessionID: newSessionID()}
		saveSession(userID, sess, st.SessionTTL)
	}

	// 显式动作优先
	switch action {
	case model.AgentActionCancel:
		dropSession(userID)
		return cancelledResponse(sess.SessionID), response.CodeSuccess
	case model.AgentActionConfirm:
		if sess.Stage == model.AgentStageCreated && sess.CreatedItemID > 0 {
			// 幂等：重复确认直接返回已创建结果
			return createdResponse(sess), response.CodeSuccess
		}
		if sess.Stage != model.AgentStageNeedConfirm || sess.Draft == nil {
			return nil, response.CodeAgentStageConflict
		}
		return s.finishCreate(userID, sess, st)
	}

	sess.Rounds++
	if sess.Stage == model.AgentStageNeedConfirm && sess.Draft != nil {
		return s.handleConfirmRound(userID, sess, text, st)
	}
	return s.handleNewRequest(userID, sess, text, images, st)
}

// CloseSession 关闭会话（幂等）
func (s *ServiceGroup) CloseSession(userID int64, req *model.AgentSessionCloseRequest) response.Code {
	dropSession(userID)
	return response.CodeSuccess
}

// ==================== 内部流程 ====================

// handleNewRequest 处理一轮新请求（判类 + 抽取 → 分支）
func (s *ServiceGroup) handleNewRequest(userID int64, sess *sessionState, text string, images []string, st config.AgentSettings) (*model.AgentChatResponse, response.Code) {
	out, code := doExtract(text, images)
	if code != response.CodeSuccess {
		return nil, code
	}
	sess.Intent = out.res.Intent
	resp := &model.AgentChatResponse{
		SessionID: sess.SessionID,
		Intent:    out.res.Intent,
		Questions: []string{},
		Matches:   []model.AgentMatchBrief{},
		Similar:   []model.AgentMatchBrief{},
	}
	it := out.res.FirstItem()

	switch {
	case schema.IsCreateIntent(out.res.Intent) && it != nil:
		draft := buildDraft(out.res.Intent, it, out.res, out.locSet)
		sess.Stage = model.AgentStageNeedConfirm
		sess.Draft = draft
		resp.Stage = model.AgentStageNeedConfirm
		resp.Draft = draft
		resp.Questions = questionList(draft)
		resp.Reply = renderConfirmAsk(draft)
	case out.res.Intent == schema.IntentMatch && it != nil:
		// 匹配意图默认按「用户是失主」方向检索（主结果=招领帖，相似区=失物帖）
		draft := buildDraft(schema.IntentCreateLost, it, out.res, out.locSet)
		sess.Draft = draft
		itemType := int8(0)
		if draft != nil {
			itemType = draft.Type
		}
		mo, mcode := runMatch(it, itemType, text, out.locSet, st, 3)
		if mcode != response.CodeSuccess {
			return nil, mcode
		}
		if mo.Matched {
			sess.Stage = model.AgentStageMatched
			resp.Stage = model.AgentStageMatched
			resp.Matches = mo.Matches
			resp.Similar = mo.Similar
			resp.Reply = renderMatched(mo.Summary, len(mo.Matches))
			resp.Questions = questionList(nil)
		} else {
			sess.Stage = model.AgentStageNoMatch
			resp.Stage = model.AgentStageNoMatch
			resp.Similar = mo.Similar
			resp.Reply = renderNoMatch()
			if mo.Question != "" {
				resp.Questions = []string{mo.Question}
			}
		}
	default:
		sess.Stage = model.AgentStageChitchat
		resp.Stage = model.AgentStageChitchat
		resp.Reply = renderChitchat()
	}
	saveSession(userID, sess, st.SessionTTL)
	return resp, response.CodeSuccess
}

// handleConfirmRound 处理确认轮回复（用户 2026-10-03 定稿：**只有明确确认或补充信息才发布**）：
//  1. 快速词表命中「取消/拒绝」→ 直接取消（不建帖）
//  2. 命中「确认」→ 用当前草稿发布
//  3. 其余：交给 LLM 判定语义
//     · provide_info → 合并补充信息后发布
//     · confirm      → 发布
//     · cancel       → 取消
//     · unrelated（以及 LLM 故障）→ **不发布**（故障时保留会话；无关内容则结束会话）
func (s *ServiceGroup) handleConfirmRound(userID int64, sess *sessionState, text string, st config.AgentSettings) (*model.AgentChatResponse, response.Code) {
	if isCancelText(text) {
		dropSession(userID)
		return cancelledResponse(sess.SessionID), response.CodeSuccess
	}
	if isConfirmText(text) {
		return s.finishCreate(userID, sess, st)
	}
	mr, code := doMerge(sess.Draft, text)
	if code != response.CodeSuccess {
		// LLM 不可用：严格模式下不发布，也不结束会话（保留草稿，用户可稍后重试）
		log.Printf("[agent] 确认轮合并失败，严格模式不发布 user_id=%d", userID)
		return nil, code
	}
	switch mr.Decision {
	case schema.DecisionCancel:
		dropSession(userID)
		return cancelledResponse(sess.SessionID), response.CodeSuccess
	case schema.DecisionProvideInfo:
		applyMergePatch(sess.Draft, mr.Patch)
		refreshDraftLocationName(sess.Draft)
		return s.finishCreate(userID, sess, st)
	case schema.DecisionConfirm:
		return s.finishCreate(userID, sess, st)
	default:
		// unrelated：严格模式不发布，结束会话
		dropSession(userID)
		resp := cancelledResponse(sess.SessionID)
		resp.Reply = renderNoExplicitConfirm()
		return resp, response.CodeSuccess
	}
}

// finishCreate 建帖（复用 ItemService.CreateService 的全部校验）
func (s *ServiceGroup) finishCreate(userID int64, sess *sessionState, st config.AgentSettings) (*model.AgentChatResponse, response.Code) {
	draft := sess.Draft
	itemID, code := createItemFromDraft(userID, draft, st)
	if code != response.CodeSuccess {
		// 创建失败：保留会话与草稿，用户可补充后重试
		saveSession(userID, sess, st.SessionTTL)
		return nil, code
	}
	sess.Stage = model.AgentStageCreated
	sess.CreatedItemID = itemID
	sess.Draft = nil
	saveSession(userID, sess, st.SessionTTL)
	resp := &model.AgentChatResponse{
		SessionID: sess.SessionID,
		Stage:     model.AgentStageCreated,
		Reply:     renderCreated(draft, itemID),
		Questions: []string{},
		Matches:   []model.AgentMatchBrief{},
		Similar:   []model.AgentMatchBrief{},
	}
	if itemID > 0 {
		id := itemID
		resp.CreatedItemID = &id
	}
	return resp, response.CodeSuccess
}

// ==================== 输入预处理 ====================

// prepareInput 清洗与校验输入：长度上限、图片 URL 归一化（相对路径拼公网前缀）与数量上限
func prepareInput(text string, imageURLs []string, st config.AgentSettings) (string, []string, response.Code) {
	t := strings.TrimSpace(text)
	if t == "" {
		return "", nil, response.CodeAgentInputInvalid
	}
	if len([]rune(t)) > st.MaxInputChars {
		return "", nil, response.CodeAgentInputInvalid
	}
	images := make([]string, 0, 3)
	for _, raw := range imageURLs {
		u := strings.TrimSpace(raw)
		if u == "" || len(u) > 500 {
			continue
		}
		if strings.HasPrefix(u, "/") { // 相对路径（/uploads/...）：拼公网前缀
			if st.PublicBaseURL == "" {
				continue
			}
			u = st.PublicBaseURL + u
		} else if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			continue
		}
		images = append(images, u)
		if len(images) >= 3 {
			break
		}
	}
	return t, images, response.CodeSuccess
}

// ==================== 回复动作判定（快速词表，命中即不调用 LLM） ====================

var agentConfirmWords = map[string]struct{}{
	"确认": {}, "确认发布": {}, "确认创建": {}, "确定": {}, "可以": {}, "好的": {}, "好": {},
	"是的": {}, "是": {}, "没问题": {}, "就这样": {}, "发布": {}, "发布吧": {}, "ok": {},
}

var agentCancelWords = map[string]struct{}{
	"取消": {}, "取消吧": {}, "算了": {}, "不用了": {}, "不用": {}, "不要了": {},
	"不创建": {}, "先不创建": {}, "先不": {}, "不发布": {}, "不发了": {}, "别发": {},
	"别发布了": {}, "不登记": {}, "撤回": {}, "放弃": {}, "no": {},
}

// isConfirmText 是否为纯确认（去掉首尾标点后精确匹配，避免「对，地址是…」被误判为确认）
func isConfirmText(t string) bool {
	_, ok := agentConfirmWords[normalizeReply(t)]
	return ok
}

// isCancelText 是否为纯取消
func isCancelText(t string) bool {
	_, ok := agentCancelWords[normalizeReply(t)]
	return ok
}

// normalizeReply 去首尾空白与常见标点，转小写
func normalizeReply(t string) string {
	t = strings.TrimSpace(t)
	t = strings.Trim(t, "。．.！!？?，,、；;:：~～ ")
	t = strings.TrimSpace(t)
	return strings.ToLower(t)
}

// ==================== 小工具 ====================

// nonNilStrings 保证 JSON 输出为 []（而不是 null）
func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

// cancelledResponse 取消/结束响应
func cancelledResponse(sessionID string) *model.AgentChatResponse {
	return &model.AgentChatResponse{
		SessionID: sessionID,
		Stage:     model.AgentStageCancelled,
		Reply:     renderCancelled(),
		Questions: []string{},
		Matches:   []model.AgentMatchBrief{},
		Similar:   []model.AgentMatchBrief{},
	}
}

// createdResponse 幂等重复确认时的响应（复用已创建的物品 ID）
func createdResponse(sess *sessionState) *model.AgentChatResponse {
	id := sess.CreatedItemID
	resp := &model.AgentChatResponse{
		SessionID: sess.SessionID,
		Stage:     model.AgentStageCreated,
		Reply:     renderCreated(nil, id),
		Questions: []string{},
		Matches:   []model.AgentMatchBrief{},
		Similar:   []model.AgentMatchBrief{},
	}
	if id > 0 {
		resp.CreatedItemID = &id
	}
	return resp
}

// sessionStage 日志辅助：nil 会话安全取阶段
func sessionStage(s *sessionState) string {
	if s == nil {
		return ""
	}
	return s.Stage
}

// sessionRounds 日志辅助
func sessionRounds(s *sessionState) int {
	if s == nil {
		return 0
	}
	return s.Rounds
}

// sessionIDOf 日志辅助
func sessionIDOf(s *sessionState) string {
	if s == nil {
		return ""
	}
	return s.SessionID
}

// refreshDraftLocationName 合并补充信息后刷新地点展示名（patch 只带 id，不带名称）
func refreshDraftLocationName(d *model.AgentDraft) {
	if d == nil || d.LocationID == nil {
		return
	}
	if locSet, err := vocab.Shared().LocationSet(); err == nil {
		d.LocationName = locSet.Name(*d.LocationID)
	}
}

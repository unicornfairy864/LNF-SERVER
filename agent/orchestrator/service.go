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

// 意图常量透出（供 QQ 侧判断是否回执）：confirm_round 表示“确认轮”消息（一定是真实业务，必回执）
const (
	IntentChitchat     = schema.IntentChitchat
	IntentOther        = schema.IntentOther
	IntentConfirmRound = "confirm_round"
)

// Chat API 会话主入口（会话域=user_id，含 API 限流）：匹配 / 建帖草稿 / 唯一一次确认轮 / 建帖
func (s *ServiceGroup) Chat(userID int64, req *model.AgentChatRequest) (*model.AgentChatResponse, response.Code) {
	return s.chat(newAPIScope(userID), userID, req, true, "", false, nil)
}

// ChatQQ QQ 会话主入口（会话域=QQ 号；QQ 侧自带每 QQ 冷却与每群频控，此处不重复限流）
// fallbackContact：联系方式兜底（QQ 侧传发送者 QQ；对话中明确给出联系方式时会覆盖它）
// reuseSession：QQ 消息不带 session_id，会话键即 QQ 号 → **续用已有会话**（否则补充信息会被当成新对话）
// onIntent：判类完成后的回调（可能被调用一次；confirm_round 表示确认轮）——QQ 侧用它决定何时发“我正在思考”回执
func (s *ServiceGroup) ChatQQ(qq string, userID int64, fallbackContact string, req *model.AgentChatRequest, onIntent func(intent string)) (*model.AgentChatResponse, response.Code) {
	return s.chat(newQQScope(qq), userID, req, false, fallbackContact, true, onIntent)
}

// chat 共用实现（API 与 QQBOT 完全同链路，仅会话域/限流策略/联系方式兜底/回调不同）
func (s *ServiceGroup) chat(sc sessionScope, userID int64, req *model.AgentChatRequest, withRateLimit bool, fallbackContact string, reuseSession bool, onIntent func(intent string)) (*model.AgentChatResponse, response.Code) {
	st := settings()
	if !st.Enabled {
		return nil, response.CodeAgentNotAvailable
	}
	if withRateLimit {
		if !allowRate(userID, st.RateLimitPerMinute) {
			return nil, response.CodeAgentRateLimited
		}
		if !allowRateTotal(st.RateLimitTotalMin) {
			return nil, response.CodeAgentRateLimited
		}
	}
	text, images, code := prepareInput(req.Text, req.ImageURLs, st)
	if code != response.CodeSuccess {
		return nil, code
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action == "" {
		action = model.AgentActionAuto
	}

	// 会话加载：
	//   · 带 session_id（API）：必须匹配，否则 120001 且**不改动已存会话**；轮次用尽时重建
	//   · QQ 侧（reuseSession=true，消息不带 session_id）：会话键即 QQ 号 → **续用已有会话**，仅在轮次用尽时重建
	//   · API 不带 session_id（严格模式）：新建并覆盖旧会话（旧 session_id 失效）
	sess, has := loadSession(sc)
	log.Printf("[agent] chat 入口 scope=%s user_id=%d action=%s req_session=%q has_session=%v stage=%s rounds=%d",
		sc.String(), userID, action, req.SessionID, has, sessionStage(sess), sessionRounds(sess))
	switch {
	case req.SessionID != "":
		if !has || sess.SessionID != req.SessionID {
			log.Printf("[agent] 会话不匹配（不改动已存会话）scope=%s req_session=%q stored=%q", sc.String(), req.SessionID, sessionIDOf(sess))
			return nil, response.CodeAgentSessionNotFound
		}
		if sess.Rounds >= agentMaxRounds {
			log.Printf("[agent] 会话轮次用尽，重建 scope=%s session=%s rounds=%d", sc.String(), sess.SessionID, sess.Rounds)
			sess = &sessionState{SessionID: newSessionID()}
			saveSession(sc, sess, st.SessionTTL)
		}
	case reuseSession && has:
		if sess.Rounds >= agentMaxRounds {
			log.Printf("[agent] QQ 会话轮次用尽，重建 scope=%s session=%s rounds=%d", sc.String(), sess.SessionID, sess.Rounds)
			sess = &sessionState{SessionID: newSessionID()}
			saveSession(sc, sess, st.SessionTTL)
		}
	default:
		sess = &sessionState{SessionID: newSessionID()}
		saveSession(sc, sess, st.SessionTTL)
	}

	// 显式动作优先
	switch action {
	case model.AgentActionCancel:
		dropSession(sc)
		return cancelledResponse(sess.SessionID), response.CodeSuccess
	case model.AgentActionConfirm:
		if sess.Stage == model.AgentStageCreated && sess.CreatedItemID > 0 {
			// 幂等：重复确认直接返回已创建结果
			return createdResponse(sess), response.CodeSuccess
		}
		if sess.Stage != model.AgentStageNeedConfirm || sess.Draft == nil {
			return nil, response.CodeAgentStageConflict
		}
		return s.finishCreate(sc, userID, sess, st, fallbackContact)
	}

	sess.Rounds++
	if sess.Stage == model.AgentStageNeedConfirm && sess.Draft != nil {
		return s.handleConfirmRound(sc, userID, sess, text, st, fallbackContact, onIntent)
	}
	return s.handleNewRequest(sc, userID, sess, text, images, st, onIntent)
}

// CloseSession 关闭会话（幂等；仅 API 域）
func (s *ServiceGroup) CloseSession(userID int64, req *model.AgentSessionCloseRequest) response.Code {
	dropSession(newAPIScope(userID))
	return response.CodeSuccess
}

// ==================== 内部流程 ====================

// handleNewRequest 处理一轮新请求（判类 + 抽取 → 分支）
func (s *ServiceGroup) handleNewRequest(sc sessionScope, userID int64, sess *sessionState, text string, images []string, st config.AgentSettings, onIntent func(intent string)) (*model.AgentChatResponse, response.Code) {
	out, code := doExtract(text, images)
	if code != response.CodeSuccess {
		return nil, code
	}
	sess.Intent = out.res.Intent
	// 判类完成回调（QQ 侧据此决定是否发“我正在思考”回执：闲聊/无关不回执）
	if onIntent != nil {
		onIntent(out.res.Intent)
	}
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
			// 站内通知：仅 QQ 侧写（用户不在网页端，站内留痕便于回看；API 侧与会话响应重复）
			if sc.kind == "qq" {
				notifyItemMatched(userID, mo.Matches)
			}
		} else {
			sess.Stage = model.AgentStageNoMatch
			resp.Stage = model.AgentStageNoMatch
			resp.Similar = mo.Similar
			// 方案 B（2026-10-07 用户裁定）：主召回为空时的同类型兜底结果放 similar，有则文案一并提示
			if len(mo.Similar) > 0 {
				resp.Reply = renderNoMatchWithRelated(len(mo.Similar))
			} else {
				resp.Reply = renderNoMatch()
			}
			if mo.Question != "" {
				resp.Questions = []string{mo.Question}
			}
		}
	default:
		sess.Stage = model.AgentStageChitchat
		resp.Stage = model.AgentStageChitchat
		resp.Reply = renderChitchat()
	}
	saveSession(sc, sess, st.SessionTTL)
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
func (s *ServiceGroup) handleConfirmRound(sc sessionScope, userID int64, sess *sessionState, text string, st config.AgentSettings, fallbackContact string, onIntent func(intent string)) (*model.AgentChatResponse, response.Code) {
	// 确认轮一定是真实业务（用户正在补充/确认），直接回执
	if onIntent != nil {
		onIntent(IntentConfirmRound)
	}
	if isCancelText(text) {
		dropSession(sc)
		return cancelledResponse(sess.SessionID), response.CodeSuccess
	}
	if isConfirmText(text) {
		return s.finishCreate(sc, userID, sess, st, fallbackContact)
	}
	mr, code := doMerge(sess.Draft, text)
	if code != response.CodeSuccess {
		// LLM 不可用：严格模式下不发布，也不结束会话（保留草稿，用户可稍后重试）
		log.Printf("[agent] 确认轮合并失败，严格模式不发布 user_id=%d", userID)
		return nil, code
	}
	switch mr.Decision {
	case schema.DecisionCancel:
		dropSession(sc)
		return cancelledResponse(sess.SessionID), response.CodeSuccess
	case schema.DecisionProvideInfo:
		applyMergePatch(sess.Draft, mr.Patch)
		refreshDraftLocationName(sess.Draft)
		return s.finishCreate(sc, userID, sess, st, fallbackContact)
	case schema.DecisionConfirm:
		return s.finishCreate(sc, userID, sess, st, fallbackContact)
	default:
		// unrelated：严格模式不发布，结束会话
		dropSession(sc)
		resp := cancelledResponse(sess.SessionID)
		resp.Reply = renderNoExplicitConfirm()
		return resp, response.CodeSuccess
	}
}

// finishCreate 建帖（复用 ItemService.CreateService 的全部校验）
func (s *ServiceGroup) finishCreate(sc sessionScope, userID int64, sess *sessionState, st config.AgentSettings, fallbackContact string) (*model.AgentChatResponse, response.Code) {
	draft := sess.Draft
	itemID, code := createItemFromDraft(userID, draft, st, fallbackContact)
	if code != response.CodeSuccess {
		// 创建失败：保留会话与草稿，用户可补充后重试
		saveSession(sc, sess, st.SessionTTL)
		return nil, code
	}
	sess.Stage = model.AgentStageCreated
	sess.CreatedItemID = itemID
	sess.Draft = nil
	saveSession(sc, sess, st.SessionTTL)
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
	// 站内通知：建帖成功（「通知站内全部」，QQ 与 API 均写）
	notifyItemCreated(userID, draft, itemID)
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

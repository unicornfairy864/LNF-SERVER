package orchestrator

import (
	"log"
	"strconv"

	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
)

// ==================== 站内通知（notification 模块打通） ====================
// 用户 2026-10-04 定稿：agent 的关键事件同步写入站内通知（「通知站内全部」）。

// 通知类型（与 notification 模块既有取值一致）
const (
	NotifyTypeSystem    int8 = 0 // 系统通知
	NotifyTypeItemMatch int8 = 1 // 物品匹配（此前无使用方，现由 agent 使用）
)

// NotifyFn 站内通知注入点：由 initialization 在启动时绑定为
// service.NotificationService.Create(0, userID, ntype, title, content, relatedID)。
// 说明：本包不直接 import service/*（会与 chensong 形成循环依赖）。
var NotifyFn func(userID int64, ntype int8, title, content string, relatedID *int64) error

// notifyItemCreated 建帖成功后给发帖人写一条站内通知
func notifyItemCreated(userID int64, draft *model.AgentDraft, itemID int64) {
	if draft == nil || itemID <= 0 {
		return
	}
	rid := itemID
	notify(userID, NotifyTypeSystem, "智能助手已为你发布信息",
		"你的"+typeLabel(draft.Type)+"信息「"+truncateRunes(draft.Title, 40)+"」已发布，可在“我的发布”中查看或修改。", &rid)
}

// notifyItemMatched 匹配到候选后写一条站内通知
// 仅 QQ 侧调用：用户当时不在网页端，站内留痕便于回看（API 侧与会话响应重复，不写）
func notifyItemMatched(userID int64, matches []model.AgentMatchBrief) {
	if len(matches) == 0 {
		return
	}
	rid := matches[0].ItemID
	notify(userID, NotifyTypeItemMatch, "为你匹配到可能相关的帖子",
		"共 "+strconv.Itoa(len(matches))+" 条，最相关的是「"+truncateRunes(matches[0].Item.Title, 40)+"」。", &rid)
}

// notify 统一写入：未注入或失败仅记日志，不影响主流程
func notify(userID int64, ntype int8, title, content string, relatedID *int64) {
	if NotifyFn == nil {
		log.Printf("[agent] 站内通知未注入，跳过 user_id=%d type=%d", userID, ntype)
		return
	}
	if err := NotifyFn(userID, ntype, title, content, relatedID); err != nil {
		log.Printf("[agent] 站内通知写入失败 user_id=%d type=%d: %v", userID, ntype, err)
	}
}

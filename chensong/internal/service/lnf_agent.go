package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/agent/orchestrator"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/client"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/model"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/utils"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	"github.com/unicornfairy864/LNF-SERVER/global"
	modelbasic "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

// ==================== QQ 侧 Agent 链路（批次 3） ====================
// 规则（agent.md/agent.md §6）：
//
//	· 仅监听 chensong.activated_group（私聊与其它群一律跳过，连预过滤都不做）
//	· 触发条件：**@ 机器人**（2026-10-04 定稿：取消关键词过滤，补充信息/确认类消息不再被误拦）
//	· 媒体（图片/语音/视频/文件，含长 URI）一律忽略；纯文本 >500 字直接跳过
//	· 每 QQ 冷却 lnf_cooldown + 每群每分钟 lnf_group_rate_per_minute 条
//	· 未绑定 QQ → 提示先在网页版绑定；闲聊/无关 → 静默
//	· 业务链路复用 agent/orchestrator（与 API 完全同源，含两步确认）
type LnfAgentService struct{}

// LnfAgent 全局实例
var LnfAgent = &LnfAgentService{}

const (
	lnfMaxTextRunes   = 500
	lnfMsgIdemTTL     = 5 * time.Minute
	lnfUnboundQQReply = "若使用陈松的 agent 功能，需要先在网页版绑定 QQ 后再使用～"
	lnfErrReply       = "智能服务开小差了，稍后再试～"
)

// Trigger 触发判定（同步轻量，不查库）：仅 activated_group + @机器人 + 文本非空且不超 500 字
// 2026-10-04 定稿：**取消关键词过滤**（补充信息、确认类消息往往不含关键词，会被误拦）；
// 成本由「每 QQ 冷却 + 每群每分钟条数」兜住，闲聊由 LLM 判类后静默。
func (s *LnfAgentService) Trigger(req model.GroupMessageEvent, text string) bool {
	cfg := global.LNF_CONFIG.ChenSong
	if req.GroupID != cfg.ActivatedGroup {
		return false // 仅监听 activated_group（用户定稿）
	}
	if strings.TrimSpace(text) == "" {
		return false
	}
	if len([]rune(text)) > lnfMaxTextRunes {
		return false // 超长请求直接拦截
	}
	return utils.HasAtBot(req, cfg.ActivatedQQ)
}

// MarkOnce 幂等：同一条 message_id 只处理一次（INCR 原子；Redis 异常时放行，避免丢消息）
func (s *LnfAgentService) MarkOnce(messageID int64) bool {
	key := "lnf:agent:qq:msg:" + strconv.FormatInt(messageID, 10)
	n, err := dao.RedisDao.INCR(key)
	if err != nil {
		return true
	}
	if n == 1 {
		_ = dao.RedisDao.SetKey(key, n, lnfMsgIdemTTL)
	}
	return n == 1
}

// Handle 异步处理（handler 用 go 调用）：限流 → 身份 → 复用编排层 → 群内回复
func (s *LnfAgentService) Handle(req model.GroupMessageEvent, text string) {
	cfg := global.LNF_CONFIG.ChenSong
	lnf := cfg.LnfValues()
	qq := strconv.FormatInt(req.UserID, 10)

	// 每 QQ 冷却（窗口内 1 次）
	if !allowInWindow("lnf:agent:rate:qq:"+qq, 1, lnf.Cooldown) {
		log.Printf("[chensong] agent 触发被冷却拦截 qq=%s", qq)
		return
	}
	// 每群每分钟条数
	groupKey := "lnf:agent:rate:group:" + strconv.FormatInt(req.GroupID, 10) + ":" + strconv.FormatInt(time.Now().Unix()/60, 10)
	if !allowInWindow(groupKey, lnf.GroupRatePerMinute, time.Minute) {
		log.Printf("[chensong] agent 触发被群频控拦截 group=%d", req.GroupID)
		return
	}

	// 身份解析：QQ → users.qq（未绑定/被禁用 → 提示并终止）
	user := dao.UserDao.GetUserByQQ(qq)
	if user.ID == 0 || user.Status == 0 {
		s.replyGroup(req, lnfUnboundQQReply)
		return
	}

	// 复用编排层（会话域=QQ 号；联系方式兜底=发送者 QQ，用户明确给出时会覆盖）
	resp, code := orchestrator.Service.ChatQQ(qq, user.ID, qq, &modelbasic.AgentChatRequest{Text: text})
	if code != response.CodeSuccess {
		if code == response.CodeAgentNotAvailable || code == response.CodeAgentRateLimited {
			return
		}
		log.Printf("[chensong] agent 处理失败 qq=%s code=%d", qq, code)
		s.replyGroup(req, lnfErrReply)
		return
	}
	if resp == nil || resp.Stage == modelbasic.AgentStageChitchat {
		return // 闲聊/无关：静默，避免打搅群
	}
	reply := strings.TrimSpace(resp.Reply)
	if reply == "" {
		return
	}
	s.replyGroup(req, reply)
}

// replyGroup 群内回复：引用原消息 + @ 用户 + 昵称（只发 activated_group）
func (s *LnfAgentService) replyGroup(req model.GroupMessageEvent, text string) {
	nickname := strings.TrimSpace(req.Sender.Card)
	if nickname == "" {
		nickname = strings.TrimSpace(req.Sender.Nickname)
	}
	msg := fmt.Sprintf("[CQ:reply,id=%d] [CQ:at,qq=%d] %s %s", req.MessageID, req.UserID, nickname, text)
	res, err := client.Client.SendGroupMessage(msg, req.GroupID)
	if err != nil || res == nil || res.Status != "ok" {
		log.Printf("[chensong] agent 回复发送失败 group=%d qq=%d: %v", req.GroupID, req.UserID, err)
	}
}

// allowInWindow INCR + TTL 计数：窗口内计数 <= limit 放行（Redis 异常放行；limit<=0 视为不限制）
func allowInWindow(key string, limit int, ttl time.Duration) bool {
	if limit <= 0 {
		return true
	}
	n, err := dao.RedisDao.INCR(key)
	if err != nil {
		return true
	}
	if n == 1 {
		_ = dao.RedisDao.SetKey(key, n, ttl)
	}
	return n <= int64(limit)
}

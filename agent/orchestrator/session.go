package orchestrator

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
)

// 会话/限流的 Redis 键前缀（API 侧；批次 3 的 QQ 侧用 qq: 前缀，见设计文档 §9）
const (
	agentSessionKeyPrefix = "lnf:agent:session:api:"
	agentRateKeyPrefix    = "lnf:agent:rate:api:"
	agentRateTotalPrefix  = "lnf:agent:rate:total:"
)

// agentMaxRounds 单会话最多处理的用户消息数（成本上界，超出自动结束旧会话）
const agentMaxRounds = 3

// sessionState 会话状态（JSON 存 Redis，TTL 滑动续期）
type sessionState struct {
	SessionID     string            `json:"session_id"`
	Stage         string            `json:"stage"`
	Intent        string            `json:"intent"`
	Draft         *model.AgentDraft `json:"draft,omitempty"`
	Rounds        int               `json:"rounds"`
	CreatedItemID int64             `json:"created_item_id,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

func sessionKey(userID int64) string {
	return agentSessionKeyPrefix + strconv.FormatInt(userID, 10)
}

// rateKey 按「用户 + 分钟」计数
func rateKey(userID int64) string {
	return agentRateKeyPrefix + strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(time.Now().Unix()/60, 10)
}

// rateTotalKey 按「全系统 + 分钟」计数
func rateTotalKey() string {
	return agentRateTotalPrefix + strconv.FormatInt(time.Now().Unix()/60, 10)
}

// newSessionID 会话 ID（对外返回，前端回传；与 user_id 绑定校验，防越权）
func newSessionID() string {
	return "as_" + uuid.NewString()
}

// loadSession 读取会话（Redis 异常时按「无会话」处理，不阻断主流程）
func loadSession(userID int64) (*sessionState, bool) {
	raw, err := dao.RedisDao.GetValueString(sessionKey(userID))
	if err != nil {
		log.Printf("[agent] 读会话失败 user_id=%d: %v", userID, err)
		return nil, false
	}
	if raw == "" {
		return nil, false
	}
	var st sessionState
	if err := json.Unmarshal([]byte(raw), &st); err != nil || st.SessionID == "" {
		log.Printf("[agent] 会话反序列化失败 user_id=%d raw=%.120s err=%v", userID, raw, err)
		return nil, false
	}
	return &st, true
}

// saveSession 写入会话（滑动续期）
func saveSession(userID int64, st *sessionState, ttl time.Duration) {
	if st == nil {
		return
	}
	st.UpdatedAt = time.Now()
	b, err := json.Marshal(st)
	if err != nil {
		log.Printf("[agent] 会话序列化失败 user_id=%d: %v", userID, err)
		return
	}
	if err := dao.RedisDao.SetKey(sessionKey(userID), string(b), ttl); err != nil {
		log.Printf("[agent] 写会话失败 user_id=%d ttl=%s: %v", userID, ttl, err)
	}
}

// dropSession 结束会话
func dropSession(userID int64) {
	_ = dao.RedisDao.DelKey(sessionKey(userID))
}

// allowRate 每用户每分钟限流；Redis 异常时放行（不因缓存故障阻断用户）
func allowRate(userID int64, limit int) bool {
	return allowByKey(rateKey(userID), limit)
}

// allowRateTotal 全系统每分钟限流（所有用户合计）；Redis 异常时放行
func allowRateTotal(limit int) bool {
	return allowByKey(rateTotalKey(), limit)
}

// allowByKey INCR + TTL 计数：计数 <= limit 放行
func allowByKey(key string, limit int) bool {
	if limit <= 0 {
		return true
	}
	n, err := dao.RedisDao.INCR(key)
	if err != nil {
		return true
	}
	if n == 1 {
		_ = dao.RedisDao.SetKey(key, n, time.Minute)
	}
	return n <= int64(limit)
}

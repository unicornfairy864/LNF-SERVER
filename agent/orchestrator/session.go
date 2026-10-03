package orchestrator

import (
	"encoding/json"
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
)

// agentMaxRounds 单会话最多处理的用户消息数（成本上界，超出自动结束旧会话）
const agentMaxRounds = 3

// agentMaxConfirmRounds need_confirm 阶段允许的「无效回复」次数
// 用户定稿：有且仅有一次「补充信息 + 确认」；再收到无关回复即结束会话
const agentMaxConfirmRounds = 1

// sessionState 会话状态（JSON 存 Redis，TTL 滑动续期）
type sessionState struct {
	SessionID     string            `json:"session_id"`
	Stage         string            `json:"stage"`
	Intent        string            `json:"intent"`
	Draft         *model.AgentDraft `json:"draft,omitempty"`
	Rounds        int               `json:"rounds"`
	ConfirmRounds int               `json:"confirm_rounds"`
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

// newSessionID 会话 ID（对外返回，前端回传；与 user_id 绑定校验，防越权）
func newSessionID() string {
	return "as_" + uuid.NewString()
}

// loadSession 读取会话（Redis 异常时按「无会话」处理，不阻断主流程）
func loadSession(userID int64) (*sessionState, bool) {
	raw, err := dao.RedisDao.GetValueString(sessionKey(userID))
	if err != nil || raw == "" {
		return nil, false
	}
	var st sessionState
	if err := json.Unmarshal([]byte(raw), &st); err != nil || st.SessionID == "" {
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
		return
	}
	_ = dao.RedisDao.SetKey(sessionKey(userID), string(b), ttl)
}

// dropSession 结束会话
func dropSession(userID int64) {
	_ = dao.RedisDao.DelKey(sessionKey(userID))
}

// allowRate 每用户每分钟限流；Redis 异常时放行（不因缓存故障阻断用户）
func allowRate(userID int64, limit int) bool {
	if limit <= 0 {
		return true
	}
	key := rateKey(userID)
	n, err := dao.RedisDao.INCR(key)
	if err != nil {
		return true
	}
	if n == 1 {
		_ = dao.RedisDao.SetKey(key, n, time.Minute)
	}
	return n <= int64(limit)
}

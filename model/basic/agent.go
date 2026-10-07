package model

// ==================== Agent（失物招领对话助手）API 契约 ====================
// 说明：Agent 的 DTO 放在 model/basic（而非 model/advanced），因为响需要复用
// ItemResponse/Location/Tag；model/basic 已依赖 model/advanced，反向依赖会成环。

// Agent 会话阶段（响应字段 stage）
const (
	AgentStageChitchat    = "chitchat"     // 闲聊/无关，已给出固定话术
	AgentStageNoMatch     = "no_match"     // 未匹配到相关帖子
	AgentStageMatched     = "matched"      // 已匹配到相关帖子
	AgentStageNeedConfirm = "need_confirm" // 草稿待确认（唯一一次「补充信息 + 确认」）
	AgentStageCreated     = "created"      // 已创建完成
	AgentStageCancelled   = "cancelled"    // 用户放弃/会话结束
)

// Agent 请求动作（请求字段 action）
const (
	AgentActionAuto    = "auto"    // 默认：按文本内容自动处理
	AgentActionConfirm = "confirm" // 确认发布当前草稿
	AgentActionCancel  = "cancel"  // 放弃当前草稿
)

// AgentChatRequest 会话主入口请求
type AgentChatRequest struct {
	SessionID string   `json:"session_id,omitempty"` // 空 = 新会话；否则续会话
	Text      string   `json:"text" binding:"required"`
	ImageURLs []string `json:"image_urls,omitempty"` // 可选 ≤3；相对路径（以 / 开头）由后端拼 agent_public_base_url 供多模态；建帖成功时自动绑定为物品图片
	Action    string   `json:"action,omitempty"`     // auto | confirm | cancel
}

// AgentMatchRequest 无状态匹配请求（只读）
type AgentMatchRequest struct {
	Text      string   `json:"text" binding:"required"`
	ImageURLs []string `json:"image_urls,omitempty"`
	TopN      int      `json:"top_n,omitempty"` // 返回条数，默认 3，最大 10
}

// AgentExtractRequest 无状态抽取请求（只读，供前端智能填充）
type AgentExtractRequest struct {
	Text      string   `json:"text" binding:"required"`
	ImageURLs []string `json:"image_urls,omitempty"`
}

// AgentSessionCloseRequest 关闭会话请求
type AgentSessionCloseRequest struct {
	SessionID string `json:"session_id,omitempty"`
}

// AgentDraft 物品草稿（前端可直接渲染成表单；QQ 侧由代码渲染为文案）
type AgentDraft struct {
	Type           int8     `json:"type"`                      // 0 失物 / 1 招领
	Title          string   `json:"title"`                     // ≤100 字节
	Description    string   `json:"description"`               // ≤200 字
	TagIDs         []int64  `json:"tag_ids"`                   // 可直接提交给 /item/create
	TagNames       []string `json:"tag_names,omitempty"`       // 展示用
	LocationID     *int64   `json:"location_id,omitempty"`     // 缺省时由后端填 140「其他地点」
	LocationName   string   `json:"location_name,omitempty"`   // 展示用（地点链路）
	LocationDetail *string  `json:"location_detail,omitempty"` // 只写链路表达不了的细节
	Contact        *string  `json:"contact,omitempty"`         // 仅当用户明确给出联系方式时才有值（不编造）
	LostFoundTime  string   `json:"lost_found_time,omitempty"` // RFC3339
	TimeFrom       string   `json:"time_from,omitempty"`       // 抽取的时间区间（展示/调试用）
	TimeTo         string   `json:"time_to,omitempty"`
	MissingFields  []string `json:"missing_fields,omitempty"` // location/location_detail/time/contact/color/features/item
	Followup       string   `json:"followup_question,omitempty"`
}

// AgentMatchBrief 单条匹配结果（Item 复用既有 ItemResponse，含 contact，与详情接口口径一致）
type AgentMatchBrief struct {
	ItemID  int64        `json:"item_id"`
	Score   float64      `json:"score"`
	Reasons []string     `json:"reasons,omitempty"`
	Risk    []string     `json:"risk,omitempty"`
	Item    ItemResponse `json:"item"`
}

// AgentChatResponse 会话主入口响应
type AgentChatResponse struct {
	SessionID     string            `json:"session_id"`
	Stage         string            `json:"stage"`
	Reply         string            `json:"reply"`
	Intent        string            `json:"intent,omitempty"`
	Questions     []string          `json:"questions"`
	Matches       []AgentMatchBrief `json:"matches"`
	Similar       []AgentMatchBrief `json:"similar"`
	Draft         *AgentDraft       `json:"draft,omitempty"`
	CreatedItemID *int64            `json:"created_item_id,omitempty"`
}

// AgentMatchResponse 无状态匹配响应
type AgentMatchResponse struct {
	Intent   string            `json:"intent"`
	Entities *AgentDraft       `json:"entities,omitempty"`
	Matches  []AgentMatchBrief `json:"matches"`
	Similar  []AgentMatchBrief `json:"similar"`
	Summary  string            `json:"summary,omitempty"`
	Verdict  string            `json:"verdict,omitempty"`
}

// AgentExtractResponse 无状态抽取响应
type AgentExtractResponse struct {
	Intent        string      `json:"intent"`
	IsLNFContext  bool        `json:"is_lnf_context"`
	Draft         *AgentDraft `json:"draft,omitempty"`
	MissingFields []string    `json:"missing_fields"`
	Questions     []string    `json:"questions"`
}

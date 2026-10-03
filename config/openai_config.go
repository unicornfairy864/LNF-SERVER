package config

import "time"

// OpenAIConfig LLM 配置（config.yaml 的 openai 段）
// Agent 相关字段：失物招领对话助手（API + QQBOT 共用），详见 agent.md/agent.md §12
type OpenAIConfig struct {
	OpenaiKey     string  `mapstructure:"openai_key"`
	OpenaiBaseUrl string  `mapstructure:"openai_base_url"`
	DefaultModel  string  `mapstructure:"default_model"`
	Temperature   float64 `mapstructure:"temperature"`

	// ==================== Agent 配置 ====================
	AgentEnabled             bool          `mapstructure:"agent_enabled"`                     // 总开关；false 时 API/QQ 链路直接短路
	AgentTimeout             time.Duration `mapstructure:"agent_timeout"`                     // 单次 LLM 调用超时
	AgentMaxInputChars       int           `mapstructure:"agent_max_input_chars"`             // 单条用户输入字符上限（超出拦截）
	AgentSessionTTL          time.Duration `mapstructure:"agent_session_ttl"`                 // 会话有效期（每次交互滑动续期）
	AgentTopK                int           `mapstructure:"agent_top_k"`                       // 送 LLM 精排的候选上限
	AgentStrongThreshold     float64       `mapstructure:"agent_strong_threshold"`            // 强匹配阈值
	AgentAmbiguousThreshold  float64       `mapstructure:"agent_ambiguous_threshold"`         // 模糊匹配下限
	AgentFollowupMaxRounds   int           `mapstructure:"agent_followup_max_rounds"`         // 追问/确认轮数上限
	AgentRateLimitTotalMin   int           `mapstructure:"agent_rate_limit_total_per_minute"` // 全系统每分钟调用上限（所有用户合计）
	AgentRateLimitPerMinute  int           `mapstructure:"agent_rate_limit_per_minute"`       // 每个用户每分钟调用上限
	AgentMatchMinScore       int           `mapstructure:"agent_match_min_score"`             // 召回严格模式最低分（location 0/1 + tag 命中数）
	AgentMatchTimeBeforeDays int           `mapstructure:"agent_match_time_before_days"`      // 时间窗向前容差（天）
	AgentMatchTimeWindowDays int           `mapstructure:"agent_match_time_window_days"`      // 时间窗总跨度上限（天）
	AgentDefaultLocationID   int64         `mapstructure:"agent_default_location_id"`         // 缺地点时默认地点 ID（「其他地点」，DB 中为 140）
	AgentPublicBaseURL       string        `mapstructure:"agent_public_base_url"`             // 多模态：图片相对路径（以 "/" 开头）拼接前缀
	// AgentFillContact 是否把「用户明确给出的」联系方式写入 items.contact（不编造、不推断）。
	// 指针类型：缺省（未配置）视为 true，显式 false 关闭。
	AgentFillContact         *bool `mapstructure:"agent_fill_contact"`
	AgentReverseMatchEnabled bool  `mapstructure:"agent_reverse_match_enabled"` // 反向匹配推送开关（批次 4）
	AgentReverseMatchDays    int   `mapstructure:"agent_reverse_match_days"`    // 反向匹配回溯天数（批次 4）
}

// Agent 配置缺省值（字段缺省/非法时回退，见 AgentValues）
const (
	AgentDefaultTimeout             = 30 * time.Second
	AgentDefaultMaxInputChars       = 500
	AgentDefaultSessionTTL          = 30 * time.Minute
	AgentDefaultTopK                = 20
	AgentDefaultStrongThreshold     = 0.80
	AgentDefaultAmbiguousThreshold  = 0.50
	AgentDefaultFollowupMaxRounds   = 1
	AgentDefaultRateLimitPerMinute  = 3
	AgentDefaultRateLimitTotalMin   = 30
	AgentDefaultMatchMinScore       = 2
	AgentDefaultMatchTimeBeforeDays = 1
	AgentDefaultMatchTimeWindowDays = 30
	AgentDefaultLocationID          = int64(140)
	AgentDefaultReverseMatchDays    = 30
	AgentDefaultFillContact         = true
)

// AgentSettings Agent 配置的归一化结果：零值/非法值一律回退缺省，业务代码只读该结构，不再做兜底判断
type AgentSettings struct {
	Enabled             bool
	Timeout             time.Duration
	MaxInputChars       int
	SessionTTL          time.Duration
	TopK                int
	StrongThreshold     float64
	AmbiguousThreshold  float64
	FollowupMaxRounds   int
	RateLimitPerMinute  int
	RateLimitTotalMin   int
	MatchMinScore       int
	MatchTimeBeforeDays int
	MatchTimeWindowDays int
	DefaultLocationID   int64
	PublicBaseURL       string
	FillContact         bool
	ReverseMatchEnabled bool
	ReverseMatchDays    int
}

// AgentValues 返回归一化后的 Agent 配置
func (c OpenAIConfig) AgentValues() AgentSettings {
	s := AgentSettings{
		Enabled:             c.AgentEnabled,
		Timeout:             c.AgentTimeout,
		MaxInputChars:       c.AgentMaxInputChars,
		SessionTTL:          c.AgentSessionTTL,
		TopK:                c.AgentTopK,
		StrongThreshold:     c.AgentStrongThreshold,
		AmbiguousThreshold:  c.AgentAmbiguousThreshold,
		FollowupMaxRounds:   c.AgentFollowupMaxRounds,
		RateLimitPerMinute:  c.AgentRateLimitPerMinute,
		RateLimitTotalMin:   c.AgentRateLimitTotalMin,
		MatchMinScore:       c.AgentMatchMinScore,
		MatchTimeBeforeDays: c.AgentMatchTimeBeforeDays,
		MatchTimeWindowDays: c.AgentMatchTimeWindowDays,
		DefaultLocationID:   c.AgentDefaultLocationID,
		PublicBaseURL:       c.AgentPublicBaseURL,
		ReverseMatchEnabled: c.AgentReverseMatchEnabled,
		ReverseMatchDays:    c.AgentReverseMatchDays,
	}
	if s.Timeout <= 0 {
		s.Timeout = AgentDefaultTimeout
	}
	if s.MaxInputChars <= 0 {
		s.MaxInputChars = AgentDefaultMaxInputChars
	}
	if s.SessionTTL <= 0 {
		s.SessionTTL = AgentDefaultSessionTTL
	}
	if s.TopK <= 0 {
		s.TopK = AgentDefaultTopK
	}
	if s.StrongThreshold <= 0 || s.StrongThreshold > 1 {
		s.StrongThreshold = AgentDefaultStrongThreshold
	}
	if s.AmbiguousThreshold <= 0 || s.AmbiguousThreshold >= s.StrongThreshold {
		s.AmbiguousThreshold = AgentDefaultAmbiguousThreshold
	}
	if s.FollowupMaxRounds <= 0 {
		s.FollowupMaxRounds = AgentDefaultFollowupMaxRounds
	}
	if s.RateLimitPerMinute <= 0 {
		s.RateLimitPerMinute = AgentDefaultRateLimitPerMinute
	}
	if s.RateLimitTotalMin <= 0 {
		s.RateLimitTotalMin = AgentDefaultRateLimitTotalMin
	}
	if s.MatchMinScore <= 0 {
		s.MatchMinScore = AgentDefaultMatchMinScore
	}
	if s.MatchTimeBeforeDays < 0 {
		s.MatchTimeBeforeDays = AgentDefaultMatchTimeBeforeDays
	}
	if s.MatchTimeWindowDays <= 0 {
		s.MatchTimeWindowDays = AgentDefaultMatchTimeWindowDays
	}
	if s.DefaultLocationID <= 0 {
		s.DefaultLocationID = AgentDefaultLocationID
	}
	if s.ReverseMatchDays <= 0 {
		s.ReverseMatchDays = AgentDefaultReverseMatchDays
	}
	s.PublicBaseURL = trimTrailingSlash(s.PublicBaseURL)
	// 联系方式：缺省开启（仅写入用户明确给出的值，不编造）；显式 false 关闭
	if c.AgentFillContact == nil {
		s.FillContact = AgentDefaultFillContact
	} else {
		s.FillContact = *c.AgentFillContact
	}
	return s
}

// trimTrailingSlash 去掉末尾 "/"，避免与以 "/" 开头的相对图片路径拼接出双斜杠
func trimTrailingSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

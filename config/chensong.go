package config

import "time"

type ChenSongConfig struct {
	ApiUrl         string        `mapstructure:"api_url"`
	ApiToken       string        `mapstructure:"api_token"`
	ActivatedQQ    int64         `mapstructure:"activated_qq"`
	ActivatedGroup int64         `mapstructure:"activated_group"`
	BindMaxTries   int64         `mapstructure:"bind_max_tries"`
	BindTimeout    time.Duration `mapstructure:"bind_timeout"`
	ReceiveToken   string        `mapstructure:"receive_token"`

	// ==================== LNF Agent（QQ 侧）配置，见 agent.md/agent.md §6 ====================
	// 说明（2026-10-04 定稿）：触发**不再依赖关键词**（关键词过滤已删除），成本由下面两项兜底
	LnfCooldown           time.Duration `mapstructure:"lnf_cooldown"`              // 同一 QQ 触发冷却
	LnfGroupRatePerMinute int           `mapstructure:"lnf_group_rate_per_minute"` // 同一群每分钟最多 agent 回复条数
}

// LNF Agent（QQ 侧）缺省值
const (
	LnfDefaultCooldown           = 60 * time.Second
	LnfDefaultGroupRatePerMinute = 3
)

// LnfSettings QQ 侧归一化配置
type LnfSettings struct {
	Cooldown           time.Duration
	GroupRatePerMinute int
}

// LnfValues 返回归一化后的 QQ 侧配置
func (c ChenSongConfig) LnfValues() LnfSettings {
	s := LnfSettings{
		Cooldown:           c.LnfCooldown,
		GroupRatePerMinute: c.LnfGroupRatePerMinute,
	}
	if s.Cooldown <= 0 {
		s.Cooldown = LnfDefaultCooldown
	}
	if s.GroupRatePerMinute <= 0 {
		s.GroupRatePerMinute = LnfDefaultGroupRatePerMinute
	}
	return s
}

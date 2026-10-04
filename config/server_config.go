package config

import "time"

type ServerConfig struct {
	Port              int           `mapstructure:"port"`
	RouterPrefix      string        `mapstructure:"router_prefix"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`

	// Claim 认领功能配置
	ClaimQQRequired bool          `mapstructure:"claim_qq_required"` // 认领是否要求用户已绑定QQ（qq非空检查开关）
	ClaimCredit     int64         `mapstructure:"claim_credit"`      // 认领成功每次加的积分数量
	ClaimAutoClose  time.Duration `mapstructure:"claim_auto_close"`  // 认领后自动关闭时长，<=0 表示不自动关闭
	// 认领超时前多久提醒发帖人（缺省 2h；<=0 用缺省；>=ClaimAutoClose 则不提醒）
	ClaimAutoCloseRemindBefore time.Duration `mapstructure:"claim_auto_close_remind_before"`
}

// ClaimDefaultAutoCloseRemindBefore 认领超时提醒的缺省提前量
const ClaimDefaultAutoCloseRemindBefore = 2 * time.Hour

// ClaimRemindBeforeValue 归一化后的提醒提前量
func (c ServerConfig) ClaimRemindBeforeValue() time.Duration {
	if c.ClaimAutoCloseRemindBefore <= 0 {
		return ClaimDefaultAutoCloseRemindBefore
	}
	return c.ClaimAutoCloseRemindBefore
}

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
}

package config

import "time"

type ServerConfig struct {
	Port              int           `mapstructure:"port"`
	RouterPrefix      string        `mapstructure:"router_prefix"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
}

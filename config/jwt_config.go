package config

import "time"

type JWTConfig struct {
	SigningKey  string        `mapstructure:"signing_key"`
	ExpiresTime time.Duration `mapstructure:"expires_time"`
	BufferTime  time.Duration `mapstructure:"buffer_time"`
	Issuer      string        `mapstructure:"issuer"`
}
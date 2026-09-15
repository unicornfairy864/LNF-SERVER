package config

type ServerConfig struct {
	Port int `mapstructure:"port"`
	RouterPrefix string `mapstructure:"router_prefix"`
}
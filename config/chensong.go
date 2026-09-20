package config

import "time"

type ChenSongConfig struct {
	ApiUrl         string        `mapstructure:"api_url"`
	ApiToken       string        `mapstructure:"api_token"`
	ActivatedQQ    int64         `mapstructure:"activated_qq"`
	ActivatedGroup int64         `mapstructure:"activated_group"`
	BindMaxTries   int           `mapstructure:"bind_max_tries"`
	BindTimeout    time.Duration `mapstructure:"bind_timeout"`
	ReceiveToken   string        `mapstructure:"receive_token"`
}

package config

type ChenSongConfig struct {
	ApiUrl         string `mapstructure:"api_url"`
	ApiToken       string `mapstructure:"api_token"`
	ActivatedQQ    int    `mapstructure:"activated_qq"`
	ActivatedGroup int    `mapstructure:"activated_group"`
}

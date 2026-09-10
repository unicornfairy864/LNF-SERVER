package config

type ApiConfig struct {
	Openai_key      string `mapstructure:"openai_key"`
	Openai_base_url string `mapstructure:"openai_base_url"`
	Default_model   string `mapstructure:"default_model"`
}
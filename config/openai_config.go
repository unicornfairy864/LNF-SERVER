package config

type ApiConfig struct {
	OpenaiKey     string `mapstructure:"openai_key"`
	OpenaiBaseUrl string `mapstructure:"openai_base_url"`
	DefaultModel  string `mapstructure:"default_model"`
}

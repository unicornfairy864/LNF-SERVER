package config

type OpenAIConfig struct {
	OpenaiKey     string  `mapstructure:"openai_key"`
	OpenaiBaseUrl string  `mapstructure:"openai_base_url"`
	DefaultModel  string  `mapstructure:"default_model"`
	Temperature   float64 `mapstructure:"temperature"`
}

package config

type Config struct {
	OpenAI   OpenAIConfig   `mapstructure:"openai"`
	Mysql    MysqlConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Server   ServerConfig   `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	ChenSong ChenSongConfig `mapstructure:"chensong"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

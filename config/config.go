package config

type Config struct {
	Api      ApiConfig      `mapstructure:"api"`
	Mysql    MysqlConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Server   ServerConfig   `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	ChenSong ChenSongConfig `mapstructure:"chensong"`
}

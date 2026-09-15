package config

type Config struct {
	Api ApiConfig `mapstructure:"api"`
	Mysql MysqlConfig `mapstructure:"mysql"`
	Server ServerConfig `mapstructure:"server"`
}
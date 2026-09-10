package config

type Config struct {
	Api ApiConfig `mapstructure:"api"`
	Mysql MysqlConfig `mapstructure:"mysql"`
}
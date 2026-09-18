package config

type MysqlConfig struct {
	DatabaseHost     string `mapstructure:"database_host"`
	DatabasePort     string `mapstructure:"database_port"`
	DatabaseDbname   string `mapstructure:"database_dbname"`
	DatabaseUser     string `mapstructure:"database_user"`
	DatabasePassword string `mapstructure:"database_password"`
}

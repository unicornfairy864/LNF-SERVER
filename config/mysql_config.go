package config

type MysqlConfig struct {
	Database_host     string `mapstructure:"database_host"`
	Database_port     string `mapstructure:"database_port"`
	Database_dbname   string `mapstructure:"database_dbname"`
	Database_user     string `mapstructure:"database_user"`
	Database_password string `mapstructure:"database_password"`
}
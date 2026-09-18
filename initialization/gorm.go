package initialization

import (
	"fmt"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	// Set up gorm connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		global.LNF_CONFIG.Mysql.DatabaseUser,
		global.LNF_CONFIG.Mysql.DatabasePassword,
		global.LNF_CONFIG.Mysql.DatabaseHost,
		global.LNF_CONFIG.Mysql.DatabasePort,
		global.LNF_CONFIG.Mysql.DatabaseDbname,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Set sqlDB configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(30 * time.Second)

	// The program can't automatically reconnect database

	return db, nil
}

func CloseDB() error {
	if global.LNF_DB == nil {
		return nil
	}
	sqlDB, err := global.LNF_DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

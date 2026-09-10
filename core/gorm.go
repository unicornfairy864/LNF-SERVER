package core

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
		global.LNF_CONFIG.Mysql.Database_user,
		global.LNF_CONFIG.Mysql.Database_password,
		global.LNF_CONFIG.Mysql.Database_host,
		global.LNF_CONFIG.Mysql.Database_port,
		global.LNF_CONFIG.Mysql.Database_dbname,
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

	// The program can't automatically reconnect

	return db, nil
}
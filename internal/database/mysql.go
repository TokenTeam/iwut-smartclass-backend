package database

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"iwut-smartclass-backend/internal/config"
	"iwut-smartclass-backend/internal/middleware"
	"time"
)

type MySQL struct {
	Database *sql.DB
}

var dbInstance *MySQL

func NewDB(cfg *config.Config) error {
	dsn := fmt.Sprintf(cfg.Database)

	// 连接数据库
	start := time.Now()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to connect to MySQL: %v", err))
		return err
	}

	// 检查数据库连接
	if err = db.Ping(); err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to ping MySQL: %v", err))
		return err
	}
	elapsed := time.Since(start)
	middleware.Logger.Log("INFO", fmt.Sprintf("[DB] Connected to MySQL in %s", elapsed))

	// 使用 GORM 自动迁移表结构
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] GORM connect failed: %v", err))
		return err
	}
	if err = gormDB.AutoMigrate(&Course{}); err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] AutoMigrate failed: %v", err))
		return err
	}
	middleware.Logger.Log("INFO", "[DB] AutoMigrate completed")

	dbInstance = &MySQL{Database: db}
	return nil
}

func GetDB() *sql.DB {
	if dbInstance == nil {
		return nil
	}
	return dbInstance.Database
}

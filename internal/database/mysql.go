package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"iwut-smartclass-backend/internal/config"
	"iwut-smartclass-backend/internal/middleware"
	"os"
	"time"
)

type MySQL struct {
	Database *sql.DB
}

var dbInstance *MySQL

// NewDB 建立数据库连接
func NewDB(cfg *config.Config) error {
	dsn := fmt.Sprintf(cfg.Database)

	// 连接数据库
	start := time.Now()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to connect to MySQL: %s", err))
		return err
	}

	// 检查数据库连接
	if err := db.Ping(); err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to ping MySQL: %s", err))
		return err
	}
	elapsed := time.Since(start)
	middleware.Logger.Log("INFO", fmt.Sprintf("[DB] Connected to MySQL in %s", elapsed))

	// 检查数据库是否为空
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&count)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to check: %s", err))
		return err
	}

	// 如果数据库为空，则初始化数据库
	if count == 0 {
		middleware.Logger.Log("INFO", "[DB] Database is empty, initializing...")
		err = initializeDatabase(db)
		if err != nil {
			return err
		}
	}

	dbInstance = &MySQL{Database: db}
	return nil
}

// GetDB 获取数据库连接
func GetDB() *sql.DB {
	if dbInstance == nil {
		return nil
	}
	return dbInstance.Database
}

// initializeDatabase 初始化数据库
func initializeDatabase(db *sql.DB) error {
	file, err := os.Open("db.sql")
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to open db.sql: %s", err))
		return err
	}
	defer file.Close()

	sqlBytes, err := io.ReadAll(file)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to read db.sql: %s", err))
		return err
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to execute db.sql: %s", err))
		return err
	}

	middleware.Logger.Log("INFO", "[DB] Database initialized")
	return nil
}

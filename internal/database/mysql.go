package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"iwut-smart-timetable-backend/internal/config"
	"iwut-smart-timetable-backend/internal/middleware"
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
		middleware.Logger.Log("ERROR", "[DB] Failed to connect to MySQL: "+err.Error())
		return err
	}

	// 检查数据库连接
	if err := db.Ping(); err != nil {
		middleware.Logger.Log("ERROR", "[DB] Failed to ping MySQL: "+err.Error())
		return err
	}
	elapsed := time.Since(start)
	middleware.Logger.Log("INFO", fmt.Sprintf("[DB] Connected to MySQL in %s", elapsed))

	// 检查数据库是否为空
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&count)
	if err != nil {
		middleware.Logger.Log("ERROR", "[DB] Failed to check: "+err.Error())
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
		middleware.Logger.Log("ERROR", "[DB] Failed to open db.sql: "+err.Error())
		return err
	}
	defer file.Close()

	sqlBytes, err := io.ReadAll(file)
	if err != nil {
		middleware.Logger.Log("ERROR", "[DB] Failed to read db.sql: "+err.Error())
		return err
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		middleware.Logger.Log("ERROR", "[DB] Failed to execute db.sql: "+err.Error())
		return err
	}

	middleware.Logger.Log("INFO", "[DB] Database initialized")
	return nil
}

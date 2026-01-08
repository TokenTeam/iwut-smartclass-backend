package database

import (
	"context"
	"fmt"
	"iwut-smartclass-backend/internal/config"
	"iwut-smartclass-backend/internal/middleware"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQL struct {
	Database *gorm.DB
}

var dbInstance *MySQL

func NewDB(cfg *config.Config) error {
	dsn := cfg.Database

	// 连接数据库
	start := time.Now()

	// 使用 GORM 连接数据库
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] GORM connect failed: %v", err))
		return err
	}

	// 获取底层的 sql.DB 以配置连接池参数
	sqlDB, err := gormDB.DB()
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to get underlying sql.DB: %v", err))
		return err
	}

	// 配置连接池参数
	// SetMaxOpenConns: 设置打开数据库连接的最大数量
	// SetMaxIdleConns: 设置空闲连接池中连接的最大数量
	// SetConnMaxLifetime: 设置连接可复用的最大时间（避免连接被服务器关闭）
	// SetConnMaxIdleTime: 设置连接在连接池中空闲的最大时间（超过此时间会被关闭）
	sqlDB.SetMaxOpenConns(25)                 // 最大打开连接数
	sqlDB.SetMaxIdleConns(10)                 // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // 连接最大生命周期（5分钟，小于MySQL的wait_timeout）
	sqlDB.SetConnMaxIdleTime(2 * time.Minute) // 连接最大空闲时间（2分钟）

	// 检查数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = sqlDB.PingContext(ctx); err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] Failed to ping MySQL: %v", err))
		return err
	}
	elapsed := time.Since(start)
	middleware.Logger.Log("INFO", fmt.Sprintf("[DB] Connected to MySQL in %s", elapsed))

	// 自动迁移表结构
	for _, s := range Structs {
		if err = gormDB.AutoMigrate(s); err != nil {
			middleware.Logger.Log("ERROR", fmt.Sprintf("[DB] AutoMigrate failed: %v", err))
			return err
		}
	}
	middleware.Logger.Log("INFO", "[DB] AutoMigrate completed")

	dbInstance = &MySQL{Database: gormDB}
	return nil
}

func GetDB() *gorm.DB {
	if dbInstance == nil {
		return nil
	}
	return dbInstance.Database
}

// GetDBWithContext 返回带超时的数据库连接上下文
// 默认查询超时时间为10秒
func GetDBWithContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	// 设置默认查询超时时间为10秒
	return context.WithTimeout(ctx, 10*time.Second)
}

// PingDB 检查数据库连接是否健康
func PingDB() error {
	if dbInstance == nil || dbInstance.Database == nil {
		return fmt.Errorf("database instance is nil")
	}

	sqlDB, err := dbInstance.Database.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}

// EnsureConnection 确保数据库连接健康，如果连接失败则尝试重新连接
func EnsureConnection() error {
	if err := PingDB(); err != nil {
		middleware.Logger.Log("WARN", fmt.Sprintf("[DB] Connection unhealthy, attempting to reconnect: %v", err))
		// 尝试ping，如果失败则记录错误，但不强制重连（让连接池自己处理）
		// 连接池会自动处理失效的连接
		return err
	}
	return nil
}

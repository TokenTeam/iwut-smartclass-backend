package middleware

import (
	"fmt"
	"io"
	"iwut-smartclass-backend/internal/config"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Logger 全局日志记录器
var Logger *Log

type Log struct {
	Config     *config.Config
	file       *os.File
	consoleLog *log.Logger
	fileLog    *log.Logger
}

// NewLogger 创建日志记录器
func NewLogger(cfg *config.Config) *Log {
	logger := &Log{Config: cfg}
	logger.splitLogFile()
	go logger.scheduleLogSplit()
	return logger
}

// Log 日志记录器
func (l *Log) Log(level, message string) {
	if l.Config.Debug || level != "DEBUG" {
		_, file, line, _ := runtime.Caller(2)
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		color := setColor(level)
		reset := "\033[0m"
		logMessage := fmt.Sprintf("%s[%s] %s %s (%s:%d)%s", color, level, timestamp, message, filepath.Base(file), line, reset)
		l.consoleLog.Println(logMessage)
		l.fileLog.Printf("[%s] %s %s (%s:%d)", level, timestamp, message, filepath.Base(file), line)
	}
}

// HttpMiddleware HTTP 调试日志中间件
func (l *Log) HttpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.Log("DEBUG", fmt.Sprintf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr))
		next.ServeHTTP(w, r)
	})
}

// splitLogFile 切割日志文件
func (l *Log) splitLogFile() {
	if l.file != nil {
		l.file.Close()
	}

	logDir := "log"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, 0755)
	}

	filename := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		l.Log("ERROR", fmt.Sprintf("Failed to open log file: %v", err))
	}

	l.file = file
	l.consoleLog = log.New(io.MultiWriter(os.Stdout), "", 0)
	l.fileLog = log.New(file, "", 0)
}

// scheduleLogSplit 按日计划切割日志
func (l *Log) scheduleLogSplit() {
	for {
		now := time.Now()
		next := now.AddDate(0, 0, 1)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		duration := next.Sub(now)
		time.Sleep(duration)
		l.splitLogFile()
	}
}

// setColor 设置输出颜色
func setColor(level string) string {
	switch level {
	case "INFO":
		return "\033[32m" // 绿
	case "WARN":
		return "\033[33m" // 黄
	case "ERROR":
		return "\033[31m" // 红
	case "DEBUG":
		return "\033[34m" // 蓝
	default:
		return "\033[0m"
	}
}

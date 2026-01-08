package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Field 日志字段接口
type Field interface {
	Key() string
	Value() interface{}
}

// String 创建字符串字段
func String(key, value string) Field {
	return &stringField{key: key, value: value}
}

type stringField struct {
	key   string
	value string
}

func (f *stringField) Key() string   { return f.key }
func (f *stringField) Value() interface{} { return f.value }

// NewLogger 创建新的日志实例
func NewLogger(cfg *Config) (Logger, error) {
	return NewSimpleLogger(cfg), nil
}

// Config 日志配置
type Config struct {
	Debug   bool
	LogSave bool
}


// 兼容旧接口的简单日志实现（临时，用于迁移）
type simpleLogger struct {
	writer io.Writer
	debug  bool
}

func NewSimpleLogger(cfg *Config) Logger {
	return &simpleLogger{
		writer: os.Stdout,
		debug:  cfg.Debug,
	}
}

func (l *simpleLogger) Debug(msg string, fields ...Field) {
	if l.debug {
		l.log("DEBUG", msg, fields...)
	}
}

func (l *simpleLogger) Info(msg string, fields ...Field) {
	l.log("INFO", msg, fields...)
}

func (l *simpleLogger) Warn(msg string, fields ...Field) {
	l.log("WARN", msg, fields...)
}

func (l *simpleLogger) Error(msg string, fields ...Field) {
	l.log("ERROR", msg, fields...)
}

func (l *simpleLogger) With(fields ...Field) Logger {
	return l
}

func (l *simpleLogger) log(level, msg string, fields ...Field) {
	timestamp := time.Now().Format(time.RFC3339)
	logMsg := "[" + level + "] " + timestamp + " " + msg
	if len(fields) > 0 {
		logMsg += " " + formatFields(fields)
	}
	logMsg += "\n"
	l.writer.Write([]byte(logMsg))
}

func formatFields(fields []Field) string {
	if len(fields) == 0 {
		return ""
	}
	result := ""
	for i, f := range fields {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%s=%v", f.Key(), f.Value())
	}
	return result
}

package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	Debug                bool
	Port                 string
	Database             string
	LogSave              bool
	SummaryWorkerCount   int
	SummaryQueueSize     int
	TencentSecretId      []string
	TencentSecretKey     []string
	BucketUrl            string
	OpenaiEndpoint       string
	OpenaiKey            string
	OpenaiModel          string
	Temperature          float32
	InfoSimple           string
	GetWeekSchedules     string
	SearchLiveCourseList string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Debug:                false,
		Port:                 "8080",
		Database:             "",
		LogSave:              false,
		SummaryWorkerCount:   2,
		SummaryQueueSize:     20,
		TencentSecretId:      []string{},
		TencentSecretKey:     []string{},
		BucketUrl:            "",
		OpenaiEndpoint:       "",
		OpenaiKey:            "",
		OpenaiModel:          "",
		Temperature:          0.3,
		InfoSimple:           "",
		GetWeekSchedules:     "",
		SearchLiveCourseList: "",
	}
}

// LoadConfig 加载配置（可选指定 .env 路径）
func LoadConfig(envPath string) (*Config, error) {
	config := DefaultConfig()

	// 从 .env 文件加载配置
	if envPath != "" {
		_ = godotenv.Load(envPath)
	} else {
		_ = godotenv.Load()
	}

	// 从环境变量加载
	if err := LoadConfigFromEnv(config); err != nil {
		return nil, err
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv(config *Config) error {
	val := reflect.ValueOf(config).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name
		envName := fieldNameToEnvName(fieldName)
		envVal := os.Getenv(envName)

		if envVal == "" {
			continue
		}

		switch fieldName {
		case "TencentSecretId", "TencentSecretKey":
			values := strings.Split(envVal, ",")
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
			field.Set(reflect.ValueOf(values))
		case "Temperature":
			if floatVal, err := strconv.ParseFloat(envVal, 32); err == nil {
				field.SetFloat(floatVal)
			}
		default:
			switch field.Kind() {
			case reflect.String:
				field.SetString(envVal)
			case reflect.Bool:
				field.SetBool(envVal == "true")
			case reflect.Int, reflect.Int64:
				if intVal, err := strconv.Atoi(envVal); err == nil {
					field.SetInt(int64(intVal))
				}
			}
		}
	}

	return nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Database == "" {
		return &ValidationError{Field: "Database", Message: "database connection string is required"}
	}
	if len(c.TencentSecretId) == 0 {
		return &ValidationError{Field: "TencentSecretId", Message: "tencent secret id is required"}
	}
	if len(c.TencentSecretKey) == 0 {
		return &ValidationError{Field: "TencentSecretKey", Message: "tencent secret key is required"}
	}
	if c.OpenaiEndpoint == "" {
		return &ValidationError{Field: "OpenaiEndpoint", Message: "openai endpoint is required"}
	}
	if c.OpenaiKey == "" {
		return &ValidationError{Field: "OpenaiKey", Message: "openai key is required"}
	}
	return nil
}

// ValidationError 配置验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func fieldNameToEnvName(name string) string {
	var result strings.Builder
	for i, char := range name {
		if i > 0 && 'A' <= char && char <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(char)
	}
	return strings.ToUpper(result.String())
}

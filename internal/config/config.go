package config

import (
	"github.com/joho/godotenv"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Config struct {
	Debug                bool
	Port                 string
	Database             string
	SummaryWorkerCount   int
	SummaryQueueSize     int
	TencentSecretId      []string
	TencentSecretKey     []string
	BucketUrl            string
	OpenaiEndpoint       string
	OpenaiKey            string
	OpenaiModel          string
	InfoSimple           string
	GetWeekSchedules     string
	SearchLiveCourseList string
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Debug:                false,
		Port:                 "8080",
		Database:             "",
		SummaryWorkerCount:   2,
		SummaryQueueSize:     20,
		TencentSecretId:      []string{},
		TencentSecretKey:     []string{},
		BucketUrl:            "",
		OpenaiEndpoint:       "",
		OpenaiKey:            "",
		OpenaiModel:          "",
		InfoSimple:           "",
		GetWeekSchedules:     "",
		SearchLiveCourseList: "",
	}
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	// 获取默认配置
	config := DefaultConfig()

	// 从 .env 文件加载配置
	_ = godotenv.Load()

	// 自动处理环境变量
	return LoadConfigFromEnv(config)
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv(config *Config) *Config {
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
			// 处理逗号分隔的字符串
			values := strings.Split(envVal, ",")
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
			field.Set(reflect.ValueOf(values))
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

	return config
}

// fieldNameToEnvName 将结构体字段名转换为环境变量名
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

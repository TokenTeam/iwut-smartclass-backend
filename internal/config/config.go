package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Debug            bool
	Port             string
	Database         string
	TencentSecretId  string
	TencentSecretKey string
	BucketUrl        string
	OpenaiEndpoint   string
	OpenaiKey        string
	OpenaiModel      string
	Prompt           string
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	debug := os.Getenv("DEBUG") == "true"
	database := os.Getenv("DATABASE")
	tencentSecretId := os.Getenv("TENCENT_SECRET_ID")
	tencentSecretKey := os.Getenv("TENCENT_SECRET_KEY")
	bucketUrl := os.Getenv("BUCKET_URL")
	openaiEndpoint := os.Getenv("OPENAI_ENDPOINT")
	openaiKey := os.Getenv("OPENAI_KEY")
	openaiModel := os.Getenv("OPENAI_MODEL")
	prompt := os.Getenv("PROMPT")

	return &Config{
		Debug:            debug,
		Port:             port,
		Database:         database,
		TencentSecretId:  tencentSecretId,
		TencentSecretKey: tencentSecretKey,
		BucketUrl:        bucketUrl,
		OpenaiEndpoint:   openaiEndpoint,
		OpenaiKey:        openaiKey,
		OpenaiModel:      openaiModel,
		Prompt:           prompt,
	}
}

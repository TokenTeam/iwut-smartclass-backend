package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Debug    bool
	Port     string
	Database string
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

	return &Config{
		Debug:    debug,
		Port:     port,
		Database: database,
	}
}

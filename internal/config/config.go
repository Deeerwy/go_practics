package config

import (
	"log"
	"os"
)

type Config struct {
	Port        string
	AuthBaseURL string
	RedisAddr   string
	LogLevel    string
}

func Load() *Config {
	cfg := &Config{
		Port:        getEnv("TASKS_PORT", "8082"),
		AuthBaseURL: getEnv("AUTH_BASE_URL", "http://127.0.0.1:8081"),
		RedisAddr:   getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	log.Printf("Config loaded: port=%s auth=%s redis=%s loglevel=%s",
		cfg.Port, cfg.AuthBaseURL, cfg.RedisAddr, cfg.LogLevel)

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

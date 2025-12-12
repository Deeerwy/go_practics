package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr      string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	CacheTTL      time.Duration
}

func FromEnv() Config {
	// конфиг из .env
	httpAddr := getenv("HTTP_ADDR", ":8080")
	redisAddr := getenv("REDIS_ADDR", "127.0.0.1:6379")
	redisPass := getenv("REDIS_PASSWORD", "")
	redisDB := atoi(getenv("REDIS_DB", "0"), 0)
	ttlSec := atoi(getenv("CACHE_TTL_SECONDS", "45"), 45)

	return Config{
		HTTPAddr:      httpAddr,
		RedisAddr:     redisAddr,
		RedisPassword: redisPass,
		RedisDB:       redisDB,
		CacheTTL:      time.Duration(ttlSec) * time.Second,
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func atoi(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
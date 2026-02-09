package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort string
	NatsURL    string
	RPCTimeout time.Duration
	LogLevel   string
}

func LoadConfig() *Config {
	rpcTimeoutSec := getEnvAsInt("RPC_TIMEOUT_SEC", 5)

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		NatsURL:    getEnv("NATS_URL", "nats://localhost:4222"),
		RPCTimeout: time.Duration(rpcTimeoutSec) * time.Second,
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

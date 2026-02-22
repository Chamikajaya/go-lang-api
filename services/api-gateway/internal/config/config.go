package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort           string
	NatsURL              string
	RPCTimeout           time.Duration
	LogLevel             string
	ShutdownTimeout      time.Duration
	RECONNECT_WAIT       time.Duration
	NATS_MAX_RECONNECTS  int
	NATS_CONNECTION_NAME string
}

func LoadConfig() *Config {
	rpcTimeoutSec := getEnvAsInt("RPC_TIMEOUT_SEC", 5)

	return &Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		RPCTimeout:      time.Duration(rpcTimeoutSec) * time.Second,
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ShutdownTimeout: time.Duration(getEnvAsInt("SHUTDOWN_TIMEOUT_SEC", 30)) * time.Second,

		// NATS
		NatsURL:              getEnv("NATS_URL", "nats://localhost:4222"),
		RECONNECT_WAIT:       time.Duration(getEnvAsInt("NATS_RECONNECT_WAIT_SEC", 2)) * time.Second,
		NATS_MAX_RECONNECTS:  getEnvAsInt("NATS_MAX_RECONNECTS", 10),
		NATS_CONNECTION_NAME: getEnv("NATS_CONNECTION_NAME", "api-gateway"),
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

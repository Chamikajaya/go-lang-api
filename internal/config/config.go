package config

import (
	"fmt"
	"os"
)


type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

func LoadConfig() (*Config, error) {

	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "user_management"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}

	return config, nil
}

// belongs to the Config struct - method on Config struct (receiver function)
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// private - unexported
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}



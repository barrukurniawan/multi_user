package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	RedisAddr      string
	RedisPassword  string
	JWTSecret      string
	JWTExpiryHours int
	ServerPort     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	expiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRY_HOURS harus berupa string angka: %w", err)
	}

	cfg := &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "taskuser"),
		DBPassword:     getEnv("DB_PASSWORD", "taskpass"),
		DBName:         getEnv("DB_NAME", "taskdb"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTExpiryHours: expiryHours,
		ServerPort:     getEnv("SERVER_PORT", "8080"),
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET wajib diisi")
	}
	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

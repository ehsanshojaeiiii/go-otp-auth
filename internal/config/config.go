package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OTP      OTPConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	SecretKey string
	ExpiryHours int
}

type OTPConfig struct {
	Length         int
	ExpiryMinutes  int
	MaxAttempts    int
	RateLimitWindow time.Duration
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "otp_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			SecretKey:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		},
		OTP: OTPConfig{
			Length:          getEnvAsInt("OTP_LENGTH", 6),
			ExpiryMinutes:   getEnvAsInt("OTP_EXPIRY_MINUTES", 2),
			MaxAttempts:     getEnvAsInt("OTP_MAX_ATTEMPTS", 3),
			RateLimitWindow: time.Duration(getEnvAsInt("OTP_RATE_LIMIT_MINUTES", 10)) * time.Minute,
		},
	}
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.DBName, c.Database.SSLMode)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

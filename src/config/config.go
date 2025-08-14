package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
	Logging   LoggingConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type RateLimitConfig struct {
	RequestsPerMinute int
	Window            time.Duration
}

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

type LoggingConfig struct {
	Level  string
	Format string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
			Window:            time.Minute,
		},
		CORS: CORSConfig{
			AllowOrigins:     getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
			AllowMethods:     getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowHeaders:     getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "Accept", "X-Requested-With"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsSlice(name string, defaultVal []string) []string {
	valStr := getEnv(name, "")
	if valStr == "" {
		return defaultVal
	}
	return strings.Split(valStr, ",")
}

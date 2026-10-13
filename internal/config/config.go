package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	Timezone string
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type Config struct {
	AppEnv  string
	AppPort string

	DB  DBConfig
	JWT JWTConfig
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s", d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.Timezone)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	ttlHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TTL_HOURS: %v", err)
	}

	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "warehouse"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			Timezone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			TTL:    time.Duration(ttlHours) * time.Hour,
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

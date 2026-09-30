package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type AppConfig struct {
	Port           string
	Env            string
	AdminEmail     string
	AdminPassword  string
	SellerEmail    string
	SellerPassword string
	BuyerEmail     string
	BuyerPassword  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	SecretKey string
}

type CORSConfig struct {
	AllowOrigins []string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func (c *Config) validate() error {
	if c.JWT.SecretKey == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.App.Env != "testing" && c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if len(c.CORS.AllowOrigins) == 0 {
		return fmt.Errorf("CORS_ALLOW_ORIGINS is required (comma-separated list of allowed origins)")
	}
	return nil
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Port:           getEnv("APP_PORT", "8080"),
			Env:            getEnv("APP_ENV", "development"),
			AdminEmail:     getEnv("ADMIN_EMAIL", "admin@example.com"),
			AdminPassword:  getEnv("ADMIN_PASSWORD", "admin123"),
			SellerEmail:    getEnv("SELLER_EMAIL", "seller@example.com"),
			SellerPassword: getEnv("SELLER_PASSWORD", "seller123"),
			BuyerEmail:     getEnv("BUYER_EMAIL", "buyer@example.com"),
			BuyerPassword:  getEnv("BUYER_PASSWORD", "buyer123"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "teepee_marketplace"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET", ""),
		},
		CORS: CORSConfig{
			AllowOrigins: parseCommaSeparated(getEnv("CORS_ALLOW_ORIGINS", "")),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

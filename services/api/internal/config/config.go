package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port         int
	ReplicaID    string
	DBHost       string
	DBPort       int
	DBUser       string
	DBPassword   string
	DBName       string
	NATSURL      string
	JWTSecret    string
	SMTPHost     string
	SMTPPort     int
	TelemetrySec int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:         getEnvInt("PORT", 8080),
		ReplicaID:    getEnv("REPLICA_ID", "api-standalone"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnvInt("DB_PORT", 5432),
		DBUser:       getEnv("DB_USER", "monitopro"),
		DBPassword:   getEnv("DB_PASSWORD", "dev_secret_password"),
		DBName:       getEnv("DB_NAME", "monitopro_db"),
		NATSURL:      getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret:    getEnv("JWT_SECRET", "dev_jwt_secret_key_minimum_32_characters_long"),
		SMTPHost:     getEnv("SMTP_HOST", "localhost"),
		SMTPPort:     getEnvInt("SMTP_PORT", 1025),
		TelemetrySec: getEnvInt("TELEMETRY_INTERVAL_SEC", 30),
	}
	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

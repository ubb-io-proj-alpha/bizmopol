package config

import (
    "log/slog"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    DBType         string
    DBURL          string
    Port           string
    AppEnvironment string
    JWTSecret      string
}

func LoadConfig() *Config {
    if err := godotenv.Load(); err != nil {
        slog.Info("No .env file found, using system environment variables")
    }

    return &Config{
        DBType:         getEnv("DB_TYPE", "sqlite"),
        DBURL:          getEnv("DATABASE_URL", "test.db"),
        Port:           getEnv("PORT", "8080"),
        AppEnvironment: getEnv("APP_ENV", "development"),
        JWTSecret:      getEnv("JWT_SECRET", "changeme-secret-key"),
    }
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}

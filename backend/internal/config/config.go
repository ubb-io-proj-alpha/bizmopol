package config

import (
    "log/slog"
    "os"
    "strings"

    "github.com/joho/godotenv"
)

type Config struct {
    DBType         string
    DBURL          string
    Port           string
    AppEnvironment string
    JWTSecret      string
    DebugLog       bool
}

func LoadConfig() *Config {
    if err := godotenv.Load(); err != nil {
        slog.Info("No .env file found, using system environment variables")
    }

    debugLog := strings.EqualFold(getEnv("DEBUG_LOG", "false"), "true")

    return &Config{
        DBType:         getEnv("DB_TYPE", "sqlite"),
        DBURL:          getEnv("DATABASE_URL", "test.db"),
        Port:           getEnv("PORT", "8080"),
        AppEnvironment: getEnv("APP_ENV", "development"),
        JWTSecret:      getEnv("JWT_SECRET", "changeme-secret-key"),
        DebugLog:       debugLog,
    }
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBType string // "postgres" or "sqlite"
	DBURL  string // DSN or file path
	Port   string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		DBType: getEnv("DB_TYPE", "sqlite"),
		DBURL:  getEnv("DATABASE_URL", "test.db"),
		Port:   getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
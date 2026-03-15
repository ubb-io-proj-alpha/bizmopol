package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(dbType, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch dbType {
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	// Connection Pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if dbType == "postgres" {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
	} else {
		sqlDB.SetMaxOpenConns(1)
	}

	log.Printf("Successfully connected to %s database", dbType)
	return db, nil
}
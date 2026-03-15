package config

import (
    "log/slog"

    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"

    "backend/internal/model"
)

func SeedDatabase(db *gorm.DB) {
	seedUsers(db)
}

func seedUsers(db *gorm.DB) {
    users := []struct {
        email    string
        password string
        name     string
    }{
        {"alice@example.com", "password123", "Alice Smith"},
        {"bob@example.com", "password123", "Bob Jones"},
        {"carol@example.com", "password123", "Carol White"},
    }

    for _, u := range users {
		// skip users that already exist
        var existing model.User
        if err := db.Where("email = ?", u.email).First(&existing).Error; err == nil {
            continue
        }

        hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
        if err != nil {
            slog.Error("SeedDatabase: failed to hash password", "email", u.email, "error", err)
            continue
        }

        user := model.User{
            Email:        u.email,
            PasswordHash: string(hash),
            Name:         u.name,
        }
        if err := db.Create(&user).Error; err != nil {
            slog.Error("SeedDatabase: failed to create user", "email", u.email, "error", err)
        }
    }
}

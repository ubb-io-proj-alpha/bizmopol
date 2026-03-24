package config

import (
    "log/slog"

    "github.com/google/uuid"
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
        role     model.Role
    }{
        {"admin@example.com", "password123", "Alice Smith", model.RoleAdmin},
        {"coach@example.com", "password123", "Bob Jones", model.RoleCoach},
        {"user@example.com", "password123", "Carol White", model.RoleUser},
    }

    for _, u := range users {
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
            ID:           uuid.NewString(),
            Email:        u.email,
            PasswordHash: string(hash),
            Name:         u.name,
            Role:         u.role,
        }
        if err := db.Create(&user).Error; err != nil {
            slog.Error("SeedDatabase: failed to create user", "email", u.email, "error", err)
        }
    }
}

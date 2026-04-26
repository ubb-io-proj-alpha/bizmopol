package config

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

		err = db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "email"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"password_hash": user.PasswordHash,
				"name":          user.Name,
				"role":          user.Role,
				"updated_at":    time.Now(),
			}),
		}).Create(&user).Error
		if err != nil {
			slog.Error("SeedDatabase: failed to upsert user", "email", u.email, "error", err)
		}
	}
}

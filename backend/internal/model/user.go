package model

import "time"

type User struct {
    ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Email        string    `gorm:"size:255;not null;uniqueIndex" json:"email"`
    PasswordHash string    `gorm:"size:255;not null" json:"-"`
    Name         string    `gorm:"size:255" json:"name"`
    CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

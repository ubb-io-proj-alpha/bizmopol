package model

import "time"

type Role string

const (
    RoleAdmin Role = "Admin"
    RoleCoach Role = "Coach"
    RoleUser  Role = "User"
)

type User struct {
    ID           string    `gorm:"primaryKey;size:36" json:"id"`
    Email        string    `gorm:"size:255;not null;uniqueIndex" json:"email"`
    PasswordHash string    `gorm:"size:255;not null" json:"-"`
    Name         string    `gorm:"size:255" json:"name"`
    Role         Role      `gorm:"size:20;not null;default:User" json:"role"`
    CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

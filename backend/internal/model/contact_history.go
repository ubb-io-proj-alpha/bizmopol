package model

import "time"

type ContactHistory struct {
    ID          string    `gorm:"primaryKey;size:36" json:"id"`
    ContactID   string    `gorm:"size:36;not null;index" json:"contact_id"`
    Action      string    `gorm:"size:50;not null" json:"action"`
    Description string    `gorm:"type:text" json:"description"`
    UserID      string    `gorm:"size:36" json:"user_id"`
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

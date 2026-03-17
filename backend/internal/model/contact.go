package model

import "time"

type Contact struct {
    ID          string    `gorm:"primaryKey;size:36" json:"id"`
    Name        string    `gorm:"size:255;not null" json:"name"`
    Email       string    `gorm:"size:255;index" json:"email"`
    Phone       string    `gorm:"size:50" json:"phone"`
    Company     string    `gorm:"size:255" json:"company"`
    Status      string    `gorm:"size:50;default:lead" json:"status"`
    Notes       string    `gorm:"type:text" json:"notes"`
    IsGroup     bool      `gorm:"default:false" json:"is_group"`
    GroupID     string    `gorm:"size:36;index" json:"group_id"`
    LeadScore   int       `gorm:"default:0" json:"lead_score"`
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

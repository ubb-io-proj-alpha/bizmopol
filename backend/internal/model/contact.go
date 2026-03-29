package model

import "time"

type Contact struct {
    ID           string             `gorm:"primaryKey;size:36" json:"id"`
    Name         string             `gorm:"size:255;not null" json:"name"`
    Email        string             `gorm:"size:255;index" json:"email"`
    Phone        string             `gorm:"size:50" json:"phone"`
    Company      string             `gorm:"size:255" json:"company"`
    Status       string             `gorm:"size:50;default:lead" json:"status"`
    Notes        string             `gorm:"type:text" json:"notes"`
    IsGroup      bool               `gorm:"default:false" json:"is_group"`
    GroupID      string             `gorm:"size:36;index" json:"group_id"`
    LeadScore    int                `gorm:"default:0" json:"lead_score"`
    DndActive    bool               `gorm:"default:false" json:"dnd_active"`
    DndType      string             `gorm:"size:20;default:''" json:"dnd_type"`
    DndReason    string             `gorm:"type:text" json:"dnd_reason"`
    DndUntil     *time.Time         `gorm:"default:null" json:"dnd_until"`
    Tags         []Tag              `gorm:"many2many:contact_tags;" json:"tags"`
    CustomValues []CustomFieldValue `gorm:"foreignKey:ContactID" json:"custom_values"`
    CreatedAt    time.Time          `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt    time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
}

type Tag struct {
    ID        string    `gorm:"primaryKey;size:36" json:"id"`
    Name      string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
    Color     string    `gorm:"size:20;default:#6366f1" json:"color"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type CustomField struct {
    ID        string    `gorm:"primaryKey;size:36" json:"id"`
    Name      string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
    FieldType string    `gorm:"size:50;not null;default:text" json:"field_type"`
    Visible   bool      `gorm:"default:true" json:"visible"`
    SortOrder int       `gorm:"default:0" json:"sort_order"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CustomFieldValue struct {
    ID            string    `gorm:"primaryKey;size:36" json:"id"`
    ContactID     string    `gorm:"size:36;not null;index" json:"contact_id"`
    CustomFieldID string    `gorm:"size:36;not null;index" json:"custom_field_id"`
    Value         string    `gorm:"type:text" json:"value"`
    CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

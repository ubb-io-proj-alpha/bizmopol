package model

import "time"

type Funnel struct {
    ID           string    `gorm:"primaryKey;size:36" json:"id"`
    Name         string    `gorm:"size:255;not null" json:"name"`
    Subdomain    string    `gorm:"size:255;index" json:"subdomain"`
    CustomDomain string    `gorm:"size:255;index" json:"custom_domain"`
    Pages        []Page    `gorm:"foreignKey:FunnelID" json:"pages"`
    CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Page struct {
    ID        string    `gorm:"primaryKey;size:36" json:"id"`
    FunnelID  string    `gorm:"size:36;not null;index" json:"funnel_id"`
    Name      string    `gorm:"size:255;not null" json:"name"`
    Path      string    `gorm:"size:255;not null;index" json:"path"`
    Structure string    `gorm:"type:text" json:"structure"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Block struct {
    ID       string                 `json:"id"`
    Type     string                 `json:"type"`
    Content  map[string]interface{} `json:"content"`
    Children []Block                `json:"children"`
}

type FunnelVisit struct {
    ID        string    `gorm:"primaryKey;size:36" json:"id"`
    FunnelID  string    `gorm:"size:36;not null;index" json:"funnel_id"`
    PageID    string    `gorm:"size:36;not null;index" json:"page_id"`
    Ip        string    `gorm:"size:50" json:"ip"`
    UserAgent string    `gorm:"type:text" json:"user_agent"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
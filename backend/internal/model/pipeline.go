package model

import "time"

type Pipeline struct {
    ID          string    `gorm:"primaryKey;size:36" json:"id"`
    Name        string    `gorm:"size:255;not null" json:"name"`
    Description string    `gorm:"type:text" json:"description"`
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
    Stages      []Stage   `gorm:"foreignKey:PipelineID;constraint:OnDelete:CASCADE" json:"stages"`
}

type Stage struct {
    ID         string    `gorm:"primaryKey;size:36" json:"id"`
    PipelineID string    `gorm:"size:36;not null;index" json:"pipeline_id"`
    Name       string    `gorm:"size:255;not null" json:"name"`
    Color      string    `gorm:"size:20;default:#38bdf8" json:"color"`
    SortOrder  int       `gorm:"default:0" json:"sort_order"`
    CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type ContactStage struct {
    ID         string    `gorm:"primaryKey;size:36" json:"id"`
    ContactID  string    `gorm:"size:36;not null;index" json:"contact_id"`
    PipelineID string    `gorm:"size:36;not null;index" json:"pipeline_id"`
    StageID    string    `gorm:"size:36;not null;index" json:"stage_id"`
    CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
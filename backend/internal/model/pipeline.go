package model

import "time"

type Pipeline struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Stages    []Stage   `gorm:"foreignKey:PipelineID;constraint:OnDelete:CASCADE" json:"stages"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Stage struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	PipelineID uint      `gorm:"not null;index" json:"pipeline_id"`
	Position   int       `gorm:"default:0" json:"position"`
	Leads      []Lead    `gorm:"foreignKey:StageID;constraint:OnDelete:SET NULL" json:"leads"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Lead struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Email     string    `gorm:"size:255" json:"email"`
	Phone     string    `gorm:"size:20" json:"phone"`
	StageID   *uint     `gorm:"index" json:"stage_id"`
	Position  int       `gorm:"default:0" json:"position"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

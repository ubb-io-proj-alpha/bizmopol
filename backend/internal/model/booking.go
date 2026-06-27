package model

import "time"


type BookingSettings struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	OwnerUserID  string    `gorm:"size:36;index" json:"owner_user_id"`
	Enabled      bool      `gorm:"default:false" json:"enabled"`
	WorkingDays  string    `gorm:"size:50;default:'1,2,3,4,5'" json:"working_days"` // CSV of time.Weekday (Sun=0..Sat=6)
	StartMinutes int       `gorm:"default:540" json:"start_minutes"`                // minutes from midnight, 540 = 09:00
	EndMinutes   int       `gorm:"default:1020" json:"end_minutes"`                 // 1020 = 17:00
	SlotMinutes  int       `gorm:"default:30" json:"slot_minutes"`                  // 
	MeetingTitle string    `gorm:"size:255;default:''" json:"meeting_title"`        // template, {name} is replaced
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

const BookingSettingsID = "default"

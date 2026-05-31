package model

import "time"

type CalendarEvent struct {
	ID            string     `gorm:"primaryKey;size:36" json:"id"`
	UserID        string     `gorm:"size:36;not null;index" json:"user_id"`
	Title         string     `gorm:"size:500;not null" json:"title"`
	Description   string     `gorm:"type:text" json:"description"`
	EventType     string     `gorm:"size:30;default:meeting" json:"event_type"`
	StartTime     time.Time  `gorm:"not null;index" json:"start_time"`
	EndTime       time.Time  `gorm:"not null" json:"end_time"`
	AllDay        bool       `gorm:"default:false" json:"all_day"`
	Location      string     `gorm:"size:500" json:"location"`
	Status        string     `gorm:"size:20;default:scheduled" json:"status"`
	ContactID     string     `gorm:"size:36;index" json:"contact_id"`
	ContactName   string     `gorm:"size:255" json:"contact_name"`
	ContactEmail  string     `gorm:"size:255" json:"contact_email"`
	ZoomMeetingID int64      `gorm:"default:0" json:"zoom_meeting_id"`
	ZoomJoinURL   string     `gorm:"size:1000" json:"zoom_join_url"`
	ZoomStartURL  string     `gorm:"size:1000" json:"zoom_start_url"`
	ZoomPasscode  string     `gorm:"size:50" json:"zoom_passcode"`
	Color         string     `gorm:"size:20;default:#38bdf8" json:"color"`
	Reminder      int        `gorm:"default:15" json:"reminder"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type ZoomAccount struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	UserID       string    `gorm:"size:36;not null;uniqueIndex" json:"user_id"`
	AccountID    string    `gorm:"size:255;not null" json:"account_id"`
	ClientID     string    `gorm:"size:255;not null" json:"client_id"`
	ClientSecret string    `gorm:"size:500" json:"-"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

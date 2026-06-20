package dto

import "time"

type EventCreateRequest struct {
	Title        string    `json:"title" binding:"required"`
	Description  string    `json:"description"`
	EventType    string    `json:"event_type"`
	StartTime    time.Time `json:"start_time" binding:"required"`
	EndTime      time.Time `json:"end_time" binding:"required"`
	AllDay       bool      `json:"all_day"`
	Location     string    `json:"location"`
	ContactID    string    `json:"contact_id"`
	ContactName  string    `json:"contact_name"`
	ContactEmail string    `json:"contact_email"`
	CreateZoom   bool      `json:"create_zoom"`
	Color        string    `json:"color"`
	Reminder     int       `json:"reminder"`
}

type EventUpdateRequest struct {
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	EventType    string     `json:"event_type"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	AllDay       *bool      `json:"all_day"`
	Location     string     `json:"location"`
	Status       string     `json:"status"`
	ContactID    string     `json:"contact_id"`
	ContactName  string     `json:"contact_name"`
	ContactEmail string     `json:"contact_email"`
	Color        string     `json:"color"`
	Reminder     *int       `json:"reminder"`
	RemoveZoom   *bool      `json:"remove_zoom"`
	AddZoom      *bool      `json:"add_zoom"`
}

type EventResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	EventType     string    `json:"event_type"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	AllDay        bool      `json:"all_day"`
	Location      string    `json:"location"`
	Status        string    `json:"status"`
	ContactID     string    `json:"contact_id"`
	ContactName   string    `json:"contact_name"`
	ContactEmail  string    `json:"contact_email"`
	ZoomMeetingID int64     `json:"zoom_meeting_id"`
	ZoomJoinURL   string    `json:"zoom_join_url"`
	ZoomStartURL  string    `json:"zoom_start_url"`
	ZoomPasscode  string    `json:"zoom_passcode"`
	Color         string    `json:"color"`
	Reminder      int       `json:"reminder"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type EventListResponse struct {
	Data       []EventResponse `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

type EventQuery struct {
	From      string `form:"from"`
	To        string `form:"to"`
	Status    string `form:"status"`
	EventType string `form:"event_type"`
	ContactID string `form:"contact_id"`
	Search    string `form:"search"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

type ZoomAccountRequest struct {
	AccountID    string `json:"account_id" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
}

type ZoomAccountResponse struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	ClientID  string    `json:"client_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type ZoomTestResponse struct {
	Status string `json:"status"`
	Email  string `json:"email,omitempty"`
	Error  string `json:"error,omitempty"`
}

type CalendarStatsResponse struct {
	TotalEvents    int64 `json:"total_events"`
	UpcomingEvents int64 `json:"upcoming_events"`
	TodayEvents    int64 `json:"today_events"`
	ZoomMeetings   int64 `json:"zoom_meetings"`
}

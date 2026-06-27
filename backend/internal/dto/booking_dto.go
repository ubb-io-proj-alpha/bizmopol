package dto

type BookingSettingsRequest struct {
	Enabled      bool   `json:"enabled"`
	WorkingDays  []int  `json:"working_days"` // time.Weekday ints (Sun=0..Sat=6)
	StartMinutes int    `json:"start_minutes"`
	EndMinutes   int    `json:"end_minutes"`
	SlotMinutes  int    `json:"slot_minutes"`
	MeetingTitle string `json:"meeting_title"`
}

type BookingSettingsResponse struct {
	Enabled      bool   `json:"enabled"`
	Configured   bool   `json:"configured"`
	WorkingDays  []int  `json:"working_days"`
	StartMinutes int    `json:"start_minutes"`
	EndMinutes   int    `json:"end_minutes"`
	SlotMinutes  int    `json:"slot_minutes"`
	MeetingTitle string `json:"meeting_title"`
}

// BookingDay is one day in the public month view: available=false means the day
// cannot be picked (weekend, outside working days, in the past, or fully booked).
type BookingDay struct {
	Date      string `json:"date"` // YYYY-MM-DD
	Available bool   `json:"available"`
}

// BookingSlot is one time slot of a day. Available=false slots are shown disabled.
type BookingSlot struct {
	Start     string `json:"start"` // HH:MM
	End       string `json:"end"`   // HH:MM
	Available bool   `json:"available"`
}

type BookingCreateRequest struct {
	Date  string `json:"date" binding:"required"`  // YYYY-MM-DD
	Start string `json:"start" binding:"required"` // HH:MM
	Name  string `json:"name" binding:"required"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type BookingCreateResponse struct {
	Success   bool   `json:"success"`
	EventID   string `json:"event_id"`
	StartTime string `json:"start_time"`
}

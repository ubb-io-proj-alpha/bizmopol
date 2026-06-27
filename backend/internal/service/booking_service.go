package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/dto"
	"backend/internal/model"
	"backend/internal/repository"
)

var (
	ErrBookingDisabled = errors.New("rezerwacje niedostępne")
	ErrSlotUnavailable = errors.New("wybrany termin jest już zajęty")
	ErrSlotOutOfWindow = errors.New("wybrany termin jest poza dostępnymi godzinami")
)

const bookingColor = "#a78bfa"

type BookingService interface {
	GetSettings(ctx context.Context) (*dto.BookingSettingsResponse, error)
	UpdateSettings(ctx context.Context, ownerUserID string, req dto.BookingSettingsRequest) (*dto.BookingSettingsResponse, error)

	MonthDays(ctx context.Context, fromDate, toDate time.Time) ([]dto.BookingDay, error)
	DaySlots(ctx context.Context, date time.Time) ([]dto.BookingSlot, error)
	CreateBooking(ctx context.Context, req dto.BookingCreateRequest) (*dto.BookingCreateResponse, error)
}

type bookingService struct {
	repo    repository.BookingRepository
	calRepo repository.CalendarRepository
}

func NewBookingService(repo repository.BookingRepository, calRepo repository.CalendarRepository) BookingService {
	return &bookingService{repo: repo, calRepo: calRepo}
}

func defaultSettings() *model.BookingSettings {
	return &model.BookingSettings{
		Enabled:      false,
		WorkingDays:  "1,2,3,4,5",
		StartMinutes: 540,
		EndMinutes:   1020,
		SlotMinutes:  30,
	}
}

func toSettingsResponse(s *model.BookingSettings, configured bool) *dto.BookingSettingsResponse {
	return &dto.BookingSettingsResponse{
		Enabled:      s.Enabled,
		Configured:   configured,
		WorkingDays:  csvToDays(s.WorkingDays),
		StartMinutes: s.StartMinutes,
		EndMinutes:   s.EndMinutes,
		SlotMinutes:  s.SlotMinutes,
		MeetingTitle: s.MeetingTitle,
	}
}

func (s *bookingService) GetSettings(ctx context.Context) (*dto.BookingSettingsResponse, error) {
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return nil, ErrInternal
	}
	if settings == nil {
		return toSettingsResponse(defaultSettings(), false), nil
	}
	return toSettingsResponse(settings, true), nil
}

func (s *bookingService) UpdateSettings(ctx context.Context, ownerUserID string, req dto.BookingSettingsRequest) (*dto.BookingSettingsResponse, error) {
	debugLog("Booking.UpdateSettings", "owner", ownerUserID, "enabled", req.Enabled, "slot", req.SlotMinutes)

	if req.SlotMinutes <= 0 {
		req.SlotMinutes = 30
	}
	if req.StartMinutes < 0 || req.EndMinutes > 24*60 || req.StartMinutes >= req.EndMinutes {
		return nil, ErrSlotOutOfWindow
	}

	settings := &model.BookingSettings{
		OwnerUserID:  ownerUserID,
		Enabled:      req.Enabled,
		WorkingDays:  daysToCSV(req.WorkingDays),
		StartMinutes: req.StartMinutes,
		EndMinutes:   req.EndMinutes,
		SlotMinutes:  req.SlotMinutes,
		MeetingTitle: req.MeetingTitle,
	}
	if err := s.repo.SaveSettings(ctx, settings); err != nil {
		debugLogResult("Booking.UpdateSettings", err)
		return nil, ErrInternal
	}
	return toSettingsResponse(settings, true), nil
}

func (s *bookingService) activeSettings(ctx context.Context) (*model.BookingSettings, error) {
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return nil, ErrInternal
	}
	if settings == nil || !settings.Enabled || settings.OwnerUserID == "" {
		return nil, nil
	}
	return settings, nil
}

func (s *bookingService) MonthDays(ctx context.Context, fromDate, toDate time.Time) ([]dto.BookingDay, error) {
	settings, err := s.activeSettings(ctx)
	if err != nil {
		return nil, err
	}
	days := make([]dto.BookingDay, 0)
	if settings == nil {
		return days, nil
	}

	rangeEnd := toDate.AddDate(0, 0, 1)
	events, err := s.calRepo.EventsInRange(ctx, settings.OwnerUserID, fromDate, rangeEnd)
	if err != nil {
		return nil, ErrInternal
	}

	now := time.Now()
	workdays := dayset(settings.WorkingDays)
	for d := fromDate; !d.After(toDate); d = d.AddDate(0, 0, 1) {
		available := false
		if workdays[int(d.Weekday())] {
			for _, slot := range buildSlots(settings, d) {
				if slot.start.After(now) && !collides(slot.start, slot.end, events) {
					available = true
					break
				}
			}
		}
		days = append(days, dto.BookingDay{Date: d.Format("2006-01-02"), Available: available})
	}
	return days, nil
}

func (s *bookingService) DaySlots(ctx context.Context, date time.Time) ([]dto.BookingSlot, error) {
	settings, err := s.activeSettings(ctx)
	if err != nil {
		return nil, err
	}
	slots := make([]dto.BookingSlot, 0)
	if settings == nil {
		return slots, nil
	}
	if !dayset(settings.WorkingDays)[int(date.Weekday())] {
		return slots, nil
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	dayEnd := dayStart.AddDate(0, 0, 1)
	events, err := s.calRepo.EventsInRange(ctx, settings.OwnerUserID, dayStart, dayEnd)
	if err != nil {
		return nil, ErrInternal
	}

	now := time.Now()
	for _, slot := range buildSlots(settings, dayStart) {
		available := slot.start.After(now) && !collides(slot.start, slot.end, events)
		slots = append(slots, dto.BookingSlot{
			Start:     slot.start.Format("15:04"),
			End:       slot.end.Format("15:04"),
			Available: available,
		})
	}
	return slots, nil
}

func (s *bookingService) CreateBooking(ctx context.Context, req dto.BookingCreateRequest) (*dto.BookingCreateResponse, error) {
	debugLog("Booking.CreateBooking", "date", req.Date, "start", req.Start, "name", req.Name)

	settings, err := s.activeSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, ErrBookingDisabled
	}

	day, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
	if err != nil {
		return nil, ErrSlotOutOfWindow
	}
	startMin, err := parseHHMM(req.Start)
	if err != nil {
		return nil, ErrSlotOutOfWindow
	}

	slotStart := time.Date(day.Year(), day.Month(), day.Day(), 0, startMin, 0, 0, time.Local)
	slotEnd := slotStart.Add(time.Duration(settings.SlotMinutes) * time.Minute)

	endMin := startMin + settings.SlotMinutes
	if !dayset(settings.WorkingDays)[int(slotStart.Weekday())] ||
		startMin < settings.StartMinutes || endMin > settings.EndMinutes ||
		(startMin-settings.StartMinutes)%settings.SlotMinutes != 0 {
		return nil, ErrSlotOutOfWindow
	}
	if !slotStart.After(time.Now()) {
		return nil, ErrSlotOutOfWindow
	}

	events, err := s.calRepo.EventsInRange(ctx, settings.OwnerUserID, slotStart, slotEnd)
	if err != nil {
		return nil, ErrInternal
	}
	if collides(slotStart, slotEnd, events) {
		return nil, ErrSlotUnavailable
	}

	title := strings.TrimSpace(settings.MeetingTitle)
	if title == "" {
		title = "Spotkanie: {name}"
	}
	title = strings.ReplaceAll(title, "{name}", req.Name)

	event := &model.CalendarEvent{
		ID:           uuid.NewString(),
		UserID:       settings.OwnerUserID,
		Title:        title,
		Description:  "Rezerwacja online z lejka sprzedażowego",
		EventType:    "meeting",
		StartTime:    slotStart,
		EndTime:      slotEnd,
		Status:       "scheduled",
		ContactName:  req.Name,
		ContactEmail: req.Email,
		Color:        bookingColor,
		Reminder:     15,
	}
	if err := s.calRepo.CreateEvent(ctx, event); err != nil {
		debugLogResult("Booking.CreateBooking", err)
		return nil, ErrInternal
	}

	debugLogResult("Booking.CreateBooking", nil, "event_id", event.ID, "start", slotStart)
	return &dto.BookingCreateResponse{
		Success:   true,
		EventID:   event.ID,
		StartTime: slotStart.Format(time.RFC3339),
	}, nil
}


type slotRange struct {
	start time.Time
	end   time.Time
}

func buildSlots(s *model.BookingSettings, day time.Time) []slotRange {
	var slots []slotRange
	if s.SlotMinutes <= 0 {
		return slots
	}
	midnight := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
	for m := s.StartMinutes; m+s.SlotMinutes <= s.EndMinutes; m += s.SlotMinutes {
		start := midnight.Add(time.Duration(m) * time.Minute)
		slots = append(slots, slotRange{start: start, end: start.Add(time.Duration(s.SlotMinutes) * time.Minute)})
	}
	return slots
}

func collides(start, end time.Time, events []*model.CalendarEvent) bool {
	for _, e := range events {
		if start.Before(e.EndTime) && end.After(e.StartTime) {
			return true
		}
	}
	return false
}

func dayset(csv string) map[int]bool {
	set := make(map[int]bool)
	for _, d := range csvToDays(csv) {
		set[d] = true
	}
	return set
}

func csvToDays(csv string) []int {
	var days []int
	for _, part := range strings.Split(csv, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if n, err := strconv.Atoi(part); err == nil && n >= 0 && n <= 6 {
			days = append(days, n)
		}
	}
	sort.Ints(days)
	return days
}

func daysToCSV(days []int) string {
	seen := make(map[int]bool)
	var valid []int
	for _, d := range days {
		if d >= 0 && d <= 6 && !seen[d] {
			seen[d] = true
			valid = append(valid, d)
		}
	}
	sort.Ints(valid)
	parts := make([]string, len(valid))
	for i, d := range valid {
		parts[i] = strconv.Itoa(d)
	}
	return strings.Join(parts, ",")
}

func parseHHMM(s string) (int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, fmt.Errorf("invalid hour")
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid minute")
	}
	return h*60 + m, nil
}

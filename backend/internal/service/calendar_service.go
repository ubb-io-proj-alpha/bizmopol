package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"

	"backend/internal/dto"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/worker"
	"backend/internal/zoom"
)

type CalendarService interface {
	CreateEvent(ctx context.Context, userID string, req dto.EventCreateRequest) (*dto.EventResponse, error)
	GetEvent(ctx context.Context, id string) (*dto.EventResponse, error)
	UpdateEvent(ctx context.Context, userID, id string, req dto.EventUpdateRequest) (*dto.EventResponse, error)
	DeleteEvent(ctx context.Context, userID, id string) error
	ListEvents(ctx context.Context, userID string, q dto.EventQuery) (*dto.EventListResponse, error)
	UpcomingEvents(ctx context.Context, userID string) ([]dto.EventResponse, error)
	GetStats(ctx context.Context, userID string) (*dto.CalendarStatsResponse, error)

	CreateZoomAccount(ctx context.Context, userID string, req dto.ZoomAccountRequest) (*dto.ZoomAccountResponse, error)
	GetZoomAccount(ctx context.Context, userID string) (*dto.ZoomAccountResponse, error)
	UpdateZoomAccount(ctx context.Context, userID string, req dto.ZoomAccountRequest) (*dto.ZoomAccountResponse, error)
	DeleteZoomAccount(ctx context.Context, userID string) error
	TestZoomConnection(ctx context.Context, userID string) (*dto.ZoomTestResponse, error)
}

type calendarService struct {
	repo   repository.CalendarRepository
	worker *worker.ZoomWorker
}

func NewCalendarService(repo repository.CalendarRepository, w *worker.ZoomWorker) CalendarService {
	return &calendarService{repo: repo, worker: w}
}

func toEventResponse(e *model.CalendarEvent) dto.EventResponse {
	return dto.EventResponse{
		ID:            e.ID,
		UserID:        e.UserID,
		Title:         e.Title,
		Description:   e.Description,
		EventType:     e.EventType,
		StartTime:     e.StartTime,
		EndTime:       e.EndTime,
		AllDay:        e.AllDay,
		Location:      e.Location,
		Status:        e.Status,
		ContactID:     e.ContactID,
		ContactName:   e.ContactName,
		ContactEmail:  e.ContactEmail,
		ZoomMeetingID: e.ZoomMeetingID,
		ZoomJoinURL:   e.ZoomJoinURL,
		ZoomStartURL:  e.ZoomStartURL,
		ZoomPasscode:  e.ZoomPasscode,
		Color:         e.Color,
		Reminder:      e.Reminder,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func (s *calendarService) getZoomToken(ctx context.Context, userID string) (string, error) {
	account, err := s.repo.FindZoomAccountByUserID(ctx, userID)
	if err != nil || account == nil {
		return "", fmt.Errorf("zoom not configured")
	}
	return zoom.GetAccessToken(account.AccountID, account.ClientID, account.ClientSecret)
}

func (s *calendarService) CreateEvent(ctx context.Context, userID string, req dto.EventCreateRequest) (*dto.EventResponse, error) {
	debugLog("Calendar.CreateEvent", "user_id", userID, "title", req.Title, "type", req.EventType, "create_zoom", req.CreateZoom, "start", req.StartTime)
	eventType := req.EventType
	if eventType == "" {
		eventType = "meeting"
	}
	color := req.Color
	if color == "" {
		switch eventType {
		case "meeting":
			color = "#38bdf8"
		case "call":
			color = "#34d399"
		case "follow_up":
			color = "#f59e0b"
		default:
			color = "#818cf8"
		}
	}
	reminder := req.Reminder
	if reminder == 0 {
		reminder = 15
	}

	event := &model.CalendarEvent{
		ID:           uuid.NewString(),
		UserID:       userID,
		Title:        req.Title,
		Description:  req.Description,
		EventType:    eventType,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		AllDay:       req.AllDay,
		Location:     req.Location,
		Status:       "scheduled",
		ContactID:    req.ContactID,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		Color:        color,
		Reminder:     reminder,
	}

	if req.CreateZoom {
		token, err := s.getZoomToken(ctx, userID)
		if err != nil {
			slog.Warn("Zoom token failed, creating event without Zoom", "error", err)
		} else {
			duration := int(req.EndTime.Sub(req.StartTime).Minutes())
			if duration <= 0 {
				duration = 30
			}
			agenda := req.Description
			if req.ContactName != "" {
				agenda = fmt.Sprintf("Meeting with %s. %s", req.ContactName, req.Description)
			}

			meeting, err := zoom.CreateMeeting(token, req.Title, agenda, req.StartTime, duration)
			if err != nil {
				slog.Error("Zoom meeting creation failed", "error", err)
			} else {
				event.ZoomMeetingID = meeting.ID
				event.ZoomJoinURL = meeting.JoinURL
				event.ZoomStartURL = meeting.StartURL
				event.ZoomPasscode = meeting.Password
				event.Location = meeting.JoinURL
			}
		}
	}

	if err := s.repo.CreateEvent(ctx, event); err != nil {
		debugLogResult("Calendar.CreateEvent", err)
		return nil, ErrInternal
	}

	debugLogResult("Calendar.CreateEvent", nil, "id", event.ID, "zoom_meeting_id", event.ZoomMeetingID)
	r := toEventResponse(event)
	return &r, nil
}

func (s *calendarService) GetEvent(ctx context.Context, id string) (*dto.EventResponse, error) {
	debugLog("Calendar.GetEvent", "id", id)
	e, err := s.repo.FindEventByID(ctx, id)
	if err != nil {
		return nil, ErrInternal
	}
	if e == nil {
		return nil, nil
	}
	r := toEventResponse(e)
	return &r, nil
}

func (s *calendarService) UpdateEvent(ctx context.Context, userID, id string, req dto.EventUpdateRequest) (*dto.EventResponse, error) {
	debugLog("Calendar.UpdateEvent", "user_id", userID, "id", id, "title", req.Title, "status", req.Status,
		"remove_zoom", req.RemoveZoom != nil && *req.RemoveZoom, "add_zoom", req.AddZoom != nil && *req.AddZoom)

	e, err := s.repo.FindEventByID(ctx, id)
	if err != nil {
		debugLogResult("Calendar.UpdateEvent", err, "step", "find_event")
		return nil, ErrInternal
	}
	if e == nil {
		debugLog("Calendar.UpdateEvent", "step", "not_found", "id", id)
		return nil, fmt.Errorf("event not found")
	}

	oldStatus := e.Status
	timeChanged := false

	if req.Title != "" {
		e.Title = req.Title
	}
	if req.Description != "" {
		e.Description = req.Description
	}
	if req.EventType != "" {
		e.EventType = req.EventType
	}
	if req.StartTime != nil {
		e.StartTime = *req.StartTime
		timeChanged = true
	}
	if req.EndTime != nil {
		e.EndTime = *req.EndTime
		timeChanged = true
	}
	if req.AllDay != nil {
		e.AllDay = *req.AllDay
	}
	if req.Location != "" {
		e.Location = req.Location
	}
	if req.Status != "" {
		e.Status = req.Status
	}
	if req.ContactID != "" {
		e.ContactID = req.ContactID
	}
	if req.ContactName != "" {
		e.ContactName = req.ContactName
	}
	if req.ContactEmail != "" {
		e.ContactEmail = req.ContactEmail
	}
	if req.Color != "" {
		e.Color = req.Color
	}
	if req.Reminder != nil {
		e.Reminder = *req.Reminder
	}

	if req.RemoveZoom != nil && *req.RemoveZoom && e.ZoomMeetingID > 0 {
		debugLog("Calendar.UpdateEvent", "step", "remove_zoom", "meeting_id", e.ZoomMeetingID)
		s.worker.Enqueue(worker.ZoomJob{
			Type:      worker.ZoomJobDelete,
			UserID:    userID,
			EventID:   id,
			MeetingID: e.ZoomMeetingID,
		})
		e.ZoomMeetingID = 0
		e.ZoomJoinURL = ""
		e.ZoomStartURL = ""
		e.ZoomPasscode = ""
		e.Location = ""
	} else if req.AddZoom != nil && *req.AddZoom && e.ZoomMeetingID == 0 {
		debugLog("Calendar.UpdateEvent", "step", "add_zoom", "event_id", id)
		desc := e.Description
		if e.ContactName != "" {
			desc = fmt.Sprintf("Meeting with %s. %s", e.ContactName, e.Description)
		}
		s.worker.Enqueue(worker.ZoomJob{
			Type:      worker.ZoomJobCreate,
			UserID:    userID,
			EventID:   id,
			Title:     e.Title,
			Desc:      desc,
			StartTime: e.StartTime,
			EndTime:   e.EndTime,
		})
	} else if e.ZoomMeetingID > 0 {
		if e.Status == "cancelled" && oldStatus != "cancelled" {
			debugLog("Calendar.UpdateEvent", "step", "cancel_zoom_with_event", "meeting_id", e.ZoomMeetingID)
			s.worker.Enqueue(worker.ZoomJob{
				Type:      worker.ZoomJobDelete,
				UserID:    userID,
				EventID:   id,
				MeetingID: e.ZoomMeetingID,
			})
			e.ZoomMeetingID = 0
			e.ZoomJoinURL = ""
			e.ZoomStartURL = ""
			e.ZoomPasscode = ""
		} else if timeChanged || req.Title != "" {
			debugLog("Calendar.UpdateEvent", "step", "sync_zoom", "meeting_id", e.ZoomMeetingID, "time_changed", timeChanged)
			s.worker.Enqueue(worker.ZoomJob{
				Type:      worker.ZoomJobSync,
				UserID:    userID,
				EventID:   id,
				MeetingID: e.ZoomMeetingID,
				Title:     e.Title,
				Desc:      e.Description,
				StartTime: e.StartTime,
				EndTime:   e.EndTime,
			})
		}
	}

	if err := s.repo.UpdateEvent(ctx, e); err != nil {
		debugLogResult("Calendar.UpdateEvent", err)
		return nil, ErrInternal
	}

	debugLogResult("Calendar.UpdateEvent", nil, "id", id, "zoom_meeting_id", e.ZoomMeetingID)
	r := toEventResponse(e)
	return &r, nil
}

func (s *calendarService) DeleteEvent(ctx context.Context, userID, id string) error {
	debugLog("Calendar.DeleteEvent", "user_id", userID, "id", id)
	e, err := s.repo.FindEventByID(ctx, id)
	if err != nil {
		debugLogResult("Calendar.DeleteEvent", err)
		return ErrInternal
	}
	if e == nil {
		return fmt.Errorf("event not found")
	}

	if e.ZoomMeetingID > 0 {
		debugLog("Calendar.DeleteEvent", "step", "delete_zoom", "meeting_id", e.ZoomMeetingID)
		s.worker.Enqueue(worker.ZoomJob{
			Type:      worker.ZoomJobDelete,
			UserID:    userID,
			EventID:   id,
			MeetingID: e.ZoomMeetingID,
		})
	}

	return s.repo.DeleteEvent(ctx, id)
}

func (s *calendarService) ListEvents(ctx context.Context, userID string, q dto.EventQuery) (*dto.EventListResponse, error) {
	debugLog("Calendar.ListEvents", "user_id", userID, "from", q.From, "to", q.To, "status", q.Status, "type", q.EventType, "page", q.Page)
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}

	var from, to *time.Time
	if q.From != "" {
		if t, err := time.Parse("2006-01-02", q.From); err == nil {
			from = &t
		}
	}
	if q.To != "" {
		if t, err := time.Parse("2006-01-02", q.To); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			to = &endOfDay
		}
	}

	events, total, err := s.repo.ListEvents(ctx, userID, q.Search, q.Status, q.EventType, q.ContactID, from, to, page, pageSize)
	if err != nil {
		debugLogResult("Calendar.ListEvents", err)
		return nil, ErrInternal
	}
	debugLogResult("Calendar.ListEvents", nil, "total", total, "returned", len(events))

	items := make([]dto.EventResponse, 0, len(events))
	for _, e := range events {
		items = append(items, toEventResponse(e))
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	return &dto.EventListResponse{
		Data:       items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *calendarService) UpcomingEvents(ctx context.Context, userID string) ([]dto.EventResponse, error) {
	debugLog("Calendar.UpcomingEvents", "user_id", userID)
	events, err := s.repo.UpcomingEvents(ctx, userID, 10)
	if err != nil {
		return nil, ErrInternal
	}
	result := make([]dto.EventResponse, 0, len(events))
	for _, e := range events {
		result = append(result, toEventResponse(e))
	}
	return result, nil
}

func (s *calendarService) GetStats(ctx context.Context, userID string) (*dto.CalendarStatsResponse, error) {
	debugLog("Calendar.GetStats", "user_id", userID)
	stats, err := s.repo.Stats(ctx, userID)
	if err != nil {
		return nil, ErrInternal
	}
	return &dto.CalendarStatsResponse{
		TotalEvents:    stats["total"],
		UpcomingEvents: stats["upcoming"],
		TodayEvents:    stats["today"],
		ZoomMeetings:   stats["zoom"],
	}, nil
}

func toZoomAccountResponse(a *model.ZoomAccount) *dto.ZoomAccountResponse {
	return &dto.ZoomAccountResponse{
		ID:        a.ID,
		AccountID: a.AccountID,
		ClientID:  a.ClientID,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt,
	}
}

func (s *calendarService) CreateZoomAccount(ctx context.Context, userID string, req dto.ZoomAccountRequest) (*dto.ZoomAccountResponse, error) {
	debugLog("Calendar.CreateZoomAccount", "user_id", userID, "account_id", req.AccountID)
	existing, _ := s.repo.FindZoomAccountByUserID(ctx, userID)
	if existing != nil {
		return nil, fmt.Errorf("zoom account already exists, update it instead")
	}

	account := &model.ZoomAccount{
		ID:           uuid.NewString(),
		UserID:       userID,
		AccountID:    req.AccountID,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		IsActive:     true,
	}

	if err := s.repo.CreateZoomAccount(ctx, account); err != nil {
		return nil, ErrInternal
	}
	return toZoomAccountResponse(account), nil
}

func (s *calendarService) GetZoomAccount(ctx context.Context, userID string) (*dto.ZoomAccountResponse, error) {
	debugLog("Calendar.GetZoomAccount", "user_id", userID)
	a, err := s.repo.FindZoomAccountByUserID(ctx, userID)
	if err != nil {
		return nil, ErrInternal
	}
	if a == nil {
		return nil, nil
	}
	return toZoomAccountResponse(a), nil
}

func (s *calendarService) UpdateZoomAccount(ctx context.Context, userID string, req dto.ZoomAccountRequest) (*dto.ZoomAccountResponse, error) {
	debugLog("Calendar.UpdateZoomAccount", "user_id", userID, "account_id", req.AccountID)
	a, err := s.repo.FindZoomAccountByUserID(ctx, userID)
	if err != nil || a == nil {
		return nil, ErrInternal
	}

	a.AccountID = req.AccountID
	a.ClientID = req.ClientID
	if req.ClientSecret != "" {
		a.ClientSecret = req.ClientSecret
	}

	if err := s.repo.UpdateZoomAccount(ctx, a); err != nil {
		return nil, ErrInternal
	}
	return toZoomAccountResponse(a), nil
}

func (s *calendarService) DeleteZoomAccount(ctx context.Context, userID string) error {
	debugLog("Calendar.DeleteZoomAccount", "user_id", userID)
	a, err := s.repo.FindZoomAccountByUserID(ctx, userID)
	if err != nil || a == nil {
		return ErrInternal
	}
	return s.repo.DeleteZoomAccount(ctx, a.ID)
}

func (s *calendarService) TestZoomConnection(ctx context.Context, userID string) (*dto.ZoomTestResponse, error) {
	debugLog("Calendar.TestZoomConnection", "user_id", userID)

	a, err := s.repo.FindZoomAccountByUserID(ctx, userID)
	if err != nil {
		debugLogResult("Calendar.TestZoomConnection", err, "step", "find_zoom_account")
		return &dto.ZoomTestResponse{Status: "error", Error: "No Zoom account configured."}, nil
	}
	if a == nil {
		debugLog("Calendar.TestZoomConnection", "step", "find_zoom_account", "result", "no account found")
		return &dto.ZoomTestResponse{Status: "error", Error: "No Zoom account configured."}, nil
	}
	debugLog("Calendar.TestZoomConnection", "step", "find_zoom_account", "account_id", a.AccountID, "has_client_id", a.ClientID != "", "has_secret", a.ClientSecret != "")

	token, err := zoom.GetAccessToken(a.AccountID, a.ClientID, a.ClientSecret)
	if err != nil {
		debugLogResult("Calendar.TestZoomConnection", err, "step", "get_access_token", "account_id", a.AccountID)
		return &dto.ZoomTestResponse{Status: "error", Error: fmt.Sprintf("Auth failed: %v", err)}, nil
	}
	debugLog("Calendar.TestZoomConnection", "step", "get_access_token", "result", "ok", "token_len", len(token))

	user, err := zoom.GetUser(token)
	if err != nil {
		debugLogResult("Calendar.TestZoomConnection", err, "step", "get_zoom_user")
		return &dto.ZoomTestResponse{Status: "error", Error: fmt.Sprintf("API test failed: %v", err)}, nil
	}
	debugLog("Calendar.TestZoomConnection", "step", "get_zoom_user", "email", user.Email, "zoom_user_id", user.ID)

	debugLogResult("Calendar.TestZoomConnection", nil, "status", "ok", "email", user.Email)
	return &dto.ZoomTestResponse{Status: "ok", Email: user.Email}, nil
}

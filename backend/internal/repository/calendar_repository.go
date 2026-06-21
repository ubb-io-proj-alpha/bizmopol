package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"backend/internal/model"
)

type CalendarRepository interface {
	CreateEvent(ctx context.Context, e *model.CalendarEvent) error
	FindEventByID(ctx context.Context, id string) (*model.CalendarEvent, error)
	UpdateEvent(ctx context.Context, e *model.CalendarEvent) error
	DeleteEvent(ctx context.Context, id string) error
	ListEvents(ctx context.Context, userID, search, status, eventType, contactID string, from, to *time.Time, page, pageSize int) ([]*model.CalendarEvent, int64, error)
	UpcomingEvents(ctx context.Context, userID string, limit int) ([]*model.CalendarEvent, error)
	EventsByDate(ctx context.Context, userID string, date time.Time) ([]*model.CalendarEvent, error)
	Stats(ctx context.Context, userID string) (map[string]int64, error)

	CreateZoomAccount(ctx context.Context, a *model.ZoomAccount) error
	FindZoomAccountByUserID(ctx context.Context, userID string) (*model.ZoomAccount, error)
	UpdateZoomAccount(ctx context.Context, a *model.ZoomAccount) error
	DeleteZoomAccount(ctx context.Context, id string) error
}

type calendarRepository struct {
	db *gorm.DB
}

func NewCalendarRepository(db *gorm.DB) CalendarRepository {
	return &calendarRepository{db: db}
}

func (r *calendarRepository) CreateEvent(ctx context.Context, e *model.CalendarEvent) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *calendarRepository) FindEventByID(ctx context.Context, id string) (*model.CalendarEvent, error) {
	var e model.CalendarEvent
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *calendarRepository) UpdateEvent(ctx context.Context, e *model.CalendarEvent) error {
	return r.db.WithContext(ctx).Model(e).Select("*").Updates(e).Error
}

func (r *calendarRepository) DeleteEvent(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.CalendarEvent{}, "id = ?", id).Error
}

func (r *calendarRepository) ListEvents(ctx context.Context, userID, search, status, eventType, contactID string, from, to *time.Time, page, pageSize int) ([]*model.CalendarEvent, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.CalendarEvent{}).Where("user_id = ?", userID)

	if search != "" {
		like := "%" + search + "%"
		q = q.Where("title LIKE ? OR description LIKE ? OR contact_name LIKE ?", like, like, like)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	if contactID != "" {
		q = q.Where("contact_id = ?", contactID)
	}
	if from != nil {
		q = q.Where("start_time >= ?", *from)
	}
	if to != nil {
		q = q.Where("start_time <= ?", *to)
	}

	var total int64
	q.Count(&total)

	if pageSize <= 0 {
		pageSize = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var events []*model.CalendarEvent
	err := q.Order("start_time ASC").Limit(pageSize).Offset(offset).Find(&events).Error
	return events, total, err
}

func (r *calendarRepository) UpcomingEvents(ctx context.Context, userID string, limit int) ([]*model.CalendarEvent, error) {
	var events []*model.CalendarEvent
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND start_time >= ? AND status = 'scheduled'", userID, time.Now()).
		Order("start_time ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *calendarRepository) EventsByDate(ctx context.Context, userID string, date time.Time) ([]*model.CalendarEvent, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var events []*model.CalendarEvent
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND start_time >= ? AND start_time < ?", userID, startOfDay, endOfDay).
		Order("start_time ASC").
		Find(&events).Error
	return events, err
}

func (r *calendarRepository) Stats(ctx context.Context, userID string) (map[string]int64, error) {
	stats := make(map[string]int64)
	var total, upcoming, today, zoom int64

	r.db.WithContext(ctx).Model(&model.CalendarEvent{}).Where("user_id = ?", userID).Count(&total)
	r.db.WithContext(ctx).Model(&model.CalendarEvent{}).Where("user_id = ? AND start_time >= ? AND status = 'scheduled'", userID, time.Now()).Count(&upcoming)

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	r.db.WithContext(ctx).Model(&model.CalendarEvent{}).Where("user_id = ? AND start_time >= ? AND start_time < ?", userID, startOfDay, endOfDay).Count(&today)
	r.db.WithContext(ctx).Model(&model.CalendarEvent{}).Where("user_id = ? AND zoom_meeting_id > 0", userID).Count(&zoom)

	stats["total"] = total
	stats["upcoming"] = upcoming
	stats["today"] = today
	stats["zoom"] = zoom
	return stats, nil
}

func (r *calendarRepository) CreateZoomAccount(ctx context.Context, a *model.ZoomAccount) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *calendarRepository) FindZoomAccountByUserID(ctx context.Context, userID string) (*model.ZoomAccount, error) {
	var a model.ZoomAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *calendarRepository) UpdateZoomAccount(ctx context.Context, a *model.ZoomAccount) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *calendarRepository) DeleteZoomAccount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.ZoomAccount{}, "id = ?", id).Error
}

package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"backend/internal/model"
)

type BookingRepository interface {
	GetSettings(ctx context.Context) (*model.BookingSettings, error)
	SaveSettings(ctx context.Context, s *model.BookingSettings) error
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

// GetSettings returns the singleton settings row, or nil if not configured yet.
func (r *bookingRepository) GetSettings(ctx context.Context) (*model.BookingSettings, error) {
	var s model.BookingSettings
	err := r.db.WithContext(ctx).First(&s, "id = ?", model.BookingSettingsID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// SaveSettings upserts the singleton row (ID is always BookingSettingsID).
func (r *bookingRepository) SaveSettings(ctx context.Context, s *model.BookingSettings) error {
	s.ID = model.BookingSettingsID
	return r.db.WithContext(ctx).Save(s).Error
}

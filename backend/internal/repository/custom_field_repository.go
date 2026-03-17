package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "backend/internal/model"
)

type CustomFieldRepository interface {
    Create(ctx context.Context, f *model.CustomField) error
    FindByID(ctx context.Context, id string) (*model.CustomField, error)
    Update(ctx context.Context, f *model.CustomField) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context) ([]model.CustomField, error)
}

type customFieldRepository struct {
    db *gorm.DB
}

func NewCustomFieldRepository(db *gorm.DB) CustomFieldRepository {
    return &customFieldRepository{db: db}
}

func (r *customFieldRepository) Create(ctx context.Context, f *model.CustomField) error {
    return r.db.WithContext(ctx).Create(f).Error
}

func (r *customFieldRepository) FindByID(ctx context.Context, id string) (*model.CustomField, error) {
    var f model.CustomField
    err := r.db.WithContext(ctx).First(&f, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &f, nil
}

func (r *customFieldRepository) Update(ctx context.Context, f *model.CustomField) error {
    return r.db.WithContext(ctx).Save(f).Error
}

func (r *customFieldRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&model.CustomField{}, "id = ?", id).Error
}

func (r *customFieldRepository) List(ctx context.Context) ([]model.CustomField, error) {
    var fields []model.CustomField
    err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&fields).Error
    return fields, err
}

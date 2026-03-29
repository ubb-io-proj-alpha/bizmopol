package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "backend/internal/model"
)

type TagRepository interface {
    Create(ctx context.Context, t *model.Tag) error
    FindByID(ctx context.Context, id string) (*model.Tag, error)
    FindByIDs(ctx context.Context, ids []string) ([]model.Tag, error)
    Update(ctx context.Context, t *model.Tag) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context) ([]model.Tag, error)
}

type tagRepository struct {
    db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
    return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, t *model.Tag) error {
    return r.db.WithContext(ctx).Create(t).Error
}

func (r *tagRepository) FindByID(ctx context.Context, id string) (*model.Tag, error) {
    var t model.Tag
    err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &t, nil
}

func (r *tagRepository) FindByIDs(ctx context.Context, ids []string) ([]model.Tag, error) {
    var tags []model.Tag
    err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&tags).Error
    return tags, err
}

func (r *tagRepository) Update(ctx context.Context, t *model.Tag) error {
    return r.db.WithContext(ctx).Save(t).Error
}

func (r *tagRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&model.Tag{}, "id = ?", id).Error
}

func (r *tagRepository) List(ctx context.Context) ([]model.Tag, error) {
    var tags []model.Tag
    err := r.db.WithContext(ctx).Order("name ASC").Find(&tags).Error
    return tags, err
}

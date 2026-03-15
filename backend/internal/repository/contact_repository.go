package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "backend/internal/model"
)

type ContactRepository interface {
    Create(ctx context.Context, c *model.Contact) error
    FindByID(ctx context.Context, id string) (*model.Contact, error)
    Update(ctx context.Context, c *model.Contact) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, search, status, sortBy, sortDir string) ([]*model.Contact, int64, error)
}

type contactRepository struct {
    db *gorm.DB
}

func NewContactRepository(db *gorm.DB) ContactRepository {
    return &contactRepository{db: db}
}

func (r *contactRepository) Create(ctx context.Context, c *model.Contact) error {
    return r.db.WithContext(ctx).Create(c).Error
}

func (r *contactRepository) FindByID(ctx context.Context, id string) (*model.Contact, error) {
    var c model.Contact
    err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &c, nil
}

func (r *contactRepository) Update(ctx context.Context, c *model.Contact) error {
    return r.db.WithContext(ctx).Save(c).Error
}

func (r *contactRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&model.Contact{}, "id = ?", id).Error
}

func (r *contactRepository) List(ctx context.Context, search, status, sortBy, sortDir string) ([]*model.Contact, int64, error) {
    q := r.db.WithContext(ctx).Model(&model.Contact{})

    if search != "" {
        like := "%" + search + "%"
        q = q.Where("name LIKE ? OR email LIKE ? OR company LIKE ?", like, like, like)
    }
    if status != "" {
        q = q.Where("status = ?", status)
    }

    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    allowed := map[string]bool{"name": true, "email": true, "company": true, "status": true, "created_at": true}
    col := "created_at"
    if allowed[sortBy] {
        col = sortBy
    }
    dir := "DESC"
    if sortDir == "asc" {
        dir = "ASC"
    }
    q = q.Order(col + " " + dir)

    var contacts []*model.Contact
    if err := q.Find(&contacts).Error; err != nil {
        return nil, 0, err
    }
    return contacts, total, nil
}

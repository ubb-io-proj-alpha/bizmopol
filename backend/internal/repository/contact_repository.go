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
    List(ctx context.Context, search, status, sortBy, sortDir string, page, pageSize int) ([]*model.Contact, int64, error)
    AddHistory(ctx context.Context, h *model.ContactHistory) error
    ListHistory(ctx context.Context, contactID string) ([]*model.ContactHistory, error)
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

func (r *contactRepository) List(ctx context.Context, search, status, sortBy, sortDir string, page, pageSize int) ([]*model.Contact, int64, error) {
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

    if pageSize <= 0 {
        pageSize = 20
    }
    if page <= 0 {
        page = 1
    }
    offset := (page - 1) * pageSize
    q = q.Limit(pageSize).Offset(offset)

    var contacts []*model.Contact
    if err := q.Find(&contacts).Error; err != nil {
        return nil, 0, err
    }
    return contacts, total, nil
}

func (r *contactRepository) AddHistory(ctx context.Context, h *model.ContactHistory) error {
    return r.db.WithContext(ctx).Create(h).Error
}

func (r *contactRepository) ListHistory(ctx context.Context, contactID string) ([]*model.ContactHistory, error) {
    var items []*model.ContactHistory
    err := r.db.WithContext(ctx).Where("contact_id = ?", contactID).Order("created_at DESC").Find(&items).Error
    return items, err
}

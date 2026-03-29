package repository

import (
    "context"
    "errors"
    "strings"

    "gorm.io/gorm"

    "backend/internal/model"
)

type ContactRepository interface {
    Create(ctx context.Context, c *model.Contact) error
    FindByID(ctx context.Context, id string) (*model.Contact, error)
    Update(ctx context.Context, c *model.Contact) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, search, status, tags, sortBy, sortDir string, page, pageSize int) ([]*model.Contact, int64, error)
    FindByIDs(ctx context.Context, ids []string) ([]*model.Contact, error)
    FindMembers(ctx context.Context, groupID string, page, pageSize int) ([]*model.Contact, int64, error)
    AddHistory(ctx context.Context, h *model.ContactHistory) error
    ListHistory(ctx context.Context, contactID string) ([]*model.ContactHistory, error)
    ListGroupHistory(ctx context.Context, groupID string, page, pageSize int) ([]*model.ContactHistory, int64, error)
    CountHistory(ctx context.Context, contactID string) (int64, error)
    LastHistoryTime(ctx context.Context, contactID string) (*model.ContactHistory, error)
    SetTags(ctx context.Context, contactID string, tags []model.Tag) error
    SetCustomValues(ctx context.Context, contactID string, values []model.CustomFieldValue) error
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
    err := r.db.WithContext(ctx).
        Preload("Tags").
        Preload("CustomValues").
        First(&c, "id = ?", id).Error
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

func (r *contactRepository) List(ctx context.Context, search, status, tags, sortBy, sortDir string, page, pageSize int) ([]*model.Contact, int64, error) {
    q := r.db.WithContext(ctx).Model(&model.Contact{}).Where("contacts.group_id = ''")

    if search != "" {
        like := "%" + search + "%"
        q = q.Where("contacts.name LIKE ? OR contacts.email LIKE ? OR contacts.company LIKE ?", like, like, like)
    }
    if status != "" {
        q = q.Where("contacts.status = ?", status)
    }
    if tags != "" {
        tagNames := strings.Split(tags, ",")
        q = q.Joins("JOIN contact_tags ON contact_tags.contact_id = contacts.id").
            Joins("JOIN tags ON tags.id = contact_tags.tag_id").
            Where("tags.name IN ?", tagNames).
            Group("contacts.id").
            Having("COUNT(DISTINCT tags.id) >= ?", len(tagNames))
    }

    var total int64
    countQ := r.db.WithContext(ctx).Model(&model.Contact{}).Where("contacts.group_id = ''")
    if search != "" {
        like := "%" + search + "%"
        countQ = countQ.Where("contacts.name LIKE ? OR contacts.email LIKE ? OR contacts.company LIKE ?", like, like, like)
    }
    if status != "" {
        countQ = countQ.Where("contacts.status = ?", status)
    }
    if tags != "" {
        tagNames := strings.Split(tags, ",")
        countQ = countQ.Joins("JOIN contact_tags ON contact_tags.contact_id = contacts.id").
            Joins("JOIN tags ON tags.id = contact_tags.tag_id").
            Where("tags.name IN ?", tagNames).
            Group("contacts.id").
            Having("COUNT(DISTINCT tags.id) >= ?", len(tagNames))
        var ids []string
        if err := countQ.Pluck("contacts.id", &ids).Error; err != nil {
            return nil, 0, err
        }
        total = int64(len(ids))
    } else {
        if err := countQ.Count(&total).Error; err != nil {
            return nil, 0, err
        }
    }

    allowed := map[string]bool{"name": true, "email": true, "company": true, "status": true, "created_at": true, "lead_score": true}
    col := "contacts.created_at"
    if allowed[sortBy] {
        col = "contacts." + sortBy
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
    if err := q.Preload("Tags").Preload("CustomValues").Find(&contacts).Error; err != nil {
        return nil, 0, err
    }
    return contacts, total, nil
}

func (r *contactRepository) FindByIDs(ctx context.Context, ids []string) ([]*model.Contact, error) {
    var contacts []*model.Contact
    err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&contacts).Error
    return contacts, err
}

func (r *contactRepository) FindMembers(ctx context.Context, groupID string, page, pageSize int) ([]*model.Contact, int64, error) {
    q := r.db.WithContext(ctx).Model(&model.Contact{}).Where("group_id = ?", groupID)

    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if pageSize <= 0 {
        pageSize = 10
    }
    if page <= 0 {
        page = 1
    }
    offset := (page - 1) * pageSize

    var members []*model.Contact
    err := q.Order("created_at ASC").Limit(pageSize).Offset(offset).Preload("Tags").Find(&members).Error
    return members, total, err
}

func (r *contactRepository) AddHistory(ctx context.Context, h *model.ContactHistory) error {
    return r.db.WithContext(ctx).Create(h).Error
}

func (r *contactRepository) ListHistory(ctx context.Context, contactID string) ([]*model.ContactHistory, error) {
    var items []*model.ContactHistory
    err := r.db.WithContext(ctx).Where("contact_id = ?", contactID).Order("created_at DESC").Find(&items).Error
    return items, err
}

func (r *contactRepository) ListGroupHistory(ctx context.Context, groupID string, page, pageSize int) ([]*model.ContactHistory, int64, error) {
    memberIDs := []string{}
    r.db.WithContext(ctx).Model(&model.Contact{}).Where("group_id = ?", groupID).Pluck("id", &memberIDs)
    allIDs := append(memberIDs, groupID)

    q := r.db.WithContext(ctx).Model(&model.ContactHistory{}).Where("contact_id IN ?", allIDs)

    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if pageSize <= 0 {
        pageSize = 10
    }
    if page <= 0 {
        page = 1
    }
    offset := (page - 1) * pageSize

    var items []*model.ContactHistory
    err := q.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&items).Error
    return items, total, err
}

func (r *contactRepository) CountHistory(ctx context.Context, contactID string) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&model.ContactHistory{}).Where("contact_id = ?", contactID).Count(&count).Error
    return count, err
}

func (r *contactRepository) LastHistoryTime(ctx context.Context, contactID string) (*model.ContactHistory, error) {
    var h model.ContactHistory
    err := r.db.WithContext(ctx).Where("contact_id = ?", contactID).Order("created_at DESC").First(&h).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &h, nil
}

func (r *contactRepository) SetTags(ctx context.Context, contactID string, tags []model.Tag) error {
    var c model.Contact
    c.ID = contactID
    return r.db.WithContext(ctx).Model(&c).Association("Tags").Replace(tags)
}

func (r *contactRepository) SetCustomValues(ctx context.Context, contactID string, values []model.CustomFieldValue) error {
    if err := r.db.WithContext(ctx).Where("contact_id = ?", contactID).Delete(&model.CustomFieldValue{}).Error; err != nil {
        return err
    }
    if len(values) == 0 {
        return nil
    }
    return r.db.WithContext(ctx).Create(&values).Error
}

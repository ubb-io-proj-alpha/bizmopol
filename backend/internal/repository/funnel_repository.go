package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "backend/internal/model"
)

type FunnelRepository interface {
    CreateFunnel(ctx context.Context, f *model.Funnel) error
    UpdateFunnel(ctx context.Context, f *model.Funnel) error
    DeleteFunnel(ctx context.Context, id string) error
    GetFunnel(ctx context.Context, id string) (*model.Funnel, error)
    ListFunnels(ctx context.Context) ([]model.Funnel, error)
    FindByDomain(ctx context.Context, domain string) (*model.Funnel, error)

    CreatePage(ctx context.Context, p *model.Page) error
    UpdatePage(ctx context.Context, p *model.Page) error
    DeletePage(ctx context.Context, id string) error
    GetPage(ctx context.Context, id string) (*model.Page, error)
    GetPageByPath(ctx context.Context, funnelId string, path string) (*model.Page, error)

    RecordVisit(ctx context.Context, v *model.FunnelVisit) error
}

type funnelRepository struct {
    db *gorm.DB
}

func NewFunnelRepository(db *gorm.DB) FunnelRepository {
    return &funnelRepository{db: db}
}

func (r *funnelRepository) CreateFunnel(ctx context.Context, f *model.Funnel) error {
    return r.db.WithContext(ctx).Create(f).Error
}

func (r *funnelRepository) UpdateFunnel(ctx context.Context, f *model.Funnel) error {
    return r.db.WithContext(ctx).Save(f).Error
}

func (r *funnelRepository) DeleteFunnel(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&model.Funnel{}, "id = ?", id).Error
}

func (r *funnelRepository) GetFunnel(ctx context.Context, id string) (*model.Funnel, error) {
    var f model.Funnel
    err := r.db.WithContext(ctx).Preload("Pages").First(&f, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &f, nil
}

func (r *funnelRepository) ListFunnels(ctx context.Context) ([]model.Funnel, error) {
    var funnels []model.Funnel
    err := r.db.WithContext(ctx).Preload("Pages").Find(&funnels).Error
    return funnels, err
}

func (r *funnelRepository) FindByDomain(ctx context.Context, domain string) (*model.Funnel, error) {
    var f model.Funnel
    err := r.db.WithContext(ctx).Where("custom_domain = ? OR subdomain = ?", domain, domain).First(&f).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &f, nil
}

func (r *funnelRepository) CreatePage(ctx context.Context, p *model.Page) error {
    return r.db.WithContext(ctx).Create(p).Error
}

func (r *funnelRepository) UpdatePage(ctx context.Context, p *model.Page) error {
    return r.db.WithContext(ctx).Save(p).Error
}

func (r *funnelRepository) DeletePage(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&model.Page{}, "id = ?", id).Error
}

func (r *funnelRepository) GetPage(ctx context.Context, id string) (*model.Page, error) {
    var p model.Page
    err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &p, nil
}

func (r *funnelRepository) GetPageByPath(ctx context.Context, funnelId string, path string) (*model.Page, error) {
    var p model.Page
    err := r.db.WithContext(ctx).Where("funnel_id = ? AND path = ?", funnelId, path).First(&p).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &p, nil
}

func (r *funnelRepository) RecordVisit(ctx context.Context, v *model.FunnelVisit) error {
    return r.db.WithContext(ctx).Create(v).Error
}
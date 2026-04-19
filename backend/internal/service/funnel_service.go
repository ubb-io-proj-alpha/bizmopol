package service

import (
    "context"

    "github.com/google/uuid"

    "backend/internal/dto"
    "backend/internal/model"
    "backend/internal/repository"
)

type FunnelService interface {
    CreateFunnel(ctx context.Context, input dto.FunnelCreateRequest) (*model.Funnel, error)
    UpdateFunnel(ctx context.Context, id string, input dto.FunnelUpdateRequest) (*model.Funnel, error)
    DeleteFunnel(ctx context.Context, id string) error
    GetFunnel(ctx context.Context, id string) (*model.Funnel, error)
    ListFunnels(ctx context.Context) ([]model.Funnel, error)

    CreatePage(ctx context.Context, funnelId string, input dto.PageCreateRequest) (*model.Page, error)
    UpdatePage(ctx context.Context, id string, input dto.PageUpdateRequest) (*model.Page, error)
    DeletePage(ctx context.Context, id string) error

    ResolvePage(ctx context.Context, host string, path string, ip string, userAgent string) (*model.Page, error)
    SubmitForm(ctx context.Context, input dto.FunnelSubmissionRequest) error
}

type funnelService struct {
    repo           repository.FunnelRepository
    contactService ContactService
}

func NewFunnelService(r repository.FunnelRepository, cs ContactService) FunnelService {
    return &funnelService{repo: r, contactService: cs}
}

func (s *funnelService) CreateFunnel(ctx context.Context, input dto.FunnelCreateRequest) (*model.Funnel, error) {
    f := &model.Funnel{
        ID:           uuid.NewString(),
        Name:         input.Name,
        Subdomain:    input.Subdomain,
        CustomDomain: input.CustomDomain,
    }
    if err := s.repo.CreateFunnel(ctx, f); err != nil {
        return nil, ErrInternal
    }
    return f, nil
}

func (s *funnelService) UpdateFunnel(ctx context.Context, id string, input dto.FunnelUpdateRequest) (*model.Funnel, error) {
    f, err := s.repo.GetFunnel(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if f == nil {
        return nil, nil
    }
    if input.Name != "" {
        f.Name = input.Name
    }
    f.Subdomain = input.Subdomain
    f.CustomDomain = input.CustomDomain
    if err := s.repo.UpdateFunnel(ctx, f); err != nil {
        return nil, ErrInternal
    }
    return f, nil
}

func (s *funnelService) DeleteFunnel(ctx context.Context, id string) error {
    return s.repo.DeleteFunnel(ctx, id)
}

func (s *funnelService) GetFunnel(ctx context.Context, id string) (*model.Funnel, error) {
    return s.repo.GetFunnel(ctx, id)
}

func (s *funnelService) ListFunnels(ctx context.Context) ([]model.Funnel, error) {
    return s.repo.ListFunnels(ctx)
}

func (s *funnelService) CreatePage(ctx context.Context, funnelId string, input dto.PageCreateRequest) (*model.Page, error) {
    p := &model.Page{
        ID:        uuid.NewString(),
        FunnelID:  funnelId,
        Name:      input.Name,
        Path:      input.Path,
        Structure: input.Structure,
    }
    if err := s.repo.CreatePage(ctx, p); err != nil {
        return nil, ErrInternal
    }
    return p, nil
}

func (s *funnelService) UpdatePage(ctx context.Context, id string, input dto.PageUpdateRequest) (*model.Page, error) {
    p, err := s.repo.GetPage(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if p == nil {
        return nil, nil
    }
    if input.Name != "" {
        p.Name = input.Name
    }
    if input.Path != "" {
        p.Path = input.Path
    }
    if input.Structure != "" {
        p.Structure = input.Structure
    }
    if err := s.repo.UpdatePage(ctx, p); err != nil {
        return nil, ErrInternal
    }
    return p, nil
}

func (s *funnelService) DeletePage(ctx context.Context, id string) error {
    return s.repo.DeletePage(ctx, id)
}

func (s *funnelService) ResolvePage(ctx context.Context, host string, path string, ip string, userAgent string) (*model.Page, error) {
    f, err := s.repo.FindByDomain(ctx, host)
    if err != nil {
        return nil, ErrInternal
    }
    if f == nil {
        return nil, nil
    }

    p, err := s.repo.GetPageByPath(ctx, f.ID, path)
    if err != nil {
        return nil, ErrInternal
    }
    if p == nil {
        return nil, nil
    }

    visit := &model.FunnelVisit{
        ID:        uuid.NewString(),
        FunnelID:  f.ID,
        PageID:    p.ID,
        Ip:        ip,
        UserAgent: userAgent,
    }
    _ = s.repo.RecordVisit(ctx, visit)

    return p, nil
}

func (s *funnelService) SubmitForm(ctx context.Context, input dto.FunnelSubmissionRequest) error {
    contactInput := dto.ContactCreateRequest{
        Name:    input.Name,
        Email:   input.Email,
        Phone:   input.Phone,
        Status:  "lead",
        Notes:   "Utworzono z formularza na lejku",
    }
    _, err := s.contactService.Create(ctx, contactInput)
    return err
}
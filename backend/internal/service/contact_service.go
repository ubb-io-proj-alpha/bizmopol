package service

import (
    "context"

    "github.com/google/uuid"

    "backend/internal/dto"
    "backend/internal/model"
    "backend/internal/repository"
)

type ContactService interface {
    Create(ctx context.Context, input dto.ContactCreateRequest) (*dto.ContactResponse, error)
    GetByID(ctx context.Context, id string) (*dto.ContactResponse, error)
    Update(ctx context.Context, id string, input dto.ContactUpdateRequest) (*dto.ContactResponse, error)
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, q dto.ContactQuery) (*dto.ContactListResponse, error)
}

type contactService struct {
    repo repository.ContactRepository
}

func NewContactService(r repository.ContactRepository) ContactService {
    return &contactService{repo: r}
}

func toContactResponse(c *model.Contact) dto.ContactResponse {
    return dto.ContactResponse{
        ID:        c.ID,
        Name:      c.Name,
        Email:     c.Email,
        Phone:     c.Phone,
        Company:   c.Company,
        Status:    c.Status,
        Notes:     c.Notes,
        CreatedAt: c.CreatedAt,
        UpdatedAt: c.UpdatedAt,
    }
}

func (s *contactService) Create(ctx context.Context, input dto.ContactCreateRequest) (*dto.ContactResponse, error) {
    status := input.Status
    if status == "" {
        status = "lead"
    }
    c := &model.Contact{
        ID:      uuid.NewString(),
        Name:    input.Name,
        Email:   input.Email,
        Phone:   input.Phone,
        Company: input.Company,
        Status:  status,
        Notes:   input.Notes,
    }
    if err := s.repo.Create(ctx, c); err != nil {
        return nil, ErrInternal
    }
    r := toContactResponse(c)
    return &r, nil
}

func (s *contactService) GetByID(ctx context.Context, id string) (*dto.ContactResponse, error) {
    c, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if c == nil {
        return nil, nil
    }
    r := toContactResponse(c)
    return &r, nil
}

func (s *contactService) Update(ctx context.Context, id string, input dto.ContactUpdateRequest) (*dto.ContactResponse, error) {
    c, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if c == nil {
        return nil, nil
    }
    if input.Name != "" {
        c.Name = input.Name
    }
    c.Email = input.Email
    c.Phone = input.Phone
    c.Company = input.Company
    if input.Status != "" {
        c.Status = input.Status
    }
    c.Notes = input.Notes
    if err := s.repo.Update(ctx, c); err != nil {
        return nil, ErrInternal
    }
    r := toContactResponse(c)
    return &r, nil
}

func (s *contactService) Delete(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}

func (s *contactService) List(ctx context.Context, q dto.ContactQuery) (*dto.ContactListResponse, error) {
    contacts, total, err := s.repo.List(ctx, q.Search, q.Status, q.SortBy, q.SortDir)
    if err != nil {
        return nil, ErrInternal
    }
    items := make([]dto.ContactResponse, 0, len(contacts))
    for _, c := range contacts {
        items = append(items, toContactResponse(c))
    }
    return &dto.ContactListResponse{Data: items, Total: total}, nil
}

package service

import (
    "context"
    "math"

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
    AddHistory(ctx context.Context, contactID, userID string, input dto.AddHistoryRequest) (*dto.ContactHistoryResponse, error)
    ListHistory(ctx context.Context, contactID string) ([]dto.ContactHistoryResponse, error)
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

func toHistoryResponse(h *model.ContactHistory) dto.ContactHistoryResponse {
    return dto.ContactHistoryResponse{
        ID:          h.ID,
        ContactID:   h.ContactID,
        Action:      h.Action,
        Description: h.Description,
        UserID:      h.UserID,
        CreatedAt:   h.CreatedAt,
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
    userID, _ := ctx.Value("userID").(string)
    _ = s.repo.AddHistory(ctx, &model.ContactHistory{
        ID:          uuid.NewString(),
        ContactID:   c.ID,
        Action:      "created",
        Description: "Kontakt został utworzony",
        UserID:      userID,
    })
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
    oldStatus := c.Status
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
    userID, _ := ctx.Value("userID").(string)
    desc := "Kontakt został zaktualizowany"
    if input.Status != "" && input.Status != oldStatus {
        desc = "Status zmieniony z " + oldStatus + " na " + input.Status
    }
    _ = s.repo.AddHistory(ctx, &model.ContactHistory{
        ID:          uuid.NewString(),
        ContactID:   c.ID,
        Action:      "updated",
        Description: desc,
        UserID:      userID,
    })
    r := toContactResponse(c)
    return &r, nil
}

func (s *contactService) Delete(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}

func (s *contactService) List(ctx context.Context, q dto.ContactQuery) (*dto.ContactListResponse, error) {
    pageSize := q.PageSize
    if pageSize <= 0 {
        pageSize = 20
    }
    page := q.Page
    if page <= 0 {
        page = 1
    }

    contacts, total, err := s.repo.List(ctx, q.Search, q.Status, q.SortBy, q.SortDir, page, pageSize)
    if err != nil {
        return nil, ErrInternal
    }
    items := make([]dto.ContactResponse, 0, len(contacts))
    for _, c := range contacts {
        items = append(items, toContactResponse(c))
    }
    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
    return &dto.ContactListResponse{
        Data:       items,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}

func (s *contactService) AddHistory(ctx context.Context, contactID, userID string, input dto.AddHistoryRequest) (*dto.ContactHistoryResponse, error) {
    h := &model.ContactHistory{
        ID:          uuid.NewString(),
        ContactID:   contactID,
        Action:      input.Action,
        Description: input.Description,
        UserID:      userID,
    }
    if err := s.repo.AddHistory(ctx, h); err != nil {
        return nil, ErrInternal
    }
    r := toHistoryResponse(h)
    return &r, nil
}

func (s *contactService) ListHistory(ctx context.Context, contactID string) ([]dto.ContactHistoryResponse, error) {
    items, err := s.repo.ListHistory(ctx, contactID)
    if err != nil {
        return nil, ErrInternal
    }
    result := make([]dto.ContactHistoryResponse, 0, len(items))
    for _, h := range items {
        result = append(result, toHistoryResponse(h))
    }
    return result, nil
}

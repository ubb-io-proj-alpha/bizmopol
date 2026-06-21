package service

import (
    "context"

    "github.com/google/uuid"

    "backend/internal/dto"
    "backend/internal/model"
    "backend/internal/repository"
)

type CustomFieldService interface {
    List(ctx context.Context) ([]dto.CustomFieldResponse, error)
    Create(ctx context.Context, input dto.CustomFieldCreateRequest) (*dto.CustomFieldResponse, error)
    Update(ctx context.Context, id string, input dto.CustomFieldUpdateRequest) (*dto.CustomFieldResponse, error)
    Delete(ctx context.Context, id string) error
}

type customFieldService struct {
    repo repository.CustomFieldRepository
}

func NewCustomFieldService(r repository.CustomFieldRepository) CustomFieldService {
    return &customFieldService{repo: r}
}

func toCFResponse(f *model.CustomField) dto.CustomFieldResponse {
    return dto.CustomFieldResponse{
        ID:        f.ID,
        Name:      f.Name,
        FieldType: f.FieldType,
        Visible:   f.Visible,
        SortOrder: f.SortOrder,
        CreatedAt: f.CreatedAt,
        UpdatedAt: f.UpdatedAt,
    }
}

func (s *customFieldService) List(ctx context.Context) ([]dto.CustomFieldResponse, error) {
    debugLog("CustomField.List")
    fields, err := s.repo.List(ctx)
    if err != nil {
        debugLogResult("CustomField.List", err)
        return nil, ErrInternal
    }
    debugLogResult("CustomField.List", nil, "count", len(fields))
    result := make([]dto.CustomFieldResponse, 0, len(fields))
    for i := range fields {
        result = append(result, toCFResponse(&fields[i]))
    }
    return result, nil
}

func (s *customFieldService) Create(ctx context.Context, input dto.CustomFieldCreateRequest) (*dto.CustomFieldResponse, error) {
    debugLog("CustomField.Create", "name", input.Name, "type", input.FieldType)
    ft := input.FieldType
    if ft == "" {
        ft = "text"
    }
    visible := true
    if input.Visible != nil {
        visible = *input.Visible
    }
    f := &model.CustomField{
        ID:        uuid.NewString(),
        Name:      input.Name,
        FieldType: ft,
        Visible:   visible,
        SortOrder: input.SortOrder,
    }
    if err := s.repo.Create(ctx, f); err != nil {
        debugLogResult("CustomField.Create", err)
        return nil, ErrInternal
    }
    debugLogResult("CustomField.Create", nil, "id", f.ID)
    r := toCFResponse(f)
    return &r, nil
}

func (s *customFieldService) Update(ctx context.Context, id string, input dto.CustomFieldUpdateRequest) (*dto.CustomFieldResponse, error) {
    debugLog("CustomField.Update", "id", id, "name", input.Name)
    f, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if f == nil {
        return nil, nil
    }
    if input.Name != "" {
        f.Name = input.Name
    }
    if input.FieldType != "" {
        f.FieldType = input.FieldType
    }
    if input.Visible != nil {
        f.Visible = *input.Visible
    }
    f.SortOrder = input.SortOrder
    if err := s.repo.Update(ctx, f); err != nil {
        debugLogResult("CustomField.Update", err)
        return nil, ErrInternal
    }
    debugLogResult("CustomField.Update", nil, "id", f.ID)
    r := toCFResponse(f)
    return &r, nil
}

func (s *customFieldService) Delete(ctx context.Context, id string) error {
    debugLog("CustomField.Delete", "id", id)
    err := s.repo.Delete(ctx, id)
    debugLogResult("CustomField.Delete", err)
    return err
}

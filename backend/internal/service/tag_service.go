package service

import (
    "context"

    "github.com/google/uuid"

    "backend/internal/dto"
    "backend/internal/model"
    "backend/internal/repository"
)

type TagService interface {
    List(ctx context.Context) ([]dto.TagResponse, error)
    Create(ctx context.Context, input dto.TagCreateRequest) (*dto.TagResponse, error)
    Update(ctx context.Context, id string, input dto.TagUpdateRequest) (*dto.TagResponse, error)
    Delete(ctx context.Context, id string) error
}

type tagService struct {
    repo repository.TagRepository
}

func NewTagService(r repository.TagRepository) TagService {
    return &tagService{repo: r}
}

func (s *tagService) List(ctx context.Context) ([]dto.TagResponse, error) {
    debugLog("Tag.List")
    tags, err := s.repo.List(ctx)
    if err != nil {
        debugLogResult("Tag.List", err)
        return nil, ErrInternal
    }
    debugLogResult("Tag.List", nil, "count", len(tags))
    result := make([]dto.TagResponse, 0, len(tags))
    for _, t := range tags {
        result = append(result, dto.TagResponse{ID: t.ID, Name: t.Name, Color: t.Color})
    }
    return result, nil
}

func (s *tagService) Create(ctx context.Context, input dto.TagCreateRequest) (*dto.TagResponse, error) {
    debugLog("Tag.Create", "name", input.Name, "color", input.Color)
    color := input.Color
    if color == "" {
        color = "#6366f1"
    }
    t := &model.Tag{
        ID:    uuid.NewString(),
        Name:  input.Name,
        Color: color,
    }
    if err := s.repo.Create(ctx, t); err != nil {
        debugLogResult("Tag.Create", err)
        return nil, ErrInternal
    }
    debugLogResult("Tag.Create", nil, "id", t.ID)
    r := dto.TagResponse{ID: t.ID, Name: t.Name, Color: t.Color}
    return &r, nil
}

func (s *tagService) Update(ctx context.Context, id string, input dto.TagUpdateRequest) (*dto.TagResponse, error) {
    debugLog("Tag.Update", "id", id, "name", input.Name, "color", input.Color)
    t, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if t == nil {
        return nil, nil
    }
    if input.Name != "" {
        t.Name = input.Name
    }
    if input.Color != "" {
        t.Color = input.Color
    }
    if err := s.repo.Update(ctx, t); err != nil {
        debugLogResult("Tag.Update", err)
        return nil, ErrInternal
    }
    debugLogResult("Tag.Update", nil, "id", t.ID)
    r := dto.TagResponse{ID: t.ID, Name: t.Name, Color: t.Color}
    return &r, nil
}

func (s *tagService) Delete(ctx context.Context, id string) error {
    debugLog("Tag.Delete", "id", id)
    err := s.repo.Delete(ctx, id)
    debugLogResult("Tag.Delete", err)
    return err
}

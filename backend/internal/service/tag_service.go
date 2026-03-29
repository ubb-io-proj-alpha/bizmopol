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
    tags, err := s.repo.List(ctx)
    if err != nil {
        return nil, ErrInternal
    }
    result := make([]dto.TagResponse, 0, len(tags))
    for _, t := range tags {
        result = append(result, dto.TagResponse{ID: t.ID, Name: t.Name, Color: t.Color})
    }
    return result, nil
}

func (s *tagService) Create(ctx context.Context, input dto.TagCreateRequest) (*dto.TagResponse, error) {
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
        return nil, ErrInternal
    }
    r := dto.TagResponse{ID: t.ID, Name: t.Name, Color: t.Color}
    return &r, nil
}

func (s *tagService) Update(ctx context.Context, id string, input dto.TagUpdateRequest) (*dto.TagResponse, error) {
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
        return nil, ErrInternal
    }
    r := dto.TagResponse{ID: t.ID, Name: t.Name, Color: t.Color}
    return &r, nil
}

func (s *tagService) Delete(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}

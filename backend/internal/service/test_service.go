package service 

import (
	"context"
	"errors"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/dto"
)

var (
	ErrTestExists = errors.New("test already exists")
	ErrInternal = errors.New("an unexpected error occurred")
)

type TestService interface {
	Create(ctx context.Context, input dto.TestCreateRequest) (*dto.TestResponse, error)
	GetAll(ctx context.Context) (*dto.TestListResponse, error)
}

type testService struct {
	repo repository.TestRepository
}

func NewTestService(r repository.TestRepository) TestService {
	return &testService{repo: r}
}

func (service *testService) Create(ctx context.Context, input dto.TestCreateRequest) (*dto.TestResponse, error) {
	test, err := service.repo.FindByName(ctx, input.Name)
	
	if err != nil {
        return nil, ErrInternal
    }

	if test != nil {
		return nil, ErrTestExists
	}

	newTest := model.Test {
		Name: input.Name,
	}

	if err := service.repo.Create(ctx, &newTest); err != nil {
		return nil, err
	}

    return &dto.TestResponse{
		ID: newTest.ID,
		Name: newTest.Name,
		CreatedAt: newTest.CreatedAt,
		UpdatedAt: newTest.UpdatedAt,
	}, nil
}

func (service *testService)	GetAll(ctx context.Context) (*dto.TestListResponse, error) {
	tests, err := service.repo.GetAll(ctx, 100, 0)
	if err != nil {
		return nil, ErrInternal
	}

	testDTOs := make([]dto.TestResponse, 0, len(tests))
	for _, t := range tests {
		testDTOs = append(testDTOs, dto.TestResponse{
			ID: t.ID,
			Name: t.Name,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		})
	}
	return &dto.TestListResponse {
		Data: testDTOs,
	}, nil
}

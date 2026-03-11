package repository

import (
	"context"
	"errors"
	"gorm.io/gorm"

	"backend/internal/model"
)

type TestRepository interface {
	Create(ctx context.Context, p *model.Test) error
	FindByName(ctx context.Context, name string) (*model.Test, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Test, error)
}

type testRepository struct {
	db *gorm.DB
}

func NewTestRepository(db *gorm.DB) TestRepository {
	return &testRepository{db: db}
}

func (testRepo *testRepository) Create(ctx context.Context, t *model.Test) error {
	return testRepo.db.WithContext(ctx).Create(t).Error;
}

func (testRepo *testRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.Test, error) {
	var products []*model.Test

	err := testRepo.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("id DESC").
		Find(&products).
		Error

	if err != nil {
		return nil, err
	}

	return products, nil
}

func (testRepo *testRepository) FindByName(ctx context.Context, name string)  (*model.Test, error) {
    var user model.Test

	err := testRepo.db.WithContext(ctx).Where("name = ?", name).First(&user).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    
    return &user, nil
}
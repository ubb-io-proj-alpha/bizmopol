package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "backend/internal/model"
)

type UserRepository interface {
    Create(ctx context.Context, u *model.User) error
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    FindByID(ctx context.Context, id string) (*model.User, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *model.User) error {
    return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

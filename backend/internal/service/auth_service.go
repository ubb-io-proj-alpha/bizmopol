package service

import (
    "context"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"

    "backend/internal/dto"
    "backend/internal/model"
    "backend/internal/repository"
)

var (
    ErrUserExists         = errors.New("user_already_exists")
    ErrInvalidCredentials = errors.New("invalid_credentials")
)

type AuthService interface {
    Register(ctx context.Context, input dto.RegisterRequest) (*dto.AuthResponse, error)
    Login(ctx context.Context, input dto.LoginRequest) (*dto.AuthResponse, error)
}

type authService struct {
    repo      repository.UserRepository
    jwtSecret string
}

func NewAuthService(r repository.UserRepository, jwtSecret string) AuthService {
    return &authService{repo: r, jwtSecret: jwtSecret}
}

func (s *authService) Register(ctx context.Context, input dto.RegisterRequest) (*dto.AuthResponse, error) {
    existing, err := s.repo.FindByEmail(ctx, input.Email)
    if err != nil {
        return nil, ErrInternal
    }
    if existing != nil {
        return nil, ErrUserExists
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, ErrInternal
    }

    user := &model.User{
        ID:           uuid.NewString(),
        Email:        input.Email,
        PasswordHash: string(hash),
        Name:         input.Name,
        Role:         model.RoleUser,
    }
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, ErrInternal
    }

    token, err := s.generateToken(user)
    if err != nil {
        return nil, ErrInternal
    }

    return &dto.AuthResponse{
        Token: token,
        User: dto.UserResponse{
            ID:        user.ID,
            Email:     user.Email,
            Name:      user.Name,
            Role:      string(user.Role),
            CreatedAt: user.CreatedAt,
        },
    }, nil
}

func (s *authService) Login(ctx context.Context, input dto.LoginRequest) (*dto.AuthResponse, error) {
    user, err := s.repo.FindByEmail(ctx, input.Email)
    if err != nil {
        return nil, ErrInternal
    }
    if user == nil {
        return nil, ErrInvalidCredentials
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
        return nil, ErrInvalidCredentials
    }

    token, err := s.generateToken(user)
    if err != nil {
        return nil, ErrInternal
    }

    return &dto.AuthResponse{
        Token: token,
        User: dto.UserResponse{
            ID:        user.ID,
            Email:     user.Email,
            Name:      user.Name,
            Role:      string(user.Role),
            CreatedAt: user.CreatedAt,
        },
    }, nil
}

func (s *authService) generateToken(user *model.User) (string, error) {
    claims := jwt.MapClaims{
        "sub":   user.ID,
        "email": user.Email,
        "role":  user.Role,
        "exp":   time.Now().Add(72 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}

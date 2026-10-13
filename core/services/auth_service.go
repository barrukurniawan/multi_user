package services

import (
	"context"
	"errors"
	"strings"

	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/barrukurniawan/multi_user/core/helpers"
	"github.com/barrukurniawan/multi_user/core/models"
	"github.com/barrukurniawan/multi_user/core/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req dto.LoginRequest) (string, error)
}

type authService struct {
	userRepo repositories.UserRepository
	jwt      *helpers.JWTManager
}

func NewAuthService(userRepo repositories.UserRepository, jwtManager *helpers.JWTManager) AuthService {
	return &authService{userRepo: userRepo, jwt: jwtManager}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, helpers.Conflict("EMAIL_EXISTS", "email sudah terdaftar")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	teamID, err := s.userRepo.FindOrCreateTeam(ctx, strings.TrimSpace(req.TeamName))
	if err != nil {
		return nil, err
	}

	user := &models.User{
		TeamID:   teamID,
		Name:     strings.TrimSpace(req.Name),
		Email:    email,
		Password: string(hashed),
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrEmailExists) {
			return nil, helpers.Conflict("EMAIL_EXISTS", "email sudah terdaftar")
		}
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return "", helpers.Unauthorized("INVALID_CREDENTIALS", "email atau password salah")
	}

	return s.jwt.Generate(user.ID)
}

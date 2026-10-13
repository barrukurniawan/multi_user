package tests

// Test authentication (nilai tambah): register, email dobel, login, dan JWT.

import (
	"context"
	"testing"

	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/barrukurniawan/multi_user/core/helpers"
	"github.com/barrukurniawan/multi_user/core/models"
	"github.com/barrukurniawan/multi_user/core/repositories"
	"github.com/barrukurniawan/multi_user/core/services"
)

// fakeUserRepo menyimpan user di map (email → user)
type fakeUserRepo struct {
	repositories.UserRepository
	users map[string]*models.User
}

func (f *fakeUserRepo) FindOrCreateTeam(ctx context.Context, name string) (string, error) {
	return "team-1", nil
}

func (f *fakeUserRepo) Create(ctx context.Context, user *models.User) error {
	user.ID = "user-" + user.Email
	f.users[user.Email] = user
	return nil
}

func (f *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return f.users[email], nil
}

func TestAuth(t *testing.T) {
	jwtManager := helpers.NewJWTManager("test-secret", 24)
	auth := services.NewAuthService(&fakeUserRepo{users: map[string]*models.User{}}, jwtManager)
	ctx := context.Background()
	req := dto.RegisterRequest{Name: "Budi", Email: "budi@mail.com", Password: "rahasia123", TeamName: "alpha"}

	if _, err := auth.Register(ctx, req); err != nil {
		t.Fatalf("register gagal: %v", err)
	}
	if _, err := auth.Register(ctx, req); errorCode(err) != "EMAIL_EXISTS" {
		t.Fatalf("register email sama: dapat %v, seharusnya EMAIL_EXISTS", err)
	}

	_, err := auth.Login(ctx, dto.LoginRequest{Email: "budi@mail.com", Password: "salah"})
	if errorCode(err) != "INVALID_CREDENTIALS" {
		t.Fatalf("login password salah: dapat %v, seharusnya INVALID_CREDENTIALS", err)
	}

	token, err := auth.Login(ctx, dto.LoginRequest{Email: "budi@mail.com", Password: "rahasia123"})
	if err != nil {
		t.Fatalf("login gagal: %v", err)
	}
	if _, err := jwtManager.Validate(token); err != nil {
		t.Fatalf("token JWT tidak valid: %v", err)
	}
}

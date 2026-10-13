package repositories

import (
	"context"
	"errors"

	"github.com/barrukurniawan/multi_user/core/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrEmailExists = errors.New("email already exists")

type UserRepository interface {
	FindOrCreateTeam(ctx context.Context, teamName string) (string, error)
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
}

type userRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{DB: db}
}

func (r *userRepository) FindOrCreateTeam(ctx context.Context, teamName string) (string, error) {
	team := models.Team{Name: teamName}
	err := r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"name"}),
		}).
		Create(&team).Error
	return team.ID, err
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	err := r.DB.WithContext(ctx).Create(user).Error

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailExists
	}
	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.findOne(ctx, "email = ?", email)
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *userRepository) findOne(ctx context.Context, query string, arg any) (*models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Where(query, arg).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

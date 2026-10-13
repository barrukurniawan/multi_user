package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/barrukurniawan/multi_user/core/models"
	"gorm.io/gorm"
)

type TaskFilter struct {
	UserID string
	Status string
	Search string
	Limit  int
	Offset int
}

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	FindAll(ctx context.Context, filter TaskFilter) ([]models.Task, int, error)
	FindByID(ctx context.Context, id string) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	UpdateAssignee(ctx context.Context, tx *gorm.DB, taskID, assigneeID string) error
}

type TaskLogRepository interface {
	Create(ctx context.Context, tx *gorm.DB, log *models.TaskLog) error
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type taskRepository struct {
	DB *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{DB: db}
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	return r.DB.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) FindAll(ctx context.Context, f TaskFilter) ([]models.Task, int, error) {

	filtered := func() *gorm.DB {
		q := r.DB.WithContext(ctx).Model(&models.Task{}).
			Where("(user_id = ? OR assignee_id = ?)", f.UserID, f.UserID)
		if f.Status != "" {
			q = q.Where("status = ?", f.Status)
		}
		if f.Search != "" {
			q = q.Where("title ILIKE ?", "%"+f.Search+"%")
		}
		return q
	}

	var total int64
	if err := filtered().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tasks := []models.Task{}
	err := filtered().Order("created_at DESC").Limit(f.Limit).Offset(f.Offset).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, int(total), nil
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*models.Task, error) {
	var task models.Task
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) Update(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now()
	return r.DB.WithContext(ctx).Model(task).
		Select("title", "description", "status", "updated_at").
		Updates(task).Error
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&models.Task{}).Error
}

func (r *taskRepository) UpdateAssignee(ctx context.Context, tx *gorm.DB, taskID, assigneeID string) error {
	result := tx.WithContext(ctx).Model(&models.Task{}).Where("id = ?", taskID).
		Updates(map[string]any{"assignee_id": assigneeID, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found")
	}
	return nil
}

type taskLogRepository struct{}

func NewTaskLogRepository() TaskLogRepository {
	return &taskLogRepository{}
}

func (r *taskLogRepository) Create(ctx context.Context, tx *gorm.DB, log *models.TaskLog) error {
	return tx.WithContext(ctx).Create(log).Error
}

type txManager struct {
	DB *gorm.DB
}

func NewTxManager(db *gorm.DB) TxManager {
	return &txManager{DB: db}
}

func (m *txManager) WithTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return m.DB.WithContext(ctx).Transaction(fn)
}

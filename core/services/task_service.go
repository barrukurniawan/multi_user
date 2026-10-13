package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/barrukurniawan/multi_user/core/helpers"
	"github.com/barrukurniawan/multi_user/core/models"
	"github.com/barrukurniawan/multi_user/core/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const idempotencyWindow = 24 * time.Hour

var errKeyProcessing = helpers.Conflict("IDEMPOTENCY_KEY_PROCESSING", "request dengan Idempotency-Key ini masih diproses, coba lagi sebentar")

type TaskService interface {
	Create(ctx context.Context, userID, idempotencyKey string, req dto.CreateTaskRequest) (*models.Task, int, error)
	List(ctx context.Context, userID string, query dto.ListTaskQuery) ([]models.Task, dto.PaginationMeta, error)
	GetByID(ctx context.Context, userID, taskID string) (*models.Task, error)
	Update(ctx context.Context, userID, taskID string, req dto.UpdateTaskRequest) (*models.Task, error)
	Delete(ctx context.Context, userID, taskID string) error
	Assign(ctx context.Context, userID, taskID, assigneeID string) (*models.Task, error)
}

type taskService struct {
	taskRepo  repositories.TaskRepository
	logRepo   repositories.TaskLogRepository
	idemRepo  repositories.IdempotencyRepository
	userRepo  repositories.UserRepository
	txManager repositories.TxManager
	notifier  NotificationService
}

func NewTaskService(
	taskRepo repositories.TaskRepository,
	logRepo repositories.TaskLogRepository,
	idemRepo repositories.IdempotencyRepository,
	userRepo repositories.UserRepository,
	txManager repositories.TxManager,
	notifier NotificationService,
) TaskService {
	return &taskService{
		taskRepo:  taskRepo,
		logRepo:   logRepo,
		idemRepo:  idemRepo,
		userRepo:  userRepo,
		txManager: txManager,
		notifier:  notifier,
	}
}

func (s *taskService) Create(ctx context.Context, userID, idempotencyKey string, req dto.CreateTaskRequest) (*models.Task, int, error) {
	if _, err := uuid.Parse(idempotencyKey); err != nil {
		return nil, 0, helpers.BadRequest("INVALID_IDEMPOTENCY_KEY", "header Idempotency-Key wajib diisi dengan format UUID")
	}

	if req.Status == "" {
		req.Status = models.StatusTodo
	}
	if !models.IsValidStatus(req.Status) {
		return nil, 0, invalidStatusError()
	}

	task, statusCode, err := s.findSavedResponse(ctx, userID, idempotencyKey)
	if err != nil || task != nil {
		return task, statusCode, err
	}

	reserved, err := s.idemRepo.Reserve(ctx, idempotencyKey, userID, idempotencyWindow)
	if err != nil {
		return nil, 0, err
	}
	if !reserved {

		task, statusCode, err = s.findSavedResponse(ctx, userID, idempotencyKey)
		if err == nil && task == nil {
			err = errKeyProcessing
		}
		return task, statusCode, err
	}

	task = &models.Task{
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		Status:      req.Status,
		UserID:      userID,
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		_ = s.idemRepo.Release(ctx, idempotencyKey)
		return nil, 0, err
	}

	body, err := json.Marshal(task)
	if err == nil {
		err = s.idemRepo.SaveResponse(ctx, idempotencyKey, userID, body, http.StatusCreated)
	}
	if err != nil {

		slog.ErrorContext(ctx, "failed to save idempotency response", "err", err, "idempotency_key", idempotencyKey)
	}
	return task, http.StatusCreated, nil
}

func (s *taskService) findSavedResponse(ctx context.Context, userID, key string) (*models.Task, int, error) {
	saved, err := s.idemRepo.Find(ctx, key)
	if err != nil || saved == nil {
		return nil, 0, err
	}
	if saved.UserID != userID {
		return nil, 0, helpers.Conflict("IDEMPOTENCY_KEY_CONFLICT", "Idempotency-Key sudah dipakai untuk request lain")
	}
	if saved.Response == nil {
		return nil, 0, errKeyProcessing
	}

	var task models.Task
	if err := json.Unmarshal(saved.Response, &task); err != nil {
		return nil, 0, err
	}
	return &task, saved.StatusCode, nil
}

func (s *taskService) List(ctx context.Context, userID string, query dto.ListTaskQuery) ([]models.Task, dto.PaginationMeta, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Status != "" && !models.IsValidStatus(query.Status) {
		return nil, dto.PaginationMeta{}, invalidStatusError()
	}

	tasks, total, err := s.taskRepo.FindAll(ctx, repositories.TaskFilter{
		UserID: userID,
		Status: query.Status,
		Search: strings.TrimSpace(query.Search),
		Limit:  query.Limit,
		Offset: (query.Page - 1) * query.Limit,
	})
	if err != nil {
		return nil, dto.PaginationMeta{}, err
	}

	return tasks, dto.NewPaginationMeta(query.Page, query.Limit, total), nil
}

func (s *taskService) GetByID(ctx context.Context, userID, taskID string) (*models.Task, error) {
	return s.findTask(ctx, userID, taskID, false)
}

func (s *taskService) Update(ctx context.Context, userID, taskID string, req dto.UpdateTaskRequest) (*models.Task, error) {
	task, err := s.findTask(ctx, userID, taskID, true)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return nil, helpers.BadRequest("INVALID_INPUT", "title tidak boleh kosong")
		}
		task.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		if !models.IsValidStatus(*req.Status) {
			return nil, invalidStatusError()
		}
		task.Status = *req.Status
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) Delete(ctx context.Context, userID, taskID string) error {
	if _, err := s.findTask(ctx, userID, taskID, true); err != nil {
		return err
	}
	return s.taskRepo.Delete(ctx, taskID)
}

func (s *taskService) Assign(ctx context.Context, userID, taskID, assigneeID string) (*models.Task, error) {
	task, err := s.findTask(ctx, userID, taskID, true)
	if err != nil {
		return nil, err
	}

	owner, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	assignee, err := s.userRepo.FindByID(ctx, assigneeID)
	if err != nil {
		return nil, err
	}
	if owner == nil || assignee == nil {
		return nil, helpers.NotFound("USER_NOT_FOUND", "user yang akan di-assign tidak ditemukan")
	}
	if owner.TeamID != assignee.TeamID {
		return nil, helpers.Forbidden("DIFFERENT_TEAM", "task hanya bisa di-assign ke user dalam tim yang sama")
	}

	oldAssignee := ""
	if task.AssigneeID != nil {
		oldAssignee = *task.AssigneeID
	}

	err = s.txManager.WithTx(ctx, func(tx *gorm.DB) error {
		if err := s.taskRepo.UpdateAssignee(ctx, tx, taskID, assigneeID); err != nil {
			return err
		}

		taskLog := &models.TaskLog{
			TaskID:    taskID,
			Action:    "assigned",
			OldValue:  oldAssignee,
			NewValue:  assigneeID,
			ChangedBy: userID,
		}
		if err := s.logRepo.Create(ctx, tx, taskLog); err != nil {
			return err
		}

		return s.notifier.NotifyTaskAssigned(ctx, taskID, assigneeID, userID)
	})
	if err != nil {
		return nil, err
	}

	return s.taskRepo.FindByID(ctx, taskID)
}

func (s *taskService) findTask(ctx context.Context, userID, taskID string, ownerOnly bool) (*models.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	isOwner := task != nil && task.UserID == userID
	isAssignee := task != nil && task.AssigneeID != nil && *task.AssigneeID == userID

	if !isOwner && !isAssignee {
		return nil, helpers.NotFound("TASK_NOT_FOUND", "task tidak ditemukan")
	}
	if ownerOnly && !isOwner {
		return nil, helpers.Forbidden("NOT_TASK_OWNER", "hanya pemilik task yang boleh melakukan aksi ini")
	}
	return task, nil
}

func invalidStatusError() error {
	return helpers.BadRequest("INVALID_STATUS", "status harus salah satu dari: todo, in_progress, done")
}

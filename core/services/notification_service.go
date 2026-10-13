package services

import (
	"context"
	"log/slog"
)

type NotificationService interface {
	NotifyTaskAssigned(ctx context.Context, taskID, assigneeID, assignedBy string) error
}

type logNotificationService struct{}

func NewNotificationService() NotificationService {
	return &logNotificationService{}
}

func (s *logNotificationService) NotifyTaskAssigned(ctx context.Context, taskID, assigneeID, assignedBy string) error {
	slog.InfoContext(ctx, "notification sent: task assigned",
		"task_id", taskID,
		"assignee_id", assigneeID,
		"assigned_by", assignedBy,
	)
	return nil
}

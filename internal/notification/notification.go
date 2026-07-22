// internal/notification/notification.go
package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ochochecharles/task-management-api/internal/db"
)

type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) TaskAssigned(ctx context.Context, assigneeID, taskID uuid.UUID, taskTitle string) error {
	_, err := s.queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:     assigneeID,
		Type:       "task_assigned",
		Message:    fmt.Sprintf("You were assigned to task: %s", taskTitle),
		EntityType: "task",
		EntityID:   taskID,
	})
	return err
}

func (s *Service) TaskStatusChanged(ctx context.Context, notifyUserID, taskID uuid.UUID, taskTitle, newStatus string) error {
	_, err := s.queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:     notifyUserID,
		Type:       "task_status_changed",
		Message:    fmt.Sprintf("Task \"%s\" status changed to %s", taskTitle, newStatus),
		EntityType: "task",
		EntityID:   taskID,
	})
	return err
}

func (s *Service) ProjectMemberAdded(ctx context.Context, userID, projectID uuid.UUID, projectName string) error {
	_, err := s.queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:     userID,
		Type:       "project_added",
		Message:    fmt.Sprintf("You were added to project: %s", projectName),
		EntityType: "project",
		EntityID:   projectID,
	})
	return err
}
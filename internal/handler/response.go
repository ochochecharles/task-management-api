package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/ochochecharles/task-management-api/internal/db"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Status      string    `json:"status"`
	DueDate     *string   `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectMemberResponse struct {
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

type TaskResponse struct {
	ID          uuid.UUID  `json:"id"`
	ProjectID   uuid.UUID  `json:"project_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *string    `json:"due_date"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	AssignedTo  *uuid.UUID `json:"assigned_to"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func toProjectResponse(p db.Project) ProjectResponse {
	var description *string
	if p.Description.Valid {
		description = &p.Description.String
	}

	var dueDate *string
	if p.DueDate.Valid {
		formatted := p.DueDate.Time.Format("2006-01-02")
		dueDate = &formatted
	}

	return ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: description,
		OwnerID:     p.OwnerID,
		Status:      p.Status,
		DueDate:     dueDate,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toTaskResponse(t db.Task) TaskResponse {
	var description *string
	if t.Description.Valid {
		description = &t.Description.String
	}

	var dueDate *string
	if t.DueDate.Valid {
		formatted := t.DueDate.Time.Format("2006-01-02")
		dueDate = &formatted
	}

	var assignedTo *uuid.UUID
	if t.AssignedTo.Valid {
		assignedTo = &t.AssignedTo.UUID
	}

	return TaskResponse{
		ID:          t.ID,
		ProjectID:   t.ProjectID,
		Title:       t.Title,
		Description: description,
		Status:      t.Status,
		Priority:    t.Priority,
		DueDate:     dueDate,
		CreatedBy:   t.CreatedBy,
		AssignedTo:  assignedTo,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
package handler

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ochochecharles/task-management-api/internal/db"
	"github.com/ochochecharles/task-management-api/internal/middleware"
)

type TaskHandler struct {
	queries *db.Queries
}

func NewTaskHandler(queries *db.Queries) *TaskHandler {
	return &TaskHandler{queries: queries}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Context().Value(middleware.UserIDKey).(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var body struct {
		Title       string  `json:"title"`
		Description *string `json:"description"`
		Priority    string  `json:"priority"`
		DueDate     *string `json:"due_date"`
		AssignedTo  *string `json:"assigned_to"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	if body.Priority == "" {
		body.Priority = "medium"
	}

	task, err := h.queries.CreateTask(r.Context(), db.CreateTaskParams{
		ProjectID:   projectID,
		Title:       body.Title,
		Description: nullableString(body.Description),
		Priority:    body.Priority,
		DueDate:     nullableDate(body.DueDate),
		CreatedBy:   userID,
		AssignedTo:  nullableUUID(body.AssignedTo),
	})
	if err != nil {
		slog.Error("failed to create task", "error", err)
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toTaskResponse(task))
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "taskID"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	task, err := h.queries.GetTaskByID(r.Context(), taskID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toTaskResponse(task))
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	tasks, err := h.queries.ListTasksByProject(r.Context(), projectID)
	if err != nil {
		slog.Error("failed to list tasks", "error", err)
		http.Error(w, "failed to list tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		response[i] = toTaskResponse(t)
	}
	json.NewEncoder(w).Encode(response)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Context().Value(middleware.UserIDKey).(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskID"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	_, role, err := h.getTaskAndCheckAccess(r, taskID, userID, projectID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	if role != "creator" {
		http.Error(w, "only the task creator can update this task", http.StatusForbidden)
		return
	}

	var body struct {
		Title       string  `json:"title"`
		Description *string `json:"description"`
		Status      string  `json:"status"`
		Priority    string  `json:"priority"`
		DueDate     *string `json:"due_date"`
		AssignedTo  *string `json:"assigned_to"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	task, err := h.queries.UpdateTask(r.Context(), db.UpdateTaskParams{
		ID:          taskID,
		Title:       body.Title,
		Description: nullableString(body.Description),
		Status:      body.Status,
		Priority:    body.Priority,
		DueDate:     nullableDate(body.DueDate),
		AssignedTo:  nullableUUID(body.AssignedTo),
	})
	if err != nil {
		slog.Error("failed to update task", "error", err)
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toTaskResponse(task))
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Context().Value(middleware.UserIDKey).(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskID"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	_, role, err := h.getTaskAndCheckAccess(r, taskID, userID, projectID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	if role != "creator" && role != "owner" {
		http.Error(w, "only the task creator or project owner can delete this task", http.StatusForbidden)
		return
	}

	if err := h.queries.DeleteTask(r.Context(), taskID); err != nil {
		slog.Error("failed to delete task", "error", err)
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Context().Value(middleware.UserIDKey).(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskID"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	_, role, err := h.getTaskAndCheckAccess(r, taskID, userID, projectID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	if role != "assignee" && role != "creator" {
		http.Error(w, "only the assignee or task creator can update the status", http.StatusForbidden)
		return
	}

	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	task, err := h.queries.UpdateTaskStatus(r.Context(), db.UpdateTaskStatusParams{
		ID:     taskID,
		Status: body.Status,
	})
	if err != nil {
		slog.Error("failed to update task status", "error", err)
		http.Error(w, "failed to update task status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toTaskResponse(task))
}

// helper functions for nullable types
func nullableString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullableDate(s *string) sql.NullTime {
	if s == nil {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func nullableUUID(s *string) uuid.NullUUID {
	if s == nil {
		return uuid.NullUUID{}
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

// authorization helper
func (h *TaskHandler) getTaskAndCheckAccess(r *http.Request, taskID uuid.UUID, userID uuid.UUID, projectID uuid.UUID) (db.Task, string, error) {
	task, err := h.queries.GetTaskByID(r.Context(), taskID)
	if err != nil {
		return db.Task{}, "", err
	}

	if task.CreatedBy == userID {
		return task, "creator", nil
	}

	member, err := h.queries.GetProjectMember(r.Context(), db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err == nil && member.Role == "owner" {
		return task, "owner", nil
	}

	if task.AssignedTo.Valid && task.AssignedTo.UUID == userID {
		return task, "assignee", nil
	}

	return task, "member", nil
}
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lib/pq"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ochochecharles/task-management-api/internal/db"
	"github.com/ochochecharles/task-management-api/internal/middleware"
)

type ProjectHandler struct {
	queries *db.Queries
}

func NewProjectHandler(queries *db.Queries) *ProjectHandler {
	return &ProjectHandler{queries: queries}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Context().Value(middleware.UserIDKey).(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	var body struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		DueDate     *string `json:"due_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	project, err := h.queries.CreateProject(r.Context(), db.CreateProjectParams{
		Name:        body.Name,
		Description: nullableString(body.Description),
		OwnerID:     userID,
		DueDate:     nullableDate(body.DueDate),
	})
	if err != nil {
		slog.Error("failed to create project", "error", err)
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	// also add the creator as owner in project_members
	if err := h.queries.AddProjectMember(r.Context(), db.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "owner",
	}); err != nil {
		slog.Error("failed to add project member", "error", err)
		http.Error(w, "failed to add project member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toProjectResponse(project))
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	project, err := h.queries.GetProjectByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toProjectResponse(project))
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Context().Value(middleware.UserIDKey).(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	projects, err := h.queries.ListProjectsByMember(r.Context(), userID)
	if err != nil {
		slog.Error("failed to list projects", "error", err)
		http.Error(w, "failed to list projects", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		response[i] = toProjectResponse(p)
	}
	json.NewEncoder(w).Encode(response)
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var body struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Status      string  `json:"status"`
		DueDate     *string `json:"due_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	project, err := h.queries.UpdateProject(r.Context(), db.UpdateProjectParams{
		ID:          projectID,
		Name:        body.Name,
		Description: nullableString(body.Description),
		Status:      body.Status,
		DueDate:     nullableDate(body.DueDate),
	})
	if err != nil {
		slog.Error("failed to update project", "error", err)
		http.Error(w, "failed to update project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toProjectResponse(project))
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	if err := h.queries.DeleteProject(r.Context(), projectID); err != nil {
		slog.Error("failed to delete project", "error", err)
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var body struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	if body.Role == "" {
		body.Role = "member"
	}

	// look up user by email
	user, err := h.queries.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		http.Error(w, "no user found with that email", http.StatusNotFound)
		return
	}

	if err := h.queries.AddProjectMember(r.Context(), db.AddProjectMemberParams{
		ProjectID: projectID,
		UserID:    user.ID,
		Role:      body.Role,
	}); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			http.Error(w, "user is already a member of this project", http.StatusConflict)
			return
		}
		slog.Error("failed to add member", "error", err)
		http.Error(w, "failed to add member", http.StatusInternalServerError)
		return
	}

	member, err := h.queries.GetProjectMember(r.Context(), db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    user.ID,
	})
	if err != nil {
		slog.Error("failed to fetch added member", "error", err)
		http.Error(w, "failed to fetch added member", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ProjectMemberResponse{
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		Name:      member.Name,
		Email:     member.Email,
		Role:      member.Role,
		JoinedAt:  member.JoinedAt,
	})
}

func (h *ProjectHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	memberID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.queries.RemoveProjectMember(r.Context(), db.RemoveProjectMemberParams{
		ProjectID: projectID,
		UserID:    memberID,
	}); err != nil {
		slog.Error("failed to remove member", "error", err)
		http.Error(w, "failed to remove member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	memberID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	member, err := h.queries.GetProjectMember(r.Context(), db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    memberID,
	})
	if err != nil {
		http.Error(w, "member not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProjectMemberResponse{
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		Name:      member.Name,
		Email:     member.Email,
		Role:      member.Role,
		JoinedAt:  member.JoinedAt,
	})
}

func (h *ProjectHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	members, err := h.queries.ListProjectMembers(r.Context(), projectID)
	if err != nil {
		slog.Error("failed to list members", "error", err)
		http.Error(w, "failed to list members", http.StatusInternalServerError)
		return
	}

	response := make([]ProjectMemberResponse, len(members))
	for i, m := range members {
		response[i] = ProjectMemberResponse{
			ProjectID: m.ProjectID,
			UserID:    m.UserID,
			Name:      m.Name,
			Email:     m.Email,
			Role:      m.Role,
			JoinedAt:  m.JoinedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
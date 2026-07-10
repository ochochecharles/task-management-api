package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/ochochecharles/task-management-api/internal/db"
	"github.com/ochochecharles/task-management-api/internal/middleware"
	"github.com/lib/pq"
)

type UserHandler struct {
	queries *db.Queries
}

func NewUserHandler(queries *db.Queries) *UserHandler {
	return &UserHandler{queries: queries}
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Context().Value(middleware.UserIDKey).(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	if err := h.queries.DeleteUser(r.Context(), userID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			http.Error(w, "cannot delete account while you still own projects, delete or transfer them first", http.StatusConflict)
			return
		}
		slog.Error("failed to delete user", "error", err)
		http.Error(w, "failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
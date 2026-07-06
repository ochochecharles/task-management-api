package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/ochochecharles/task-management-api/internal/db"
	"github.com/ochochecharles/task-management-api/internal/handler"
	"github.com/ochochecharles/task-management-api/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("failed to load .env file", "error", err)
		os.Exit(1)
	}

	conn, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		slog.Error("failed to open database connection", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	slog.Info("database connection established")

	queries := db.New(conn)

	authHandler := handler.NewAuthHandler(queries)
	projectHandler := handler.NewProjectHandler(queries)
	taskHandler := handler.NewTaskHandler(queries)

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// public routes
	r.Get("/auth/google", authHandler.GoogleLogin)
	r.Get("/auth/google/callback", authHandler.GoogleCallback)

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		r.Post("/projects", projectHandler.CreateProject)
		r.Get("/projects", projectHandler.ListProjects)
		r.Get("/projects/{id}", projectHandler.GetProject)
		r.Put("/projects/{id}", projectHandler.UpdateProject)
		r.Delete("/projects/{id}", projectHandler.DeleteProject)

		r.Post("/projects/{id}/members", projectHandler.AddMember)
		r.Delete("/projects/{id}/members/{userID}", projectHandler.RemoveMember)

		r.Post("/projects/{id}/tasks", taskHandler.CreateTask)
		r.Get("/projects/{id}/tasks", taskHandler.ListTasks)
		r.Get("/projects/{id}/tasks/{taskID}", taskHandler.GetTask)
		r.Put("/projects/{id}/tasks/{taskID}", taskHandler.UpdateTask)
		r.Delete("/projects/{id}/tasks/{taskID}", taskHandler.DeleteTask)
	})

	port := os.Getenv("PORT")
	slog.Info("server starting", "port", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
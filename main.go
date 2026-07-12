package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/cors"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/ochochecharles/task-management-api/internal/db"
	"github.com/ochochecharles/task-management-api/internal/handler"
	"github.com/ochochecharles/task-management-api/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
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
	userHandler := handler.NewUserHandler(queries)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{os.Getenv("FRONTEND_URL")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

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
		r.Get("/projects/{id}/members", projectHandler.ListMembers)
		r.Get("/projects/{id}/members/{userID}", projectHandler.GetMember)

		r.Post("/projects/{id}/tasks", taskHandler.CreateTask)
		r.Get("/projects/{id}/tasks", taskHandler.ListTasks)
		r.Get("/projects/{id}/tasks/{taskID}", taskHandler.GetTask)
		r.Put("/projects/{id}/tasks/{taskID}", taskHandler.UpdateTask)
		r.Delete("/projects/{id}/tasks/{taskID}", taskHandler.DeleteTask)

		r.Delete("/users/me", userHandler.DeleteUser)
		r.Get("/users/me", userHandler.GetMe)
	})

	port := os.Getenv("PORT")
	slog.Info("server starting", "port", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
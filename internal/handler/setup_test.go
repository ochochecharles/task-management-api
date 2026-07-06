package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/ochochecharles/task-management-api/internal/db"
	"github.com/ochochecharles/task-management-api/internal/middleware"
)

type testApp struct {
	router  *chi.Mux
	queries *db.Queries
	conn    *sql.DB
}

func setupTestApp(t *testing.T) *testApp {
	t.Helper()

	if err := godotenv.Load("../../.env"); err != nil {
		slog.Warn("no .env file found, using environment variables")
	}

	connStr := "host=localhost port=5433 user=ochoche password=ochoche dbname=task_management_test sslmode=disable"

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := conn.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	queries := db.New(conn)

	projectHandler := NewProjectHandler(queries)
	taskHandler := NewTaskHandler(queries)

	r := chi.NewRouter()

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

	return &testApp{
		router:  r,
		queries: queries,
		conn:    conn,
	}
}

func (a *testApp) cleanup(t *testing.T) {
	t.Helper()

	_, err := a.conn.Exec(`
		TRUNCATE TABLE tasks, project_members, projects, users RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("failed to cleanup test database: %v", err)
	}
}

func (a *testApp) makeRequest(method, url, token string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	a.router.ServeHTTP(rr, req)
	return rr
}

func createTestUser(t *testing.T, queries *db.Queries) (db.User, string) {
	t.Helper()

	user, err := queries.CreateUser(context.Background(), db.CreateUserParams{
		GoogleID:  "test-google-id",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarUrl: sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	return user, tokenString
}
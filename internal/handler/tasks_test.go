package handler

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateTask(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup(t)

	_, token := createTestUser(t, app.queries)

	// create a project first
	rr := app.makeRequest("POST", "/projects", token, map[string]interface{}{"name": "Test Project"})
	var project ProjectResponse
	json.NewDecoder(rr.Body).Decode(&project)

	t.Run("success", func(t *testing.T) {
		body := map[string]interface{}{
			"title":       "Test Task",
			"description": "A test task",
			"priority":    "high",
			"due_date":    "2025-11-30",
		}

		rr := app.makeRequest("POST", "/projects/"+project.ID.String()+"/tasks", token, body)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", rr.Code)
		}

		var response TaskResponse
		json.NewDecoder(rr.Body).Decode(&response)

		if response.Title != "Test Task" {
			t.Errorf("expected title 'Test Task', got '%s'", response.Title)
		}

		if response.Priority != "high" {
			t.Errorf("expected priority 'high', got '%s'", response.Priority)
		}

		if response.Status != "pending" {
			t.Errorf("expected status 'pending', got '%s'", response.Status)
		}
	})

	t.Run("missing title", func(t *testing.T) {
		body := map[string]interface{}{
			"priority": "high",
		}

		rr := app.makeRequest("POST", "/projects/"+project.ID.String()+"/tasks", token, body)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		body := map[string]interface{}{
			"title": "Test Task",
		}

		rr := app.makeRequest("POST", "/projects/"+project.ID.String()+"/tasks", "", body)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})
}

func TestListTasks(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup(t)

	_, token := createTestUser(t, app.queries)

	// create a project and two tasks
	rr := app.makeRequest("POST", "/projects", token, map[string]interface{}{"name": "Test Project"})
	var project ProjectResponse
	json.NewDecoder(rr.Body).Decode(&project)

	app.makeRequest("POST", "/projects/"+project.ID.String()+"/tasks", token, map[string]interface{}{"title": "Task One"})
	app.makeRequest("POST", "/projects/"+project.ID.String()+"/tasks", token, map[string]interface{}{"title": "Task Two"})

	t.Run("success", func(t *testing.T) {
		rr := app.makeRequest("GET", "/projects/"+project.ID.String()+"/tasks", token, nil)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		var response []TaskResponse
		json.NewDecoder(rr.Body).Decode(&response)

		if len(response) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(response))
		}
	})
}

func TestUpdateTask(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup(t)

	_, token := createTestUser(t, app.queries)

	// create a project and a task
	rr := app.makeRequest("POST", "/projects", token, map[string]interface{}{"name": "Test Project"})
	var project ProjectResponse
	json.NewDecoder(rr.Body).Decode(&project)

	rr = app.makeRequest("POST", "/projects/"+project.ID.String()+"/tasks", token, map[string]interface{}{"title": "Old Title"})
	var task TaskResponse
	json.NewDecoder(rr.Body).Decode(&task)

	t.Run("success", func(t *testing.T) {
		body := map[string]interface{}{
			"title":    "New Title",
			"status":   "in_progress",
			"priority": "low",
		}

		rr := app.makeRequest("PUT", "/projects/"+project.ID.String()+"/tasks/"+task.ID.String(), token, body)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		var response TaskResponse
		json.NewDecoder(rr.Body).Decode(&response)

		if response.Title != "New Title" {
			t.Errorf("expected title 'New Title', got '%s'", response.Title)
		}

		if response.Status != "in_progress" {
			t.Errorf("expected status 'in_progress', got '%s'", response.Status)
		}
	})
}

func TestDeleteTask(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup(t)

	_, token := createTestUser(t, app.queries)

	// create a project and a task
	rr := app.makeRequest("POST", "/projects", token, map[string]interface{}{"name": "Test Project"})
	var project ProjectResponse
	json.NewDecoder(rr.Body).Decode(&project)

	rr = app.makeRequest("POST", "/projects/"+project.ID.String()+"/tasks", token, map[string]interface{}{"title": "To Delete"})
	var task TaskResponse
	json.NewDecoder(rr.Body).Decode(&task)

	t.Run("success", func(t *testing.T) {
		rr := app.makeRequest("DELETE", "/projects/"+project.ID.String()+"/tasks/"+task.ID.String(), token, nil)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rr.Code)
		}

		// confirm it's gone
		rr = app.makeRequest("GET", "/projects/"+project.ID.String()+"/tasks/"+task.ID.String(), token, nil)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404 after delete, got %d", rr.Code)
		}
	})
}
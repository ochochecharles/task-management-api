package handler

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateProject(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup(t)

	_, token := createTestUser(t, app.queries)

	t.Run("success", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "Test Project",
			"description": "A test project",
			"due_date":    "2025-12-31",
		}

		rr := app.makeRequest("POST", "/projects", token, body)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", rr.Code)
		}

		var response ProjectResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Name != "Test Project" {
			t.Errorf("expected name 'Test Project', got '%s'", response.Name)
		}

		if response.Status != "active" {
			t.Errorf("expected status 'active', got '%s'", response.Status)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		body := map[string]interface{}{
			"description": "A test project",
		}

		rr := app.makeRequest("POST", "/projects", token, body)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Test Project",
		}

		rr := app.makeRequest("POST", "/projects", "", body)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})
}

func TestGetProject(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup(t)

	_, token := createTestUser(t, app.queries)

	// create a project first
	body := map[string]interface{}{
		"name": "Test Project",
	}
	rr := app.makeRequest("POST", "/projects", token, body)
	var created ProjectResponse
	json.NewDecoder(rr.Body).Decode(&created)

	t.Run("success", func(t *testing.T) {
		rr := app.makeRequest("GET", "/projects/"+created.ID.String(), token, nil)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		var response ProjectResponse
		json.NewDecoder(rr.Body).Decode(&response)

		if response.ID != created.ID {
			t.Errorf("expected project ID %s, got %s", created.ID, response.ID)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		rr := app.makeRequest("GET", "/projects/not-a-uuid", token, nil)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		rr := app.makeRequest("GET", "/projects/"+created.ID.String(), "", nil)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})
}

func TestListProjects(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup(t)

	_, token := createTestUser(t, app.queries)

	// create two projects
	app.makeRequest("POST", "/projects", token, map[string]interface{}{"name": "Project One"})
	app.makeRequest("POST", "/projects", token, map[string]interface{}{"name": "Project Two"})

	t.Run("success", func(t *testing.T) {
		rr := app.makeRequest("GET", "/projects", token, nil)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		var response []ProjectResponse
		json.NewDecoder(rr.Body).Decode(&response)

		if len(response) != 2 {
			t.Errorf("expected 2 projects, got %d", len(response))
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		rr := app.makeRequest("GET", "/projects", "", nil)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})
}

func TestUpdateProject(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup(t)

	_, token := createTestUser(t, app.queries)

	// create a project first
	rr := app.makeRequest("POST", "/projects", token, map[string]interface{}{"name": "Old Name"})
	var created ProjectResponse
	json.NewDecoder(rr.Body).Decode(&created)

	t.Run("success", func(t *testing.T) {
		body := map[string]interface{}{
			"name":   "New Name",
			"status": "active",
		}

		rr := app.makeRequest("PUT", "/projects/"+created.ID.String(), token, body)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		var response ProjectResponse
		json.NewDecoder(rr.Body).Decode(&response)

		if response.Name != "New Name" {
			t.Errorf("expected name 'New Name', got '%s'", response.Name)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		body := map[string]interface{}{
			"status": "active",
		}

		rr := app.makeRequest("PUT", "/projects/"+created.ID.String(), token, body)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})
}

func TestDeleteProject(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup(t)

	_, token := createTestUser(t, app.queries)

	// create a project first
	rr := app.makeRequest("POST", "/projects", token, map[string]interface{}{"name": "To Delete"})
	var created ProjectResponse
	json.NewDecoder(rr.Body).Decode(&created)

	t.Run("success", func(t *testing.T) {
		rr := app.makeRequest("DELETE", "/projects/"+created.ID.String(), token, nil)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rr.Code)
		}

		// confirm it's gone
		rr = app.makeRequest("GET", "/projects/"+created.ID.String(), token, nil)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404 after delete, got %d", rr.Code)
		}
	})
}
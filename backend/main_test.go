package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestHandler() http.Handler {
	app := &App{store: NewTaskStore()}
	return NewRouter(app)
}

func TestHealthAndReady(t *testing.T) {
	handler := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /health: expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /ready: expected 200, got %d", rec.Code)
	}
}

func TestListTasksReturnsSeedData(t *testing.T) {
	handler := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success=true")
	}
}

func TestTaskCRUDLifecycle(t *testing.T) {
	handler := newTestHandler()

	// --- create ---
	body, _ := json.Marshal(map[string]string{"title": "Write documentation", "status": "todo"})
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}
	id := created.Data.ID
	if id == "" {
		t.Fatal("expected a non-empty task id")
	}

	// --- create without title should fail ---
	body, _ = json.Marshal(map[string]string{"description": "no title given"})
	req = httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing title, got %d", rec.Code)
	}

	// --- get by id ---
	req = httptest.NewRequest(http.MethodGet, "/api/tasks/"+id, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", rec.Code)
	}

	// --- update status ---
	body, _ = json.Marshal(map[string]string{"status": "done"})
	req = httptest.NewRequest(http.MethodPut, "/api/tasks/"+id, bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d", rec.Code)
	}

	// --- delete ---
	req = httptest.NewRequest(http.MethodDelete, "/api/tasks/"+id, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", rec.Code)
	}

	// --- get after delete should 404 ---
	req = httptest.NewRequest(http.MethodGet, "/api/tasks/"+id, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec.Code)
	}
}

package main

import (
	"encoding/json"
	"net/http"
)

// App holds shared dependencies for HTTP handlers.
type App struct {
	store *TaskStore
}

// APIResponse is the standard JSON envelope returned by every endpoint.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func apiInfoHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"service":   "cloudtask-backend",
		"version":   "1.0.0",
		"language":  "Go",
		"endpoints": []string{"/api/tasks", "/health", "/ready"},
	})
}

func (app *App) listTasksHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: app.store.GetAll()})
}

func (app *App) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := app.store.GetByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, APIResponse{Success: false, Message: "Task not found"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: task})
}

type taskRequestBody struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

func (app *App) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var body taskRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Message: "Invalid JSON body"})
		return
	}
	if body.Title == nil || *body.Title == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Message: `Field "title" is required`})
		return
	}

	description := ""
	if body.Description != nil {
		description = *body.Description
	}
	status := ""
	if body.Status != nil {
		status = *body.Status
	}

	task := app.store.Create(*body.Title, description, status)
	writeJSON(w, http.StatusCreated, APIResponse{Success: true, Data: task})
}

func (app *App) updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body taskRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Message: "Invalid JSON body"})
		return
	}

	task, ok := app.store.Update(id, body.Title, body.Description, body.Status)
	if !ok {
		writeJSON(w, http.StatusNotFound, APIResponse{Success: false, Message: "Task not found"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: task})
}

func (app *App) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !app.store.Delete(id) {
		writeJSON(w, http.StatusNotFound, APIResponse{Success: false, Message: "Task not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

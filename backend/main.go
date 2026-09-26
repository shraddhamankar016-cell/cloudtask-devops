package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

// corsMiddleware allows the static frontend (served from a different
// origin/port in local dev) to call this API from the browser.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs every request with its method, path and duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// NewRouter wires up all HTTP routes. Uses Go 1.22's method+pattern
// matching in net/http.ServeMux, so no third-party router is needed.
func NewRouter(app *App) http.Handler {
	mux := http.NewServeMux()

	// Kubernetes probes
	mux.HandleFunc("GET /health", healthHandler) // liveness
	mux.HandleFunc("GET /ready", readyHandler)   // readiness

	mux.HandleFunc("GET /api", apiInfoHandler)

	mux.HandleFunc("GET /api/tasks", app.listTasksHandler)
	mux.HandleFunc("POST /api/tasks", app.createTaskHandler)
	mux.HandleFunc("GET /api/tasks/{id}", app.getTaskHandler)
	mux.HandleFunc("PUT /api/tasks/{id}", app.updateTaskHandler)
	mux.HandleFunc("DELETE /api/tasks/{id}", app.deleteTaskHandler)

	return loggingMiddleware(corsMiddleware(mux))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := &App{store: NewTaskStore()}
	handler := NewRouter(app)

	log.Printf("CloudTask backend (Go) listening on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

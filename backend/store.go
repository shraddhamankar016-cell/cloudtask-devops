package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Task represents a single to-do item.
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

// TaskStore is a thread-safe in-memory data store.
//
// This keeps the demo app dependency-free and easy to run anywhere
// (local machine, Docker container, Kubernetes pod) with no database
// connection required out of the box.
//
// In a real production rollout, this is the single place you would
// swap out for a persistent store — e.g. Amazon RDS (Postgres), which
// is already provisioned for you in /terraform. Only this file would
// need to change; the HTTP handlers talk to it through the same
// method signatures regardless of backing store.
type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
	order []string // preserves insertion order for listing
}

// NewTaskStore creates a store pre-populated with a couple of sample tasks.
func NewTaskStore() *TaskStore {
	s := &TaskStore{tasks: make(map[string]*Task)}
	s.seed()
	return s
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *TaskStore) seed() {
	seedTasks := []*Task{
		{
			ID:          generateID(),
			Title:       "Set up CI/CD pipeline",
			Description: "Configure Jenkins and GitHub Actions for automated build/test/deploy",
			Status:      "in-progress",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Title:       "Provision AWS infrastructure with Terraform",
			Description: "VPC, EKS cluster, ECR repositories",
			Status:      "todo",
			CreatedAt:   time.Now(),
		},
	}
	for _, t := range seedTasks {
		s.tasks[t.ID] = t
		s.order = append(s.order, t.ID)
	}
}

// GetAll returns every task in insertion order.
func (s *TaskStore) GetAll() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Task, 0, len(s.order))
	for _, id := range s.order {
		if t, ok := s.tasks[id]; ok {
			result = append(result, t)
		}
	}
	return result
}

// GetByID returns a task by its ID.
func (s *TaskStore) GetByID(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

// Create adds a new task and returns it.
func (s *TaskStore) Create(title, description, status string) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	if status == "" {
		status = "todo"
	}

	t := &Task{
		ID:          generateID(),
		Title:       title,
		Description: description,
		Status:      status,
		CreatedAt:   time.Now(),
	}
	s.tasks[t.ID] = t
	s.order = append(s.order, t.ID)
	return t
}

// Update applies partial updates to an existing task.
func (s *TaskStore) Update(id string, title, description, status *string) (*Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return nil, false
	}
	if title != nil && *title != "" {
		t.Title = *title
	}
	if description != nil {
		t.Description = *description
	}
	if status != nil && *status != "" {
		t.Status = *status
	}
	now := time.Now()
	t.UpdatedAt = &now
	return t, true
}

// Delete removes a task by ID. Returns false if it did not exist.
func (s *TaskStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	for i, oid := range s.order {
		if oid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}

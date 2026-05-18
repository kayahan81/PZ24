package storage

import (
	"fmt"
	"sync"
	"time"

	"tech-ip-sem2/services/tasks/internal/models"
)

type MemoryStorage struct {
	mu    sync.RWMutex
	tasks map[string]models.Task
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		tasks: make(map[string]models.Task),
	}
}

func (s *MemoryStorage) Create(task models.Task) models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.ID = fmt.Sprintf("t_%d", time.Now().UnixNano())[:8]
	task.CreatedAt = time.Now()
	s.tasks[task.ID] = task
	return task
}

func (s *MemoryStorage) GetAll() []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *MemoryStorage) GetByID(id string) (models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	return task, ok
}

func (s *MemoryStorage) Update(id string, updated models.Task) (models.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return models.Task{}, false
	}

	if updated.Title != "" {
		task.Title = updated.Title
	}
	if updated.Description != "" {
		task.Description = updated.Description
	}
	if updated.DueDate != "" {
		task.DueDate = updated.DueDate
	}
	task.Done = updated.Done

	s.tasks[id] = task
	return task, true
}

func (s *MemoryStorage) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; ok {
		delete(s.tasks, id)
		return true
	}
	return false
}

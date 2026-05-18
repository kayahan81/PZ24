package handler

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"

	"tech-ip-sem2/services/tasks/internal/models"
	"tech-ip-sem2/services/tasks/internal/repository"
	"tech-ip-sem2/shared/middleware"

	"go.uber.org/zap"
)

// Санитизация description
func sanitizeDescription(desc string) string {
	// Экранируем HTML теги
	return html.EscapeString(desc)
}

type TasksHandler struct {
	repo repository.TaskRepository // ← просто используем существующий интерфейс
	log  *zap.Logger
}

func NewTasksHandler(repo repository.TaskRepository, log *zap.Logger) *TasksHandler {
	return &TasksHandler{repo: repo, log: log}
}

func (h *TasksHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := h.log.With(zap.String("request_id", requestID), zap.String("handler", "SearchTasks"))

	keyword := r.URL.Query().Get("title")
	if keyword == "" {
		http.Error(w, `{"error":"title parameter required"}`, http.StatusBadRequest)
		return
	}

	tasks, err := h.repo.SearchByTitle(keyword)
	if err != nil {
		log.Error("search failed", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	log.Info("search completed", zap.String("keyword", keyword), zap.Int("count", len(tasks)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks) // ← возвращаем МАССИВ, а не одну задачу
}

// SearchTasksUnsafe - УЯЗВИМЫЙ поиск (только для демонстрации SQL инъекции!)
func (h *TasksHandler) SearchTasksUnsafe(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := h.log.With(zap.String("request_id", requestID), zap.String("handler", "SearchTasksUnsafe"))

	keyword := r.URL.Query().Get("title")
	if keyword == "" {
		http.Error(w, `{"error":"title parameter required"}`, http.StatusBadRequest)
		return
	}

	tasks, err := h.repo.SearchByTitleUnsafe(keyword)
	if err != nil {
		log.Error("unsafe search failed", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	log.Warn("unsafe search executed", zap.String("keyword", keyword), zap.Int("count", len(tasks)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TasksHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := h.log.With(zap.String("request_id", requestID), zap.String("handler", "CreateTask"))

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Warn("invalid request body", zap.Error(err))
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		log.Warn("title required")
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}

	task.Description = sanitizeDescription(task.Description)

	created, err := h.repo.Create(task)
	if err != nil {
		log.Error("failed to create task", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	log.Info("task created", zap.String("task_id", created.ID), zap.String("title", created.Title))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *TasksHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := h.log.With(zap.String("request_id", requestID), zap.String("handler", "GetTasks"))

	tasks, err := h.repo.GetAll()
	if err != nil {
		log.Error("failed to GetAll", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	log.Info("tasks retrieved", zap.Int("count", len(tasks)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TasksHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := h.log.With(zap.String("request_id", requestID), zap.String("handler", "GetTask"))

	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if id == "" {
		log.Warn("task id missing")
		http.Error(w, `{"error":"task id required"}`, http.StatusBadRequest)
		return
	}

	task, err := h.repo.GetByID(id)
	if err != nil {
		log.Error("failed to GetByID", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	log.Info("task retrieved", zap.String("task_id", id))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TasksHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := h.log.With(zap.String("request_id", requestID), zap.String("handler", "UpdateTask"))

	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if id == "" {
		log.Warn("task id missing")
		http.Error(w, `{"error":"task id required"}`, http.StatusBadRequest)
		return
	}

	var update models.Task
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Warn("invalid request body", zap.Error(err))
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	updated, err := h.repo.Update(id, update)
	if err != nil {
		log.Warn("task not found", zap.String("task_id", id))
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}

	log.Info("task updated", zap.String("task_id", id))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *TasksHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := h.log.With(zap.String("request_id", requestID), zap.String("handler", "DeleteTask"))

	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if id == "" {
		log.Warn("task id missing")
		http.Error(w, `{"error":"task id required"}`, http.StatusBadRequest)
		return
	}

	err := h.repo.Delete(id)
	if err != nil {
		log.Error("failed to delete task", zap.Error(err), zap.String("task_id", id))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	log.Info("task deleted", zap.String("task_id", id))
	w.WriteHeader(http.StatusNoContent)
}

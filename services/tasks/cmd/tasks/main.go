package main

import (
	"net/http"
	"os"

	"tech-ip-sem2/services/tasks/internal/handler"
	"tech-ip-sem2/services/tasks/internal/middleware"
	"tech-ip-sem2/services/tasks/internal/repository"
	"tech-ip-sem2/shared/logger"
	sharedMiddleware "tech-ip-sem2/shared/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("tasks")
	defer log.Sync()

	port := os.Getenv("TASKS_PORT")
	if port == "" {
		port = "8082"
	}

	log.Info("starting tasks service",
		zap.String("http_port", port),
	)

	// Инициализация БД
	repo, err := repository.NewPostgresRepo(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer repo.Close()

	tasksHandler := handler.NewTasksHandler(repo, log)

	// Роутер
	mux := http.NewServeMux()

	// Защищенные маршруты (регистрируем ТОЛЬКО один раз)
	// AuthCookieMiddleware проверяет session cookie
	// CSRFMiddleware проверяет X-CSRF-Token для опасных методов
	mux.Handle("POST /v1/tasks", middleware.AuthCookieMiddleware(
		middleware.CSRFMiddleware(
			http.HandlerFunc(tasksHandler.CreateTask),
		),
	))
	mux.Handle("GET /v1/tasks", middleware.AuthCookieMiddleware(
		http.HandlerFunc(tasksHandler.GetTasks),
	))
	mux.Handle("GET /v1/tasks/{id}", middleware.AuthCookieMiddleware(
		http.HandlerFunc(tasksHandler.GetTask),
	))
	mux.Handle("PATCH /v1/tasks/{id}", middleware.AuthCookieMiddleware(
		middleware.CSRFMiddleware(
			http.HandlerFunc(tasksHandler.UpdateTask),
		),
	))
	mux.Handle("DELETE /v1/tasks/{id}", middleware.AuthCookieMiddleware(
		middleware.CSRFMiddleware(
			http.HandlerFunc(tasksHandler.DeleteTask),
		),
	))

	// Поиск (без CSRF, так как GET)
	mux.Handle("GET /v1/tasks/search", middleware.AuthCookieMiddleware(
		http.HandlerFunc(tasksHandler.SearchTasks),
	))

	// ДЕМОНСТРАЦИЯ УЯЗВИМОСТИ (только для учебных целей)
	mux.Handle("GET /v1/tasks/search/unsafe", middleware.AuthCookieMiddleware(
		http.HandlerFunc(tasksHandler.SearchTasksUnsafe),
	))

	// Метрики (без авторизации)
	mux.Handle("GET /metrics", promhttp.Handler())

	// Цепочка глобальных middleware
	handler := sharedMiddleware.RequestIDMiddleware(
		sharedMiddleware.AccessLogMiddleware(log)(
			middleware.SecurityHeadersMiddleware(
				middleware.MetricsMiddleware(mux),
			),
		),
	)

	log.Info("HTTP server listening", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}

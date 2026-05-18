package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Метрики
var (
	// http_requests_total - счётчик запросов (Counter)
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	// http_request_duration_seconds - гистограмма длительности (Histogram)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"method", "route"},
	)

	// http_in_flight_requests - текущее число активных запросов (Gauge)
	inFlightRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_in_flight_requests",
			Help: "Current number of in-flight HTTP requests",
		},
	)
)

func init() {
	// Регистрируем метрики
	prometheus.MustRegister(requestsTotal, requestDuration, inFlightRequests)
}

// MetricsMiddleware собирает метрики по запросам
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Увеличиваем счётчик активных запросов
		inFlightRequests.Inc()
		defer inFlightRequests.Dec()

		// Засекаем время начала
		start := time.Now()

		// Оборачиваем ResponseWriter для перехвата статуса
		wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		// Обрабатываем запрос
		next.ServeHTTP(wrapped, r)

		// Получаем route (нормализованный путь)
		route := normalizeRoute(r.URL.Path)

		// Записываем метрики
		duration := time.Since(start).Seconds()
		requestDuration.WithLabelValues(r.Method, route).Observe(duration)
		requestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(wrapped.statusCode)).Inc()
	})
}

// statusRecorder для перехвата статус кода
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// normalizeRoute заменяет динамические ID на шаблон
func normalizeRoute(path string) string {
	// Пример: /v1/tasks/123 -> /v1/tasks/{id}
	// Для простоты оставляем как есть, но можно улучшить
	return path
}

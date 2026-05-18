package middleware

import "net/http"

// SecurityHeadersMiddleware добавляет заголовки безопасности
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Защита от MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Простая CSP (запрещает inline скрипты)
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Запрещаем iframe (кликджекинг)
		w.Header().Set("X-Frame-Options", "DENY")

		next.ServeHTTP(w, r)
	})
}

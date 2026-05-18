package middleware

import (
	"net/http"
)

// CSRFMiddleware проверяет X-CSRF-Token заголовок против csrf_token cookie
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем только "опасные" методы
		if r.Method == http.MethodPost || r.Method == http.MethodPatch ||
			r.Method == http.MethodPut || r.Method == http.MethodDelete {

			// Получаем CSRF токен из cookie
			csrfCookie, err := r.Cookie("csrf_token")
			if err != nil {
				http.Error(w, `{"error":"csrf token missing"}`, http.StatusForbidden)
				return
			}

			// Получаем CSRF токен из заголовка
			csrfHeader := r.Header.Get("X-CSRF-Token")
			if csrfHeader == "" {
				http.Error(w, `{"error":"csrf header missing"}`, http.StatusForbidden)
				return
			}

			// Сравниваем
			if csrfCookie.Value != csrfHeader {
				http.Error(w, `{"error":"csrf token mismatch"}`, http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

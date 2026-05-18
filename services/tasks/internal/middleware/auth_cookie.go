package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const UserSubjectKey contextKey = "subject"

// AuthCookieMiddleware проверяет session cookie
func AuthCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем наличие session cookie
		sessionCookie, err := r.Cookie("session_id")
		if err != nil || sessionCookie.Value == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// В учебной версии считаем любую непустую session_id валидной
		// В реальном приложении нужно проверять в БД/Redis

		// Кладём subject в контекст (можно вытащить из session)
		ctx := context.WithValue(r.Context(), UserSubjectKey, "student")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetSubject из контекста
func GetSubject(ctx context.Context) string {
	if val := ctx.Value(UserSubjectKey); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

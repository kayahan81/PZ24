package middleware

import (
	"context"
	"net/http"
	"strings"

	"tech-ip-sem2/services/tasks/internal/client/authgrpc"
	"tech-ip-sem2/shared/middleware"

	"go.uber.org/zap"
)

type AuthMiddleware struct {
	authClient *authgrpc.Client
	log        *zap.Logger
}

func NewAuthMiddleware(authClient *authgrpc.Client, log *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
		log:        log,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())
		log := m.log.With(zap.String("request_id", requestID), zap.String("component", "auth_middleware"))

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn("missing or invalid auth header")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		valid, subject, err := m.authClient.VerifyToken(r.Context(), authHeader)

		if err != nil {
			switch e := err.(type) {
			case *authgrpc.AuthError:
				switch e.Code {
				case "TIMEOUT", "UNAVAILABLE":
					log.Error("auth service unavailable", zap.String("reason", e.Code))
					http.Error(w, `{"error":"auth service unavailable"}`, http.StatusServiceUnavailable)
					return
				case "UNAUTHORIZED":
					log.Warn("invalid token")
					http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
			default:
				log.Error("unexpected error", zap.Error(err))
				http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
				return
			}
			return
		}

		if !valid {
			log.Warn("token validation failed")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		log.Info("token validated", zap.String("subject", subject))
		ctx := context.WithValue(r.Context(), "subject", subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

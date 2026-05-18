package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"tech-ip-sem2/shared/logger"
	"tech-ip-sem2/shared/middleware"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const validToken = "demo-token"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type VerifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AuthHandler struct {
	log *zap.Logger
}

func NewAuthHandler(log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		log: log,
	}
}

// Login выдаёт session cookie и csrf_token cookie + access_token для совместимости
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := logger.WithRequestID(h.log, requestID)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("invalid login request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Info("login attempt", zap.String("username", req.Username))

	// Генерируем session ID и CSRF токен
	sessionID := uuid.New().String()
	csrfToken := uuid.New().String()

	// 1. Session cookie (HttpOnly, Secure, SameSite=Lax)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,                 // JS не может прочитать
		Secure:   true,                 // только по HTTPS
		SameSite: http.SameSiteLaxMode, // Lax для обычной навигации
		MaxAge:   3600,                 // 1 час
	})

	// 2. CSRF token cookie (НЕ HttpOnly, Secure, SameSite=Lax)
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false, // JS может прочитать (для отправки в заголовке)
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	// 3. Access token для совместимости с существующими клиентами (Tasks gRPC)
	resp := LoginResponse{
		AccessToken: validToken,
		TokenType:   "Bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Info("login successful",
		zap.String("username", req.Username),
		zap.String("session_id", sessionID),
	)
}

// Verify — для совместимости с gRPC и HTTP клиентами (Bearer token)
func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := logger.WithRequestID(h.log, requestID)

	// Сначала проверяем session cookie (если есть)
	sessionCookie, err := r.Cookie("session_id")
	if err == nil && sessionCookie.Value != "" {
		// Аутентификация через cookie успешна
		log.Info("session verified via cookie", zap.String("session_id", sessionCookie.Value))
		resp := VerifyResponse{
			Valid:   true,
			Subject: "student",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Если нет cookie, проверяем Bearer token (для совместимости с gRPC) (большая уязвимость для учебного стенда)
	authHeader := r.Header.Get("Authorization")

	resp := VerifyResponse{}
	status := http.StatusOK

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		resp.Valid = false
		resp.Error = "unauthorized"
		status = http.StatusUnauthorized
		log.Warn("missing or invalid auth header")
	} else {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == validToken {
			resp.Valid = true
			resp.Subject = "student"
			log.Info("token verified", zap.String("subject", resp.Subject))
		} else {
			resp.Valid = false
			resp.Error = "unauthorized"
			status = http.StatusUnauthorized
			log.Warn("invalid token")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

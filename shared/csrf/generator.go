package csrf

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateToken создаёт случайный CSRF токен
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

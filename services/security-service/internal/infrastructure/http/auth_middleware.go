package http

import (
	"context"
	"net/http"
	"strings"

	"aurora/services/security-service/internal/infrastructure/security"
)

// ContextKey tipo para chaves de contexto
type ContextKey string

const (
	UserIDKey ContextKey = "userID"
	EmailKey  ContextKey = "email"
)

// AuthMiddleware middleware de autenticação JWT e opcionalmente X-Device-Key (regras/automação)
type AuthMiddleware struct {
	jwtValidator *security.JWTValidator
	deviceAPIKey string
}

// NewAuthMiddleware cria uma nova instância de AuthMiddleware
func NewAuthMiddleware(jwtValidator *security.JWTValidator, deviceAPIKey string) *AuthMiddleware {
	return &AuthMiddleware{jwtValidator: jwtValidator, deviceAPIKey: deviceAPIKey}
}

// Authenticate middleware: aceita JWT ou X-Device-Key (para chamadas do rules-service)
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Primeiro, aceita chamadas internas autenticadas via X-Device-Key
		if m.deviceAPIKey != "" && r.Header.Get("X-Device-Key") == m.deviceAPIKey {
			ctx := context.WithValue(r.Context(), UserIDKey, "*")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Fallback: exige JWT para chamadas externas (frontend)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtValidator.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

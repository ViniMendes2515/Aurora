package http

import (
	"context"
	"net/http"
	"strings"

	"aurora/services/sensors-service/internal/infrastructure/security"
)

// ContextKey tipo para chaves de contexto
type ContextKey string

const (
	// UserIDKey chave para o userID no contexto
	UserIDKey ContextKey = "userID"
	// EmailKey chave para o email no contexto
	EmailKey ContextKey = "email"
)

// AuthMiddleware middleware de autenticação JWT
type AuthMiddleware struct {
	jwtValidator *security.JWTValidator
}

// NewAuthMiddleware cria uma nova instância de AuthMiddleware
func NewAuthMiddleware(jwtValidator *security.JWTValidator) *AuthMiddleware {
	return &AuthMiddleware{
		jwtValidator: jwtValidator,
	}
}

// Authenticate middleware que valida o token JWT
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair token do header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Verificar formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validar token via jwt_validator
		claims, err := m.jwtValidator.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Adicionar userID e email ao contexto
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)

		// Continuar para o próximo handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

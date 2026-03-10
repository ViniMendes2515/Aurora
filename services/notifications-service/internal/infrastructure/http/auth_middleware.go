package http

import (
	pkgjwt "aurora/pkg/jwt"
	"aurora/services/notifications-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// NewAuthMiddleware retorna um gin.HandlerFunc que valida o token JWT.
func NewAuthMiddleware(jwtValidator *security.JWTValidator) gin.HandlerFunc {
	return pkgjwt.AuthMiddleware(jwtValidator)
}

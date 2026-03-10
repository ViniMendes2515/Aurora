package http

import (
	pkgjwt "aurora/pkg/jwt"
	"aurora/services/security-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// NewAuthMiddleware retorna um gin.HandlerFunc que aceita JWT ou X-Device-Key.
// X-Device-Key permite chamadas internas do rules-service sem JWT.
func NewAuthMiddleware(jwtValidator *security.JWTValidator, deviceAPIKey string) gin.HandlerFunc {
	return pkgjwt.AuthMiddlewareWithDeviceKey(jwtValidator, deviceAPIKey)
}

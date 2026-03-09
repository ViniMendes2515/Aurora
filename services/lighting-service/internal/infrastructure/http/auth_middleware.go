package http

import (
	"net/http"
	"strings"

	"aurora/services/lighting-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// NewAuthMiddleware retorna um gin.HandlerFunc que aceita JWT ou X-Device-Key.
// X-Device-Key permite chamadas internas do rules-service sem JWT.
func NewAuthMiddleware(jwtValidator *security.JWTValidator, deviceAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Autenticação por chave de dispositivo (chamadas internas do rules-service)
		if deviceAPIKey != "" && c.GetHeader("X-Device-Key") == deviceAPIKey {
			c.Set("userID", "*")
			c.Next()
			return
		}

		// Autenticação por JWT (chamadas do frontend)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := jwtValidator.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

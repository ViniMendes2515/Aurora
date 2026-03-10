package jwt

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	contextKeyUserID = "userID"
	contextKeyEmail  = "email"
)

// AuthMiddleware returns a Gin middleware that validates a Bearer JWT token.
// On success it sets "userID" and "email" in the context.
// Intended for services where only human users authenticate (sensors, rules).
func AuthMiddleware(validator *JWTValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := validateBearerToken(c, validator)
		if !ok {
			return
		}
		c.Set(contextKeyUserID, claims.UserID)
		c.Set(contextKeyEmail, claims.Email)
		c.Next()
	}
}

// AuthMiddlewareWithDeviceKey returns a Gin middleware that accepts either a
// valid Bearer JWT token or a matching X-Device-Key header.
// Intended for services where both humans (JWT) and ESP32 devices (device key) call in.
func AuthMiddlewareWithDeviceKey(validator *JWTValidator, deviceAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key := c.GetHeader("X-Device-Key"); key != "" {
			if key != deviceAPIKey {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid device key"})
				return
			}
			c.Set(contextKeyUserID, "*")
			c.Next()
			return
		}

		claims, ok := validateBearerToken(c, validator)
		if !ok {
			return
		}
		c.Set(contextKeyUserID, claims.UserID)
		c.Set(contextKeyEmail, claims.Email)
		c.Next()
	}
}

// validateBearerToken extracts and validates the Bearer token from the
// Authorization header. Returns (claims, true) on success, or aborts the
// request and returns (nil, false) on failure.
func validateBearerToken(c *gin.Context, validator *JWTValidator) (*TokenClaims, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
		return nil, false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		return nil, false
	}

	claims, err := validator.ValidateToken(parts[1])
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return nil, false
	}

	return claims, true
}

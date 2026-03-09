package http

import (
	"net/http"

	"aurora/services/rules-service/internal/application"
	"aurora/services/rules-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registra todas as rotas do rules-service
func RegisterRoutes(router *gin.Engine, rulesEngine *application.RulesEngine, jwtValidator *security.JWTValidator) {
	handlers := NewHandlers(rulesEngine)
	authMiddleware := NewAuthMiddleware(jwtValidator)

	// Grupo /rules protegido por JWT
	rules := router.Group("/rules")
	rules.Use(authMiddleware)
	{
		rules.GET("", handlers.ListRules)
		rules.POST("", handlers.CreateRule)
		rules.DELETE("/:id", handlers.DeleteRule)
		rules.PUT("/:id", handlers.UpdateRule)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "rules-service"})
	})
}

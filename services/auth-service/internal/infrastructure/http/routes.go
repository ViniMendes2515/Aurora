package http

import (
	"net/http"

	"aurora/services/auth-service/internal/application"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registra todas as rotas do auth-service
func RegisterRoutes(router *gin.Engine, authService *application.AuthService) {
	handlers := NewHandlers(authService)

	// Grupo /auth agrupa register e login sem middleware (endpoints públicos)
	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "auth-service"})
	})
}

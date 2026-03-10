package http

import (
	"github.com/gin-gonic/gin"

	"aurora/services/notifications-service/internal/application"
	"aurora/services/notifications-service/internal/infrastructure/security"
)

// RegisterRoutes registra todas as rotas do notifications-service no router Gin
func RegisterRoutes(router *gin.Engine, service *application.NotificationService, jwtValidator *security.JWTValidator) {
	handler := NewHandler(service)
	auth := NewAuthMiddleware(jwtValidator)

	router.GET("/health", handler.Health)

	protected := router.Group("/")
	protected.Use(auth)
	{
		protected.GET("/notifications", handler.GetNotifications)
		protected.PUT("/notifications/:id/read", handler.MarkAsRead)
	}
}

package http

import (
	"net/http"

	"aurora/services/security-service/internal/application"
	"aurora/services/security-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registra todas as rotas do security-service
func RegisterRoutes(router *gin.Engine, alarmService *application.AlarmService, jwtValidator *security.JWTValidator, deviceAPIKey string) {
	handlers := NewHandlers(alarmService)
	authMiddleware := NewAuthMiddleware(jwtValidator, deviceAPIKey)

	// Grupo /alarms protegido por JWT ou X-Device-Key
	alarms := router.Group("/alarms")
	alarms.Use(authMiddleware)
	{
		alarms.GET("", handlers.GetRecentAlarms)
		alarms.POST("/trigger", handlers.TriggerAlarm)
		alarms.POST("/:id/silence", handlers.SilenceAlarm)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "security-service"})
	})
}

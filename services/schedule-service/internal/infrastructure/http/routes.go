package http

import (
	"net/http"

	pkgjwt "aurora/pkg/jwt"
	"aurora/services/schedule-service/internal/application"
	"aurora/services/schedule-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registra todas as rotas do schedule-service
func RegisterRoutes(router *gin.Engine, service *application.ScheduleService, jwtValidator *security.JWTValidator) {
	handlers := NewHandlers(service)

	// Grupo /schedules protegido por JWT
	schedules := router.Group("/schedules")
	schedules.Use(pkgjwt.AuthMiddleware(jwtValidator))
	{
		schedules.POST("", handlers.CreateSchedule)
		schedules.GET("", handlers.ListSchedules)
		schedules.GET("/:id", handlers.GetSchedule)
		schedules.PUT("/:id", handlers.UpdateSchedule)
		schedules.PATCH("/:id/toggle", handlers.ToggleSchedule)
		schedules.DELETE("/:id", handlers.DeleteSchedule)
		schedules.GET("/:id/history", handlers.ListExecutions)
		schedules.POST("/preview", handlers.PreviewSchedule)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "schedule-service"})
	})
}

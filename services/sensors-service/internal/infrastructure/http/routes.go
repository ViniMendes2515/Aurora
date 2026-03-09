package http

import (
	"net/http"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/domain"
	"aurora/services/sensors-service/internal/infrastructure/security"
	"aurora/services/sensors-service/internal/infrastructure/ws"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registra todas as rotas do sensors-service
func RegisterRoutes(router *gin.Engine, motionService *application.MotionService, lightService *application.LightService, sensorRepo domain.SensorRepository, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub) {
	handlers := NewHandlers(motionService, lightService, sensorRepo, deviceAPIKey)
	authMiddleware := NewAuthMiddleware(jwtValidator)

	// Endpoints para dispositivos físicos (autenticação via X-Device-Key no handler)
	device := router.Group("/sensors/device")
	{
		device.POST("/:id/motion", handlers.RegisterDeviceMotion)
		device.POST("/:id/light", handlers.RegisterDeviceLight)
	}

	// WebSocket para stream em tempo real (sem JWT — o frontend conecta diretamente)
	// gin.WrapF adapta func(http.ResponseWriter, *http.Request) para gin.HandlerFunc
	router.GET("/ws", gin.WrapF(hub.ServeWS))

	// Endpoints protegidos por JWT
	sensors := router.Group("/sensors")
	sensors.Use(authMiddleware)
	{
		sensors.GET("", handlers.ListSensors)
		sensors.POST("/:id/motion", handlers.RegisterMotion)
		sensors.GET("/:id/motion", handlers.GetMotionHistory)
		sensors.GET("/:id/light", handlers.GetLightHistory)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "sensors-service"})
	})
}

package http

import (
	"net/http"

	"aurora/services/lighting-service/internal/application"
	"aurora/services/lighting-service/internal/infrastructure/device"
	"aurora/services/lighting-service/internal/infrastructure/security"
	"aurora/services/lighting-service/internal/infrastructure/ws"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registra todas as rotas do lighting-service
func RegisterRoutes(router *gin.Engine, lightService *application.LightService, esp32Client *device.ESP32Client, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub) {
	handlers := NewHandlers(lightService, esp32Client, deviceAPIKey)
	authMiddleware := NewAuthMiddleware(jwtValidator, deviceAPIKey)

	// Registro de dispositivo ESP32 (autenticação via X-Device-Key no handler)
	router.POST("/devices/register", handlers.RegisterDevice)

	// WebSocket para atualizações em tempo real do estado das luzes
	router.GET("/ws", gin.WrapF(hub.ServeWS))

	// Endpoints de luzes (JWT ou X-Device-Key para o rules-service)
	lights := router.Group("/lights")
	lights.Use(authMiddleware)
	{
		lights.GET("", handlers.ListLights)
		lights.POST("/:id/on", handlers.TurnOn)
		lights.POST("/:id/off", handlers.TurnOff)
		lights.GET("/:id/status", handlers.GetLightStatus)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "lighting-service"})
	})
}

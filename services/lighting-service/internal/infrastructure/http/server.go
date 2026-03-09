package http

import (
	"log"

	"aurora/services/lighting-service/internal/application"
	"aurora/services/lighting-service/internal/infrastructure/device"
	"aurora/services/lighting-service/internal/infrastructure/security"
	"aurora/services/lighting-service/internal/infrastructure/ws"

	"github.com/gin-gonic/gin"
)

// Server representa o servidor HTTP
type Server struct {
	lightService *application.LightService
	esp32Client  *device.ESP32Client
	jwtValidator *security.JWTValidator
	deviceAPIKey string
	hub          *ws.Hub
	port         string
}

// NewServer cria uma nova instância do servidor
func NewServer(lightService *application.LightService, esp32Client *device.ESP32Client, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub, port string) *Server {
	return &Server{
		lightService: lightService,
		esp32Client:  esp32Client,
		jwtValidator: jwtValidator,
		deviceAPIKey: deviceAPIKey,
		hub:          hub,
		port:         port,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.lightService, s.esp32Client, s.jwtValidator, s.deviceAPIKey, s.hub)

	log.Printf("Lighting Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

package http

import (
	"log"

	"aurora/services/lighting-service/internal/application"
	"aurora/services/lighting-service/internal/infrastructure/device"
	"aurora/services/lighting-service/internal/infrastructure/security"
	"aurora/services/lighting-service/internal/infrastructure/ws"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server representa o servidor HTTP
type Server struct {
	lightService *application.LightService
	esp32Client  *device.ESP32Client
	jwtValidator *security.JWTValidator
	deviceAPIKey string
	hub          *ws.Hub
	port         string
	debug        bool
}

// NewServer cria uma nova instância do servidor
func NewServer(lightService *application.LightService, esp32Client *device.ESP32Client, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub, port string, debug bool) *Server {
	return &Server{
		lightService: lightService,
		esp32Client:  esp32Client,
		jwtValidator: jwtValidator,
		deviceAPIKey: deviceAPIKey,
		hub:          hub,
		port:         port,
		debug:        debug,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.lightService, s.esp32Client, s.jwtValidator, s.deviceAPIKey, s.hub)

	if s.debug {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Printf("Swagger UI disponível em http://localhost:%s/swagger/index.html", s.port)
	}

	log.Printf("Lighting Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

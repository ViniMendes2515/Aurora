package http

import (
	"log"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/domain"
	"aurora/services/sensors-service/internal/infrastructure/security"
	"aurora/services/sensors-service/internal/infrastructure/ws"

	"github.com/gin-gonic/gin"
)

// Server representa o servidor HTTP
type Server struct {
	motionService *application.MotionService
	lightService  *application.LightService
	sensorRepo    domain.SensorRepository
	jwtValidator  *security.JWTValidator
	deviceAPIKey  string
	hub           *ws.Hub
	port          string
}

// NewServer cria uma nova instância do servidor
func NewServer(motionService *application.MotionService, lightService *application.LightService, sensorRepo domain.SensorRepository, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub, port string) *Server {
	return &Server{
		motionService: motionService,
		lightService:  lightService,
		sensorRepo:    sensorRepo,
		jwtValidator:  jwtValidator,
		deviceAPIKey:  deviceAPIKey,
		hub:           hub,
		port:          port,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.motionService, s.lightService, s.sensorRepo, s.jwtValidator, s.deviceAPIKey, s.hub)

	log.Printf("Sensors Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

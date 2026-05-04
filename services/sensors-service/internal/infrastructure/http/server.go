package http

import (
	"log"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/domain"
	"aurora/services/sensors-service/internal/infrastructure/security"
	"aurora/services/sensors-service/internal/infrastructure/ws"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	debug         bool
}

// NewServer cria uma nova instância do servidor
func NewServer(motionService *application.MotionService, lightService *application.LightService, sensorRepo domain.SensorRepository, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub, port string, debug bool) *Server {
	return &Server{
		motionService: motionService,
		lightService:  lightService,
		sensorRepo:    sensorRepo,
		jwtValidator:  jwtValidator,
		deviceAPIKey:  deviceAPIKey,
		hub:           hub,
		port:          port,
		debug:         debug,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.motionService, s.lightService, s.sensorRepo, s.jwtValidator, s.deviceAPIKey, s.hub)

	if s.debug {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Printf("Swagger UI disponível em http://localhost:%s/swagger/index.html", s.port)
	}

	log.Printf("Sensors Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

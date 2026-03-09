package http

import (
	"log"

	"aurora/services/security-service/internal/application"
	"aurora/services/security-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// Server representa o servidor HTTP do security-service
type Server struct {
	alarmService *application.AlarmService
	jwtValidator *security.JWTValidator
	deviceAPIKey string
	port         string
}

// NewServer cria uma nova instância do servidor
func NewServer(alarmService *application.AlarmService, jwtValidator *security.JWTValidator, deviceAPIKey string, port string) *Server {
	return &Server{
		alarmService: alarmService,
		jwtValidator: jwtValidator,
		deviceAPIKey: deviceAPIKey,
		port:         port,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.alarmService, s.jwtValidator, s.deviceAPIKey)

	log.Printf("Security Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

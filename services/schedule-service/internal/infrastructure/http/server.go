package http

import (
	"log"

	"aurora/services/schedule-service/internal/application"
	"aurora/services/schedule-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// Server representa o servidor HTTP do schedule-service
type Server struct {
	service      *application.ScheduleService
	jwtValidator *security.JWTValidator
	port         string
}

// NewServer cria uma nova instância do servidor
func NewServer(service *application.ScheduleService, jwtValidator *security.JWTValidator, port string) *Server {
	return &Server{
		service:      service,
		jwtValidator: jwtValidator,
		port:         port,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.service, s.jwtValidator)

	log.Printf("Schedule Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

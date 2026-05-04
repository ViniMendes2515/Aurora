package http

import (
	"log"

	"aurora/services/schedule-service/internal/application"
	"aurora/services/schedule-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server representa o servidor HTTP do schedule-service
type Server struct {
	service      *application.ScheduleService
	jwtValidator *security.JWTValidator
	port         string
	debug        bool
}

// NewServer cria uma nova instância do servidor
func NewServer(service *application.ScheduleService, jwtValidator *security.JWTValidator, port string, debug bool) *Server {
	return &Server{
		service:      service,
		jwtValidator: jwtValidator,
		port:         port,
		debug:        debug,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.service, s.jwtValidator)

	if s.debug {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Printf("Swagger UI disponível em http://localhost:%s/swagger/index.html", s.port)
	}

	log.Printf("Schedule Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

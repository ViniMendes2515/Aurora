package http

import (
	"github.com/gin-gonic/gin"

	"aurora/services/notifications-service/internal/application"
	"aurora/services/notifications-service/internal/infrastructure/security"
)

// Server encapsula o servidor HTTP do notifications-service
type Server struct {
	router       *gin.Engine
	port         string
}

// NewServer cria e configura um novo servidor HTTP
func NewServer(service *application.NotificationService, jwtValidator *security.JWTValidator, port string) *Server {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, service, jwtValidator)

	return &Server{router: router, port: port}
}

// Start inicia o servidor HTTP na porta configurada
func (s *Server) Start() error {
	return s.router.Run(":" + s.port)
}

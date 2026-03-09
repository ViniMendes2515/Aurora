package http

import (
	"log"

	"aurora/services/auth-service/internal/application"

	"github.com/gin-gonic/gin"
)

// Server representa o servidor HTTP
type Server struct {
	authService *application.AuthService
	port        string
}

// NewServer cria uma nova instância do servidor
func NewServer(authService *application.AuthService, port string) *Server {
	return &Server{
		authService: authService,
		port:        port,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.authService)

	log.Printf("Auth Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

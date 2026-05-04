package http

import (
	"log"

	"aurora/services/auth-service/internal/application"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
)

// Server representa o servidor HTTP
type Server struct {
	authService *application.AuthService
	port        string
	debug       bool
}

// NewServer cria uma nova instância do servidor
func NewServer(authService *application.AuthService, port string, debug bool) *Server {
	return &Server{
		authService: authService,
		port:        port,
		debug:       debug,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.authService)

	if s.debug {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Printf("Swagger UI disponível em http://localhost:%s/swagger/index.html", s.port)
	}

	log.Printf("Auth Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

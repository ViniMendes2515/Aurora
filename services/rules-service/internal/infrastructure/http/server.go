package http

import (
	"log"

	"aurora/services/rules-service/internal/application"
	"aurora/services/rules-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// Server representa o servidor HTTP do rules-service
type Server struct {
	rulesEngine  *application.RulesEngine
	jwtValidator *security.JWTValidator
	port         string
}

// NewServer cria uma nova instância do servidor
func NewServer(rulesEngine *application.RulesEngine, jwtValidator *security.JWTValidator, port string) *Server {
	return &Server{
		rulesEngine:  rulesEngine,
		jwtValidator: jwtValidator,
		port:         port,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.rulesEngine, s.jwtValidator)

	log.Printf("Rules Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

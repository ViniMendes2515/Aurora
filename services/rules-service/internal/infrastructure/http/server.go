package http

import (
	"log"

	"aurora/services/rules-service/internal/application"
	"aurora/services/rules-service/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server representa o servidor HTTP do rules-service
type Server struct {
	rulesEngine  *application.RulesEngine
	jwtValidator *security.JWTValidator
	port         string
	debug        bool
}

// NewServer cria uma nova instância do servidor
func NewServer(rulesEngine *application.RulesEngine, jwtValidator *security.JWTValidator, port string, debug bool) *Server {
	return &Server{
		rulesEngine:  rulesEngine,
		jwtValidator: jwtValidator,
		port:         port,
		debug:        debug,
	}
}

// Start inicia o servidor HTTP com Gin
func (s *Server) Start() error {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, s.rulesEngine, s.jwtValidator)

	if s.debug {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Printf("Swagger UI disponível em http://localhost:%s/swagger/index.html", s.port)
	}

	log.Printf("Rules Service listening on :%s", s.port)
	return router.Run(":" + s.port)
}

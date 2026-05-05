package http

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"aurora/services/notifications-service/internal/application"
	"aurora/services/notifications-service/internal/infrastructure/security"
)

// Server encapsula o servidor HTTP do notifications-service
type Server struct {
	router *gin.Engine
	port   string
	debug  bool
}

// NewServer cria e configura um novo servidor HTTP
func NewServer(service *application.NotificationService, telegramService *application.TelegramService, jwtValidator *security.JWTValidator, port string, debug bool) *Server {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	RegisterRoutes(router, service, telegramService, jwtValidator)

	if debug {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Printf("Swagger UI disponível em http://localhost:%s/swagger/index.html", port)
	}

	return &Server{router: router, port: port, debug: debug}
}

// Start inicia o servidor HTTP na porta configurada
func (s *Server) Start() error {
	return s.router.Run(":" + s.port)
}

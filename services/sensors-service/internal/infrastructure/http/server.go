package http

import (
	"log"
	"net/http"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/infrastructure/security"
)

// Server representa o servidor HTTP
type Server struct {
	motionService *application.MotionService
	jwtValidator  *security.JWTValidator
	port          string
}

// NewServer cria uma nova instância do servidor
func NewServer(motionService *application.MotionService, jwtValidator *security.JWTValidator, port string) *Server {
	return &Server{
		motionService: motionService,
		jwtValidator:  jwtValidator,
		port:          port,
	}
}

// Start inicia o servidor HTTP
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Registrar rotas
	RegisterRoutes(mux, s.motionService, s.jwtValidator)

	log.Printf("Server listening on :%s", s.port)
	return http.ListenAndServe(":"+s.port, mux)
}

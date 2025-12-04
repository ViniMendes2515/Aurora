package http

import (
	"log"
	"net/http"

	"aurora/services/auth-service/internal/application"
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

// Start inicia o servidor HTTP
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Registrar rotas
	RegisterRoutes(mux, s.authService)

	log.Printf("Server listening on :%s", s.port)
	return http.ListenAndServe(":"+s.port, mux)
}

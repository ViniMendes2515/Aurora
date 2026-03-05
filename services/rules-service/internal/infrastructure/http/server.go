package http

import (
	"log"
	"net/http"

	"aurora/services/rules-service/internal/application"
	"aurora/services/rules-service/internal/infrastructure/security"
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

// Start inicia o servidor HTTP
func (s *Server) Start() error {
	mux := http.NewServeMux()
	RegisterRoutes(mux, s.rulesEngine, s.jwtValidator)
	log.Printf("Rules Service listening on :%s", s.port)
	return http.ListenAndServe(":"+s.port, mux)
}

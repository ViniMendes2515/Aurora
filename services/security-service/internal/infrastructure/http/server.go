package http

import (
	"log"
	"net/http"

	"aurora/services/security-service/internal/application"
	"aurora/services/security-service/internal/infrastructure/security"
)

// Server representa o servidor HTTP do security-service
type Server struct {
	alarmService *application.AlarmService
	jwtValidator *security.JWTValidator
	deviceAPIKey string
	port         string
}

// NewServer cria uma nova instância do servidor
func NewServer(alarmService *application.AlarmService, jwtValidator *security.JWTValidator, deviceAPIKey string, port string) *Server {
	return &Server{
		alarmService: alarmService,
		jwtValidator: jwtValidator,
		deviceAPIKey: deviceAPIKey,
		port:         port,
	}
}

// Start inicia o servidor HTTP
func (s *Server) Start() error {
	mux := http.NewServeMux()
	RegisterRoutes(mux, s.alarmService, s.jwtValidator, s.deviceAPIKey)
	log.Printf("Security Service listening on :%s", s.port)
	return http.ListenAndServe(":"+s.port, mux)
}

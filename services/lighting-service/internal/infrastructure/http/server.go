package http

import (
	"log"
	"net/http"

	"aurora/services/lighting-service/internal/application"
	"aurora/services/lighting-service/internal/infrastructure/device"
	"aurora/services/lighting-service/internal/infrastructure/security"
	"aurora/services/lighting-service/internal/infrastructure/ws"
)

// Server representa o servidor HTTP
type Server struct {
	lightService *application.LightService
	esp32Client  *device.ESP32Client
	jwtValidator *security.JWTValidator
	deviceAPIKey string
	hub          *ws.Hub
	port         string
}

// NewServer cria uma nova instância do servidor
func NewServer(lightService *application.LightService, esp32Client *device.ESP32Client, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub, port string) *Server {
	return &Server{
		lightService: lightService,
		esp32Client:  esp32Client,
		jwtValidator: jwtValidator,
		deviceAPIKey: deviceAPIKey,
		hub:          hub,
		port:         port,
	}
}

// Start inicia o servidor HTTP
func (s *Server) Start() error {
	mux := http.NewServeMux()
	RegisterRoutes(mux, s.lightService, s.esp32Client, s.jwtValidator, s.deviceAPIKey, s.hub)
	log.Printf("Lighting Service listening on :%s", s.port)
	return http.ListenAndServe(":"+s.port, mux)
}

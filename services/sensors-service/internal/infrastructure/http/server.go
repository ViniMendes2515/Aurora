package http

import (
	"log"
	"net/http"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/domain"
	"aurora/services/sensors-service/internal/infrastructure/security"
	"aurora/services/sensors-service/internal/infrastructure/ws"
)

// Server representa o servidor HTTP
type Server struct {
	motionService *application.MotionService
	lightService  *application.LightService
	sensorRepo    domain.SensorRepository
	jwtValidator  *security.JWTValidator
	deviceAPIKey  string
	hub           *ws.Hub
	port          string
}

// NewServer cria uma nova instância do servidor
func NewServer(motionService *application.MotionService, lightService *application.LightService, sensorRepo domain.SensorRepository, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub, port string) *Server {
	return &Server{
		motionService: motionService,
		lightService:  lightService,
		sensorRepo:    sensorRepo,
		jwtValidator:  jwtValidator,
		deviceAPIKey:  deviceAPIKey,
		hub:           hub,
		port:          port,
	}
}

// Start inicia o servidor HTTP
func (s *Server) Start() error {
	mux := http.NewServeMux()

	RegisterRoutes(mux, s.motionService, s.lightService, s.sensorRepo, s.jwtValidator, s.deviceAPIKey, s.hub)

	log.Printf("Sensors Service listening on :%s", s.port)
	return http.ListenAndServe(":"+s.port, mux)
}

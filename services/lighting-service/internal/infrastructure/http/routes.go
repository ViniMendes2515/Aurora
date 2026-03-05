package http

import (
	"net/http"
	"strings"

	"aurora/services/lighting-service/internal/application"
	"aurora/services/lighting-service/internal/infrastructure/device"
	"aurora/services/lighting-service/internal/infrastructure/security"
	"aurora/services/lighting-service/internal/infrastructure/ws"
)

// RegisterRoutes registra todas as rotas do lighting-service
func RegisterRoutes(mux *http.ServeMux, lightService *application.LightService, esp32Client *device.ESP32Client, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub) {
	handlers := NewHandlers(lightService, esp32Client, deviceAPIKey)
	authMiddleware := NewAuthMiddleware(jwtValidator, deviceAPIKey)

	// Registro de dispositivo (sem JWT — usa API key)
	mux.HandleFunc("/devices/register", handlers.RegisterDevice)

	// WebSocket — atualizações em tempo real de estado das luzes
	mux.HandleFunc("/ws", hub.ServeWS)

	// Rotas de luzes (JWT obrigatório)
	mux.Handle("/lights", authMiddleware.Authenticate(http.HandlerFunc(handlers.ListLights)))
	mux.Handle("/lights/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/on") && r.Method == http.MethodPost:
			handlers.TurnOn(w, r)
		case strings.HasSuffix(path, "/off") && r.Method == http.MethodPost:
			handlers.TurnOff(w, r)
		case strings.HasSuffix(path, "/status") && r.Method == http.MethodGet:
			handlers.GetLightStatus(w, r)
		default:
			http.NotFound(w, r)
		}
	})))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","service":"lighting-service"}`))
	})
}

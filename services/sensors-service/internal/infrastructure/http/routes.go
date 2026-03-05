package http

import (
	"net/http"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/domain"
	"aurora/services/sensors-service/internal/infrastructure/security"
	"aurora/services/sensors-service/internal/infrastructure/ws"
)

// RegisterRoutes registra todas as rotas do sensors-service
func RegisterRoutes(mux *http.ServeMux, motionService *application.MotionService, lightService *application.LightService, sensorRepo domain.SensorRepository, jwtValidator *security.JWTValidator, deviceAPIKey string, hub *ws.Hub) {
	handlers := NewHandlers(motionService, lightService, sensorRepo, deviceAPIKey)
	authMiddleware := NewAuthMiddleware(jwtValidator)

	// Rotas de dispositivos (API Key — sem JWT)
	mux.HandleFunc("/sensors/device/", func(w http.ResponseWriter, r *http.Request) {
		if containsPath(r.URL.Path, "/motion") {
			handlers.RegisterDeviceMotion(w, r)
		} else if containsPath(r.URL.Path, "/light") {
			handlers.RegisterDeviceLight(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// WebSocket para stream em tempo real de eventos de sensores
	mux.HandleFunc("/ws", hub.ServeWS)

	// Rotas protegidas por JWT
	mux.Handle("/sensors", authMiddleware.Authenticate(http.HandlerFunc(handlers.ListSensors)))
	mux.Handle("/sensors/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if containsPath(r.URL.Path, "/motion") && r.Method == http.MethodPost {
			handlers.RegisterMotion(w, r)
		} else if containsPath(r.URL.Path, "/motion") && r.Method == http.MethodGet {
			handlers.GetMotionHistory(w, r)
		} else if containsPath(r.URL.Path, "/light") && r.Method == http.MethodGet {
			handlers.GetLightHistory(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))

	// Health check
	mux.HandleFunc("/health", healthCheck)
}

func containsPath(path, segment string) bool {
	return len(path) > len(segment) && path[len(path)-len(segment):] == segment ||
		contains(path, segment+"/") || contains(path, "/"+segment)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"sensors-service"}`))
}

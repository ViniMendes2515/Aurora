package http

import (
	"net/http"
	"strings"

	"aurora/services/security-service/internal/application"
	"aurora/services/security-service/internal/infrastructure/security"
)

// RegisterRoutes registra todas as rotas do security-service
func RegisterRoutes(mux *http.ServeMux, alarmService *application.AlarmService, jwtValidator *security.JWTValidator, deviceAPIKey string) {
	handlers := NewHandlers(alarmService)
	authMiddleware := NewAuthMiddleware(jwtValidator, deviceAPIKey)

	// POST /alarms/trigger — disparo manual (registrado antes do subtree /alarms/)
	mux.Handle("/alarms/trigger", authMiddleware.Authenticate(http.HandlerFunc(handlers.TriggerAlarm)))

	// POST /alarms/{id}/silence
	mux.Handle("/alarms/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/silence") {
			handlers.SilenceAlarm(w, r)
			return
		}
		http.NotFound(w, r)
	})))

	// GET /alarms
	mux.Handle("/alarms", authMiddleware.Authenticate(http.HandlerFunc(handlers.GetRecentAlarms)))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","service":"security-service"}`))
	})
}

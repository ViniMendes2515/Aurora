package http

import (
	"net/http"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/infrastructure/security"
)

// RegisterRoutes registra todas as rotas do sensors-service
func RegisterRoutes(mux *http.ServeMux, motionService *application.MotionService, jwtValidator *security.JWTValidator) {
	handlers := NewHandlers(motionService)
	authMiddleware := NewAuthMiddleware(jwtValidator)

	// Rota protegida para sensores: POST /sensors/{id}/motion
	mux.Handle("/sensors/", authMiddleware.Authenticate(http.HandlerFunc(handlers.RegisterMotion)))

	// Health check
	mux.HandleFunc("/health", healthCheck)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"sensors-service"}`))
}

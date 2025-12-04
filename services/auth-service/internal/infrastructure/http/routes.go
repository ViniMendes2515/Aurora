package http

import (
	"net/http"

	"aurora/services/auth-service/internal/application"
)

// RegisterRoutes registra todas as rotas do auth-service
func RegisterRoutes(mux *http.ServeMux, authService *application.AuthService) {
	handlers := NewHandlers(authService)

	// Auth routes
	mux.HandleFunc("/auth/register", handlers.Register)
	mux.HandleFunc("/auth/login", handlers.Login)

	// Health check
	mux.HandleFunc("/health", healthCheck)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"auth-service"}`))
}

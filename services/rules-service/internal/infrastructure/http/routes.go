package http

import (
	"net/http"

	"aurora/services/rules-service/internal/application"
	"aurora/services/rules-service/internal/infrastructure/security"
)

// RegisterRoutes registra todas as rotas do rules-service
func RegisterRoutes(mux *http.ServeMux, rulesEngine *application.RulesEngine, jwtValidator *security.JWTValidator) {
	handlers := NewHandlers(rulesEngine)
	authMiddleware := NewAuthMiddleware(jwtValidator)

	mux.Handle("/rules", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ListRules(w, r)
		case http.MethodPost:
			handlers.CreateRule(w, r)
		default:
			http.NotFound(w, r)
		}
	})))

	mux.Handle("/rules/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			handlers.DeleteRule(w, r)
		case http.MethodPut:
			handlers.UpdateRule(w, r)
		default:
			http.NotFound(w, r)
		}
	})))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","service":"rules-service"}`))
	})
}

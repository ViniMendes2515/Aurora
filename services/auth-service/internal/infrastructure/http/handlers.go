package http

import (
	"encoding/json"
	"net/http"

	"aurora/services/auth-service/internal/application"
	"aurora/services/auth-service/internal/domain"
)

// Handlers contém os handlers HTTP
type Handlers struct {
	authService *application.AuthService
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(authService *application.AuthService) *Handlers {
	return &Handlers{
		authService: authService,
	}
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Error string `json:"error"`
}

// Register handler para POST /auth/register
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req application.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.authService.Register(req)
	if err != nil {
		switch err {
		case domain.ErrInvalidEmail:
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		case domain.ErrInvalidPassword:
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		case domain.ErrUserAlreadyExists:
			h.respondWithError(w, http.StatusConflict, err.Error())
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusCreated, response)
}

// Login handler para POST /auth/login
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req application.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			h.respondWithError(w, http.StatusUnauthorized, err.Error())
		case domain.ErrTokenGeneration:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *Handlers) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}

package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/domain"
)

// Handlers contém os handlers HTTP
type Handlers struct {
	motionService *application.MotionService
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(motionService *application.MotionService) *Handlers {
	return &Handlers{
		motionService: motionService,
	}
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Error string `json:"error"`
}

// RegisterMotion handler para POST /sensors/{id}/motion
func (h *Handlers) RegisterMotion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extrair sensor ID da URL: /sensors/{id}/motion
	sensorID := extractSensorID(r.URL.Path)
	if sensorID == "" {
		h.respondWithError(w, http.StatusBadRequest, "sensor ID is required")
		return
	}

	// Extrair userID do contexto (adicionado pelo middleware de autenticação)
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Chamar caso de uso (application layer)
	req := application.RegisterMotionRequest{
		SensorID: sensorID,
		UserID:   userID,
	}

	response, err := h.motionService.RegisterMotion(req)
	if err != nil {
		switch err {
		case domain.ErrSensorNotFound:
			h.respondWithError(w, http.StatusNotFound, err.Error())
		case domain.ErrSensorAccessDenied:
			h.respondWithError(w, http.StatusForbidden, err.Error())
		case domain.ErrInvalidSensorID:
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// extractSensorID extrai o ID do sensor do path /sensors/{id}/motion
func extractSensorID(path string) string {
	path = strings.TrimPrefix(path, "/sensors/")
	parts := strings.Split(path, "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

func (h *Handlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *Handlers) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}

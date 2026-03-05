package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"aurora/services/lighting-service/internal/application"
	"aurora/services/lighting-service/internal/domain"
	"aurora/services/lighting-service/internal/infrastructure/device"
)

// Handlers contém os handlers HTTP do lighting-service
type Handlers struct {
	lightService *application.LightService
	esp32Client  *device.ESP32Client
	deviceAPIKey string
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(lightService *application.LightService, esp32Client *device.ESP32Client, deviceAPIKey string) *Handlers {
	return &Handlers{
		lightService: lightService,
		esp32Client:  esp32Client,
		deviceAPIKey: deviceAPIKey,
	}
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Error string `json:"error"`
}

// ListLights handler para GET /lights
func (h *Handlers) ListLights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := r.Context().Value(UserIDKey).(string)
	lights, err := h.lightService.ListLights(userID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "failed to fetch lights")
		return
	}

	h.respondWithJSON(w, http.StatusOK, lights)
}

// TurnOn handler para POST /lights/{id}/on
func (h *Handlers) TurnOn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	lightID := extractLightID(r.URL.Path)
	if lightID == "" {
		h.respondWithError(w, http.StatusBadRequest, "light ID is required")
		return
	}

	userID := r.Context().Value(UserIDKey).(string)
	response, err := h.lightService.TurnOn(lightID, userID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// TurnOff handler para POST /lights/{id}/off
func (h *Handlers) TurnOff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	lightID := extractLightID(r.URL.Path)
	if lightID == "" {
		h.respondWithError(w, http.StatusBadRequest, "light ID is required")
		return
	}

	userID := r.Context().Value(UserIDKey).(string)
	response, err := h.lightService.TurnOff(lightID, userID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetLightStatus handler para GET /lights/{id}/status
func (h *Handlers) GetLightStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	lightID := extractLightID(r.URL.Path)
	if lightID == "" {
		h.respondWithError(w, http.StatusBadRequest, "light ID is required")
		return
	}

	userID := r.Context().Value(UserIDKey).(string)
	response, err := h.lightService.GetStatus(lightID, userID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// RegisterDevice handler para POST /devices/register (ESP32 registra seu IP)
func (h *Handlers) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.Header.Get("X-Device-Key") != h.deviceAPIKey {
		h.respondWithError(w, http.StatusUnauthorized, "invalid device key")
		return
	}

	var body struct {
		DeviceID string `json:"device_id"`
		IP       string `json:"ip"`
		Port     int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Port == 0 {
		body.Port = 80
	}

	dev := &domain.Device{
		ID:        body.DeviceID,
		IPAddress: body.IP,
		Port:      body.Port,
	}
	h.esp32Client.Register(dev)

	h.respondWithJSON(w, http.StatusOK, map[string]string{
		"status":    "registered",
		"device_id": body.DeviceID,
		"ip":        body.IP,
	})
}

func extractLightID(path string) string {
	path = strings.TrimPrefix(path, "/lights/")
	parts := strings.Split(path, "/")
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func (h *Handlers) handleDomainError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrLightNotFound:
		h.respondWithError(w, http.StatusNotFound, err.Error())
	case domain.ErrLightAccessDenied:
		h.respondWithError(w, http.StatusForbidden, err.Error())
	case domain.ErrInvalidLightID:
		h.respondWithError(w, http.StatusBadRequest, err.Error())
	case domain.ErrDeviceUnreachable:
		h.respondWithError(w, http.StatusServiceUnavailable, err.Error())
	default:
		h.respondWithError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *Handlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *Handlers) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}

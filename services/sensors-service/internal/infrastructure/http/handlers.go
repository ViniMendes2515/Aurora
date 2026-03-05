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
	lightService  *application.LightService
	sensorRepo    domain.SensorRepository
	deviceAPIKey  string
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(motionService *application.MotionService, lightService *application.LightService, sensorRepo domain.SensorRepository, deviceAPIKey string) *Handlers {
	return &Handlers{
		motionService: motionService,
		lightService:  lightService,
		sensorRepo:    sensorRepo,
		deviceAPIKey:  deviceAPIKey,
	}
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Error string `json:"error"`
}

// ---------------------------------------------------------------
// Endpoints para usuários autenticados (JWT)
// ---------------------------------------------------------------

// RegisterMotion handler para POST /sensors/{id}/motion
func (h *Handlers) RegisterMotion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sensorID := extractSensorID(r.URL.Path)
	if sensorID == "" {
		h.respondWithError(w, http.StatusBadRequest, "sensor ID is required")
		return
	}

	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	req := application.RegisterMotionRequest{
		SensorID: sensorID,
		UserID:   userID,
	}

	response, err := h.motionService.RegisterMotion(req)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// ListSensors handler para GET /sensors
func (h *Handlers) ListSensors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	sensors, err := h.sensorRepo.FindByOwnerID(userID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "failed to fetch sensors")
		return
	}

	h.respondWithJSON(w, http.StatusOK, sensors)
}

// GetMotionHistory handler para GET /sensors/{id}/motion
func (h *Handlers) GetMotionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sensorID := extractSensorID(r.URL.Path)
	if sensorID == "" {
		h.respondWithError(w, http.StatusBadRequest, "sensor ID is required")
		return
	}

	records, err := h.sensorRepo.GetMotionRecords(sensorID, 20)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "failed to fetch motion records")
		return
	}

	h.respondWithJSON(w, http.StatusOK, records)
}

// GetLightHistory handler para GET /sensors/{id}/light
func (h *Handlers) GetLightHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sensorID := extractSensorID(r.URL.Path)
	if sensorID == "" {
		h.respondWithError(w, http.StatusBadRequest, "sensor ID is required")
		return
	}

	records, err := h.lightService.GetLightRecords(sensorID, 20)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "failed to fetch light records")
		return
	}

	h.respondWithJSON(w, http.StatusOK, records)
}

// ---------------------------------------------------------------
// Endpoints para dispositivos (API Key — sem JWT)
// ---------------------------------------------------------------

// RegisterDeviceMotion handler para POST /sensors/device/{id}/motion
func (h *Handlers) RegisterDeviceMotion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if !h.validateDeviceKey(r) {
		h.respondWithError(w, http.StatusUnauthorized, "invalid device key")
		return
	}

	sensorID := extractDeviceSensorID(r.URL.Path)
	if sensorID == "" {
		h.respondWithError(w, http.StatusBadRequest, "sensor ID is required")
		return
	}

	req := application.RegisterMotionRequest{
		SensorID: sensorID,
		UserID:   "device",
	}

	response, err := h.motionService.RegisterMotion(req)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// RegisterDeviceLight handler para POST /sensors/device/{id}/light
func (h *Handlers) RegisterDeviceLight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if !h.validateDeviceKey(r) {
		h.respondWithError(w, http.StatusUnauthorized, "invalid device key")
		return
	}

	sensorID := extractDeviceSensorID(r.URL.Path)
	if sensorID == "" {
		h.respondWithError(w, http.StatusBadRequest, "sensor ID is required")
		return
	}

	var body struct {
		Value float64 `json:"value"`
		Raw   int     `json:"raw"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req := application.RegisterLightLevelRequest{
		SensorID: sensorID,
		Value:    body.Value,
		Raw:      body.Raw,
	}

	response, err := h.lightService.RegisterLightLevel(req)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// ---------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------

func (h *Handlers) validateDeviceKey(r *http.Request) bool {
	return h.deviceAPIKey != "" && r.Header.Get("X-Device-Key") == h.deviceAPIKey
}

// extractSensorID extrai o ID do sensor do path /sensors/{id}/...
func extractSensorID(path string) string {
	path = strings.TrimPrefix(path, "/sensors/")
	parts := strings.Split(path, "/")
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

// extractDeviceSensorID extrai o ID do sensor do path /sensors/device/{id}/...
func extractDeviceSensorID(path string) string {
	path = strings.TrimPrefix(path, "/sensors/device/")
	parts := strings.Split(path, "/")
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func (h *Handlers) handleDomainError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrSensorNotFound:
		h.respondWithError(w, http.StatusNotFound, err.Error())
	case domain.ErrSensorAccessDenied:
		h.respondWithError(w, http.StatusForbidden, err.Error())
	case domain.ErrInvalidSensorID:
		h.respondWithError(w, http.StatusBadRequest, err.Error())
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

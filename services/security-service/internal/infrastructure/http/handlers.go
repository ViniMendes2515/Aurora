package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"aurora/services/security-service/internal/application"
)

// Handlers contém os handlers HTTP do security-service
type Handlers struct {
	alarmService *application.AlarmService
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(alarmService *application.AlarmService) *Handlers {
	return &Handlers{alarmService: alarmService}
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Error string `json:"error"`
}

// GetRecentAlarms handler para GET /alarms
func (h *Handlers) GetRecentAlarms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	alarms, err := h.alarmService.GetRecentAlarms(20)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "failed to fetch alarms")
		return
	}

	h.respondWithJSON(w, http.StatusOK, alarms)
}

// TriggerAlarm handler para POST /alarms/trigger
// - Chamadas do frontend normalmente mandam apenas { location }
// - Chamadas internas (rules-service) podem enviar trigger_type / sensor_id
func (h *Handlers) TriggerAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		Location    string `json:"location"`
		TriggerType string `json:"trigger_type"`
		SensorID    string `json:"sensor_id"`
	}

	// Body é opcional — se não vier nada, usamos defaults
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		// Ignora erro de body vazio/sem JSON
	}
	if body.Location == "" {
		body.Location = "manual"
	}
	triggerType := body.TriggerType
	if triggerType == "" {
		triggerType = "manual"
	}
	sensorID := body.SensorID
	if sensorID == "" {
		sensorID = "manual"
	}

	response, err := h.alarmService.TriggerAlarm(triggerType, sensorID, body.Location)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "failed to trigger alarm")
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// SilenceAlarm handler para POST /alarms/{id}/silence
func (h *Handlers) SilenceAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	alarmID := extractAlarmID(r.URL.Path)
	if alarmID == "" {
		h.respondWithError(w, http.StatusBadRequest, "alarm ID is required")
		return
	}

	response, err := h.alarmService.SilenceAlarm(alarmID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "alarm not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func extractAlarmID(path string) string {
	path = strings.TrimPrefix(path, "/alarms/")
	parts := strings.Split(path, "/")
	if len(parts) >= 1 && parts[0] != "" {
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

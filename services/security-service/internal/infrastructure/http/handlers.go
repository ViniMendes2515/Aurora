package http

import (
	"net/http"

	"aurora/services/security-service/internal/application"

	"github.com/gin-gonic/gin"
)

// Handlers contém os handlers HTTP do security-service
type Handlers struct {
	alarmService *application.AlarmService
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(alarmService *application.AlarmService) *Handlers {
	return &Handlers{alarmService: alarmService}
}

// GetRecentAlarms handler para GET /alarms
func (h *Handlers) GetRecentAlarms(c *gin.Context) {
	alarms, err := h.alarmService.GetRecentAlarms(20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch alarms"})
		return
	}
	c.JSON(http.StatusOK, alarms)
}

// TriggerAlarm handler para POST /alarms/trigger
// Body opcional: { location, trigger_type, sensor_id }
// Chamadas internas (rules-service via X-Device-Key) podem enviar trigger_type e sensor_id.
func (h *Handlers) TriggerAlarm(c *gin.Context) {
	var body struct {
		Location    string `json:"location"`
		TriggerType string `json:"trigger_type"`
		SensorID    string `json:"sensor_id"`
	}

	// Body é opcional — ignoramos erro de decode (ex: body vazio)
	_ = c.ShouldBindJSON(&body)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to trigger alarm"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SilenceAlarm handler para POST /alarms/:id/silence
func (h *Handlers) SilenceAlarm(c *gin.Context) {
	alarmID := c.Param("id")
	if alarmID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alarm ID is required"})
		return
	}

	response, err := h.alarmService.SilenceAlarm(alarmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alarm not found"})
		return
	}

	c.JSON(http.StatusOK, response)
}

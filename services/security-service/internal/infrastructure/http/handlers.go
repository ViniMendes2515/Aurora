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

// GetRecentAlarms godoc
// @Summary      Lista os alarmes recentes
// @Tags         alarms
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   interface{}
// @Failure      500  {object}  map[string]string
// @Router       /alarms [get]
func (h *Handlers) GetRecentAlarms(c *gin.Context) {
	alarms, err := h.alarmService.GetRecentAlarms(20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch alarms"})
		return
	}
	c.JSON(http.StatusOK, alarms)
}

// TriggerAlarm godoc
// @Summary      Dispara um alarme
// @Tags         alarms
// @Accept       json
// @Produce      json
// @Param        body  body      object  false  "Dados opcionais do alarme (location, trigger_type, sensor_id)"
// @Success      200   {object}  interface{}
// @Failure      500   {object}  map[string]string
// @Router       /alarms/trigger [post]
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

	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	response, err := h.alarmService.TriggerAlarm(userIDStr, triggerType, sensorID, body.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to trigger alarm"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SilenceAlarm godoc
// @Summary      Silencia um alarme ativo
// @Tags         alarms
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID do alarme"
// @Success      200  {object}  interface{}
// @Failure      404  {object}  map[string]string
// @Router       /alarms/{id}/silence [post]
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

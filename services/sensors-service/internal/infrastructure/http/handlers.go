package http

import (
	"net/http"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/domain"

	"github.com/gin-gonic/gin"
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

// ---------------------------------------------------------------
// Endpoints para usuários autenticados (JWT)
// ---------------------------------------------------------------

// RegisterMotion godoc
// @Summary      Registra evento de movimento via usuário autenticado
// @Tags         sensors
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID do sensor"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /sensors/{id}/motion [post]
func (h *Handlers) RegisterMotion(c *gin.Context) {
	sensorID := c.Param("id")
	if sensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sensor ID is required"})
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	req := application.RegisterMotionRequest{
		SensorID: sensorID,
		UserID:   userID,
	}

	response, err := h.motionService.RegisterMotion(req)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListSensors godoc
// @Summary      Lista sensores do usuário autenticado
// @Tags         sensors
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   interface{}
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /sensors [get]
func (h *Handlers) ListSensors(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	sensors, err := h.sensorRepo.FindByOwnerID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sensors"})
		return
	}

	c.JSON(http.StatusOK, sensors)
}

// GetMotionHistory godoc
// @Summary      Histórico de eventos de movimento de um sensor
// @Tags         sensors
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID do sensor"
// @Success      200  {array}   interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /sensors/{id}/motion [get]
func (h *Handlers) GetMotionHistory(c *gin.Context) {
	sensorID := c.Param("id")
	if sensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sensor ID is required"})
		return
	}

	records, err := h.sensorRepo.GetMotionRecords(sensorID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch motion records"})
		return
	}

	c.JSON(http.StatusOK, records)
}

// GetLightHistory godoc
// @Summary      Histórico de eventos de luminosidade de um sensor
// @Tags         sensors
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID do sensor"
// @Success      200  {array}   interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /sensors/{id}/light [get]
func (h *Handlers) GetLightHistory(c *gin.Context) {
	sensorID := c.Param("id")
	if sensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sensor ID is required"})
		return
	}

	records, err := h.lightService.GetLightRecords(sensorID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch light records"})
		return
	}

	c.JSON(http.StatusOK, records)
}

// ---------------------------------------------------------------
// Endpoints para dispositivos (API Key — sem JWT)
// ---------------------------------------------------------------

// RegisterDeviceMotion godoc
// @Summary      Registra evento de movimento via dispositivo (ESP32)
// @Tags         device
// @Produce      json
// @Param        X-Device-Key  header    string  true  "Chave de autenticação do dispositivo"
// @Param        id            path      string  true  "ID do sensor"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /sensors/device/{id}/motion [post]
func (h *Handlers) RegisterDeviceMotion(c *gin.Context) {
	if !h.validateDeviceKey(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid device key"})
		return
	}

	sensorID := c.Param("id")
	if sensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sensor ID is required"})
		return
	}

	response, err := h.motionService.RegisterDeviceMotion(sensorID)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// RegisterDeviceLight godoc
// @Summary      Registra evento de luminosidade via dispositivo (ESP32)
// @Tags         device
// @Accept       json
// @Produce      json
// @Param        X-Device-Key  header    string  true   "Chave de autenticação do dispositivo"
// @Param        id            path      string  true   "ID do sensor"
// @Param        body          body      object  true   "Valor de luminosidade"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /sensors/device/{id}/light [post]
func (h *Handlers) RegisterDeviceLight(c *gin.Context) {
	if !h.validateDeviceKey(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid device key"})
		return
	}

	sensorID := c.Param("id")
	if sensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sensor ID is required"})
		return
	}

	var body struct {
		Value float64 `json:"value"`
		Raw   int     `json:"raw"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req := application.RegisterLightLevelRequest{
		SensorID: sensorID,
		Value:    body.Value,
		Raw:      body.Raw,
	}

	response, err := h.lightService.RegisterLightLevel(req)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ---------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------

func (h *Handlers) validateDeviceKey(c *gin.Context) bool {
	return h.deviceAPIKey != "" && c.GetHeader("X-Device-Key") == h.deviceAPIKey
}

func (h *Handlers) handleDomainError(c *gin.Context, err error) {
	switch err {
	case domain.ErrSensorNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrSensorAccessDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case domain.ErrInvalidSensorID:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

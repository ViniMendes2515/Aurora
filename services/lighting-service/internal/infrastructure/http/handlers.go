package http

import (
	"net/http"

	"aurora/services/lighting-service/internal/application"
	"aurora/services/lighting-service/internal/domain"
	"aurora/services/lighting-service/internal/infrastructure/device"

	"github.com/gin-gonic/gin"
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

// ListLights godoc
// @Summary      Lista todas as luzes do usuário
// @Tags         lights
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   interface{}
// @Failure      500  {object}  map[string]string
// @Router       /lights [get]
func (h *Handlers) ListLights(c *gin.Context) {
	userID := c.GetString("userID")
	lights, err := h.lightService.ListLights(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch lights"})
		return
	}
	c.JSON(http.StatusOK, lights)
}

// TurnOn godoc
// @Summary      Liga uma luz
// @Tags         lights
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID da luz"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /lights/{id}/on [post]
func (h *Handlers) TurnOn(c *gin.Context) {
	lightID := c.Param("id")
	if lightID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "light ID is required"})
		return
	}

	userID := c.GetString("userID")
	response, err := h.lightService.TurnOn(lightID, userID)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// TurnOff godoc
// @Summary      Desliga uma luz
// @Tags         lights
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID da luz"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /lights/{id}/off [post]
func (h *Handlers) TurnOff(c *gin.Context) {
	lightID := c.Param("id")
	if lightID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "light ID is required"})
		return
	}

	userID := c.GetString("userID")
	response, err := h.lightService.TurnOff(lightID, userID)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetLightStatus godoc
// @Summary      Retorna o status de uma luz
// @Tags         lights
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID da luz"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /lights/{id}/status [get]
func (h *Handlers) GetLightStatus(c *gin.Context) {
	lightID := c.Param("id")
	if lightID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "light ID is required"})
		return
	}

	userID := c.GetString("userID")
	response, err := h.lightService.GetStatus(lightID, userID)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// RegisterDevice godoc
// @Summary      Registra um dispositivo ESP32
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Device-Key  header    string  true  "Chave de autenticação do dispositivo"
// @Param        body          body      object  true  "Dados do dispositivo"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /devices/register [post]
func (h *Handlers) RegisterDevice(c *gin.Context) {
	if c.GetHeader("X-Device-Key") != h.deviceAPIKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid device key"})
		return
	}

	var body struct {
		DeviceID string `json:"device_id"`
		IP       string `json:"ip"`
		Port     int    `json:"port"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
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

	c.JSON(http.StatusOK, gin.H{
		"status":    "registered",
		"device_id": body.DeviceID,
		"ip":        body.IP,
	})
}

func (h *Handlers) handleDomainError(c *gin.Context, err error) {
	switch err {
	case domain.ErrLightNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrLightAccessDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case domain.ErrInvalidLightID:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case domain.ErrDeviceUnreachable:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

package http

import (
	"net/http"
	"strconv"

	"aurora/services/schedule-service/internal/application"
	"aurora/services/schedule-service/internal/domain"

	"github.com/gin-gonic/gin"
)

// Handlers contém os handlers HTTP do schedule-service
type Handlers struct {
	service *application.ScheduleService
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(service *application.ScheduleService) *Handlers {
	return &Handlers{service: service}
}

// CreateSchedule godoc
// @Summary      Cria um novo agendamento
// @Tags         schedules
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      application.CreateScheduleRequest  true  "Dados do agendamento"
// @Success      201   {object}  interface{}
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /schedules [post]
func (h *Handlers) CreateSchedule(c *gin.Context) {
	userID := c.GetString("userID")

	var req application.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.OwnerID = userID

	resp, err := h.service.CreateSchedule(req)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListSchedules godoc
// @Summary      Lista agendamentos do usuário
// @Tags         schedules
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   interface{}
// @Failure      500  {object}  map[string]string
// @Router       /schedules [get]
func (h *Handlers) ListSchedules(c *gin.Context) {
	userID := c.GetString("userID")

	schedules, err := h.service.ListSchedules(userID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, schedules)
}

// GetSchedule godoc
// @Summary      Retorna um agendamento por ID
// @Tags         schedules
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID do agendamento"
// @Success      200  {object}  interface{}
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /schedules/{id} [get]
func (h *Handlers) GetSchedule(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	resp, err := h.service.GetSchedule(id, userID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateSchedule godoc
// @Summary      Atualiza um agendamento
// @Tags         schedules
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string                          true  "ID do agendamento"
// @Param        body  body      application.UpdateScheduleRequest  true  "Dados do agendamento"
// @Success      200   {object}  interface{}
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /schedules/{id} [put]
func (h *Handlers) UpdateSchedule(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	var req application.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.UpdateSchedule(id, userID, req)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ToggleSchedule godoc
// @Summary      Ativa/desativa um agendamento
// @Tags         schedules
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID do agendamento"
// @Success      200  {object}  interface{}
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /schedules/{id}/toggle [patch]
func (h *Handlers) ToggleSchedule(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	resp, err := h.service.ToggleSchedule(id, userID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteSchedule godoc
// @Summary      Remove um agendamento
// @Tags         schedules
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID do agendamento"
// @Success      200  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /schedules/{id} [delete]
func (h *Handlers) DeleteSchedule(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	if err := h.service.DeleteSchedule(id, userID); err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ListExecutions godoc
// @Summary      Histórico de execuções de um agendamento
// @Tags         schedules
// @Security     BearerAuth
// @Produce      json
// @Param        id     path      string  true   "ID do agendamento"
// @Param        limit  query     int     false  "Limite de registros (padrão 50)"
// @Success      200    {array}   interface{}
// @Failure      403    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Router       /schedules/{id}/history [get]
func (h *Handlers) ListExecutions(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	executions, err := h.service.ListExecutions(id, userID, limit)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, executions)
}

// PreviewSchedule godoc
// @Summary      Pré-visualiza as próximas execuções de um agendamento
// @Tags         schedules
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "schedule_id e count"
// @Success      200   {array}   interface{}
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Router       /schedules/preview [post]
func (h *Handlers) PreviewSchedule(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		ScheduleID string `json:"schedule_id"`
		Count      int    `json:"count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.PreviewSchedule(req.ScheduleID, userID, req.Count)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// mapError traduz erros de domínio para respostas HTTP adequadas.
func mapError(c *gin.Context, err error) {
	switch err {
	case domain.ErrScheduleNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrScheduleAccessDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case domain.ErrInvalidSchedule, domain.ErrMissingCronExpression, domain.ErrMissingRunAt, domain.ErrInvalidAction:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

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

// CreateSchedule handler para POST /schedules
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

// ListSchedules handler para GET /schedules
func (h *Handlers) ListSchedules(c *gin.Context) {
	userID := c.GetString("userID")

	schedules, err := h.service.ListSchedules(userID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, schedules)
}

// GetSchedule handler para GET /schedules/:id
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

// UpdateSchedule handler para PUT /schedules/:id
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

// ToggleSchedule handler para PATCH /schedules/:id/toggle
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

// DeleteSchedule handler para DELETE /schedules/:id
func (h *Handlers) DeleteSchedule(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	if err := h.service.DeleteSchedule(id, userID); err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ListExecutions handler para GET /schedules/:id/history
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

// PreviewSchedule handler para POST /schedules/preview
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

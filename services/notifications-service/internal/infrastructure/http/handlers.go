package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"aurora/services/notifications-service/internal/application"
	"aurora/services/notifications-service/internal/domain"
)

// Handler contem os handlers HTTP do notifications-service
type Handler struct {
	service *application.NotificationService
}

// NewHandler cria um novo Handler
func NewHandler(service *application.NotificationService) *Handler {
	return &Handler{service: service}
}

// Health godoc
// @Summary      Health check do serviço
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "notifications-service"})
}

// GetNotifications godoc
// @Summary      Lista notificações do usuário autenticado
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   domain.Notification
// @Failure      500  {object}  map[string]string
// @Router       /notifications [get]
func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.GetString("userID")

	notifications, err := h.service.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar notificacoes"})
		return
	}

	// Garante que a resposta seja [] e nao null quando vazia
	if notifications == nil {
		notifications = []*domain.Notification{}
	}
	c.JSON(http.StatusOK, notifications)
}

// MarkAsRead godoc
// @Summary      Marca uma notificação como lida
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID da notificação"
// @Success      200  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /notifications/{id}/read [patch]
func (h *Handler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	if err := h.service.MarkAsRead(id, userID); err != nil {
		switch err {
		case domain.ErrNotificationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "notificacao nao encontrada"})
		case domain.ErrAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": "acesso negado"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao marcar notificacao"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "notificacao marcada como lida"})
}

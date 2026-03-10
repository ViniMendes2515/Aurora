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

// Health retorna o status de saude do servico
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "notifications-service"})
}

// GetNotifications lista as notificacoes do usuario autenticado (proprias + broadcast)
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

// MarkAsRead marca uma notificacao como lida pelo seu ID
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

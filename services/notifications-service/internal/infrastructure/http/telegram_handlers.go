package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"aurora/services/notifications-service/internal/application"
	"aurora/services/notifications-service/internal/domain"
)

// TelegramHandler contem os handlers HTTP para gerenciamento de preferencias Telegram
type TelegramHandler struct {
	service *application.TelegramService
}

func NewTelegramHandler(service *application.TelegramService) *TelegramHandler {
	return &TelegramHandler{service: service}
}

// GenerateLinkToken godoc
// @Summary      Gera token para vincular conta Telegram
// @Tags         telegram
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /telegram/link [post]
func (h *TelegramHandler) GenerateLinkToken(c *gin.Context) {
	userID := c.GetString("userID")

	token, err := h.service.GenerateLinkToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "Abra o bot Aurora no Telegram e envie: /start " + token,
		"expires": "15 minutos",
	})
}

// UnlinkTelegram godoc
// @Summary      Remove vinculacao com Telegram
// @Tags         telegram
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /telegram/link [delete]
func (h *TelegramHandler) UnlinkTelegram(c *gin.Context) {
	userID := c.GetString("userID")

	if err := h.service.UnlinkByUserID(userID); err != nil {
		if err == domain.ErrTelegramNotLinked {
			c.JSON(http.StatusNotFound, gin.H{"error": "nenhuma conta Telegram vinculada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao desvincular conta"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conta Telegram desvinculada com sucesso"})
}

// GetTelegramPreferences godoc
// @Summary      Retorna preferencias Telegram do usuario
// @Tags         telegram
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  domain.TelegramPreference
// @Failure      404  {object}  map[string]string
// @Router       /telegram/preferences [get]
func (h *TelegramHandler) GetTelegramPreferences(c *gin.Context) {
	userID := c.GetString("userID")

	pref, err := h.service.GetPreferences(userID)
	if err != nil {
		if err == domain.ErrTelegramNotLinked {
			c.JSON(http.StatusNotFound, gin.H{"error": "nenhuma conta Telegram vinculada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar preferencias"})
		return
	}

	c.JSON(http.StatusOK, pref)
}

// UpdateTelegramPreferences godoc
// @Summary      Atualiza tipos de notificacao habilitados para Telegram
// @Tags         telegram
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      map[string][]string  true  "enabled_types"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /telegram/preferences [put]
func (h *TelegramHandler) UpdateTelegramPreferences(c *gin.Context) {
	userID := c.GetString("userID")

	var body struct {
		EnabledTypes []string `json:"enabled_types" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo enabled_types e obrigatorio"})
		return
	}

	if err := h.service.UpdatePreferences(userID, body.EnabledTypes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao atualizar preferencias"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "preferencias atualizadas com sucesso"})
}

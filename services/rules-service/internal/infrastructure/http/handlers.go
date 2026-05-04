package http

import (
	"net/http"

	"aurora/services/rules-service/internal/application"
	"aurora/services/rules-service/internal/domain"

	"github.com/gin-gonic/gin"
)

// Handlers contém os handlers HTTP do rules-service
type Handlers struct {
	rulesEngine *application.RulesEngine
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(rulesEngine *application.RulesEngine) *Handlers {
	return &Handlers{rulesEngine: rulesEngine}
}

// ListRules godoc
// @Summary      Lista regras de automação do usuário
// @Tags         rules
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   interface{}
// @Failure      500  {object}  map[string]string
// @Router       /rules [get]
func (h *Handlers) ListRules(c *gin.Context) {
	userID := c.GetString("userID")
	rules, err := h.rulesEngine.ListRules(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

// CreateRule godoc
// @Summary      Cria uma nova regra de automação
// @Tags         rules
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      application.CreateRuleRequest  true  "Dados da regra"
// @Success      201   {object}  interface{}
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /rules [post]
func (h *Handlers) CreateRule(c *gin.Context) {
	userID := c.GetString("userID")

	var req application.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.OwnerID = userID

	response, err := h.rulesEngine.CreateRule(req)
	if err != nil {
		switch err {
		case domain.ErrInvalidRule:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create rule"})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

// DeleteRule godoc
// @Summary      Remove uma regra de automação
// @Tags         rules
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID da regra"
// @Success      200  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /rules/{id} [delete]
func (h *Handlers) DeleteRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule ID is required"})
		return
	}

	userID := c.GetString("userID")
	if err := h.rulesEngine.DeleteRule(ruleID, userID); err != nil {
		switch err {
		case domain.ErrRuleNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case domain.ErrRuleAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete rule"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ToggleRule godoc
// @Summary      Ativa/desativa uma regra de automação
// @Tags         rules
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "ID da regra"
// @Success      200  {object}  interface{}
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /rules/{id}/toggle [patch]
func (h *Handlers) ToggleRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule ID is required"})
		return
	}

	userID := c.GetString("userID")
	resp, err := h.rulesEngine.ToggleRule(ruleID, userID)
	if err != nil {
		switch err {
		case domain.ErrRuleNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case domain.ErrRuleAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle rule"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateRule godoc
// @Summary      Atualiza uma regra de automação
// @Tags         rules
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string                         true  "ID da regra"
// @Param        body  body      application.CreateRuleRequest  true  "Dados da regra"
// @Success      200   {object}  interface{}
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /rules/{id} [put]
func (h *Handlers) UpdateRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule ID is required"})
		return
	}

	userID := c.GetString("userID")

	var req application.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.rulesEngine.UpdateRule(ruleID, userID, req)
	if err != nil {
		switch err {
		case domain.ErrRuleNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case domain.ErrRuleAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case domain.ErrInvalidRule:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update rule"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

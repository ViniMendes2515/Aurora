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

// ListRules handler para GET /rules
func (h *Handlers) ListRules(c *gin.Context) {
	userID := c.GetString("userID")
	rules, err := h.rulesEngine.ListRules(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

// CreateRule handler para POST /rules
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

// DeleteRule handler para DELETE /rules/:id
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

// ToggleRule handler para PATCH /rules/:id/toggle
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

// UpdateRule handler para PUT /rules/:id
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

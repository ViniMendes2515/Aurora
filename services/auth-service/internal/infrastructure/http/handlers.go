package http

import (
	"net/http"

	"aurora/services/auth-service/internal/application"
	"aurora/services/auth-service/internal/domain"

	"github.com/gin-gonic/gin"
)

// Handlers contém os handlers HTTP
type Handlers struct {
	authService *application.AuthService
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(authService *application.AuthService) *Handlers {
	return &Handlers{authService: authService}
}

// Register handler para POST /auth/register
func (h *Handlers) Register(c *gin.Context) {
	var req application.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.authService.Register(req)
	if err != nil {
		switch err {
		case domain.ErrInvalidEmail:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrInvalidPassword:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handler para POST /auth/login
func (h *Handlers) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case domain.ErrTokenGeneration:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

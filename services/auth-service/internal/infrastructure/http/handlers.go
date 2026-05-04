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

// Register godoc
// @Summary      Registro de novo usuário
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      application.RegisterRequest   true  "Dados de registro"
// @Success      201   {object}  application.RegisterResponse
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/register [post]
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

// Login godoc
// @Summary      Login de usuário
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      application.LoginRequest   true  "Credenciais de login"
// @Success      200   {object}  application.LoginResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/login [post]
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

package application

import (
	"aurora/services/auth-service/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

// AuthService implementa os casos de uso de autenticação
type AuthService struct {
	userRepo   domain.UserRepository
	jwtManager domain.JWTManager
}

// NewAuthService cria uma nova instância de AuthService
func NewAuthService(userRepo domain.UserRepository, jwtManager domain.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// RegisterRequest representa os dados de entrada para registro
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse representa a resposta do registro
type RegisterResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// LoginRequest representa os dados de entrada para login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse representa a resposta do login
type LoginResponse struct {
	Token string `json:"token"`
}

// Register executa o caso de uso de registro de usuário
func (s *AuthService) Register(req RegisterRequest) (*RegisterResponse, error) {
	// Validar senha
	if len(req.Password) < 6 {
		return nil, domain.ErrInvalidPassword
	}

	// Verificar se usuário já existe
	if s.userRepo.ExistsByEmail(req.Email) {
		return nil, domain.ErrUserAlreadyExists
	}

	// Gerar hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Criar entidade de domínio
	user := domain.NewUser(req.Email, string(hashedPassword))

	// Validar email
	if err := user.ValidateEmail(); err != nil {
		return nil, err
	}

	// Salvar via repositório (infraestrutura)
	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

// Login executa o caso de uso de login
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	// Buscar usuário pelo email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Comparar senha com hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Gerar JWT via contrato (implementado na infraestrutura)
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, domain.ErrTokenGeneration
	}

	return &LoginResponse{
		Token: token,
	}, nil
}

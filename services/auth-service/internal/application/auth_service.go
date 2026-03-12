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
	if len(req.Password) < 6 {
		return nil, domain.ErrInvalidPassword
	}

	if s.userRepo.ExistsByEmail(req.Email) {
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(req.Email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		ID:    user.ID,
		Email: user.Email.String(),
	}, nil
}

// Login executa o caso de uso de login
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := s.jwtManager.GenerateToken(user.ID, user.Email.String())
	if err != nil {
		return nil, domain.ErrTokenGeneration
	}

	return &LoginResponse{
		Token: token,
	}, nil
}

package domain

import "errors"

var (
	// ErrInvalidEmail indica que o formato do email é inválido
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidPassword indica que a senha não atende aos requisitos
	ErrInvalidPassword = errors.New("password must be at least 6 characters")

	// ErrUserNotFound indica que o usuário não foi encontrado
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists indica que já existe um usuário com o mesmo email
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidCredentials indica credenciais inválidas
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrTokenGeneration indica falha na geração do token
	ErrTokenGeneration = errors.New("failed to generate token")

	// ErrTokenValidation indica falha na validação do token
	ErrTokenValidation = errors.New("invalid or expired token")
)

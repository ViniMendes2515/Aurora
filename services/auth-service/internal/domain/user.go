package domain

import (
	"time"

	"github.com/google/uuid"
)

// User representa a entidade de usuário no domínio
type User struct {
	ID           string
	Email        Email
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser cria uma nova instância de User. Retorna erro se o email for inválido.
func NewUser(email, passwordHash string) (*User, error) {
	e, err := NewEmail(email)

	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Email:        e,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

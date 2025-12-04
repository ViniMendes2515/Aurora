package domain

import (
	"time"

	"github.com/google/uuid"
)

// User representa a entidade de usuário no domínio
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser cria uma nova instância de User
func NewUser(email, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ValidateEmail verifica se o email tem formato válido
func (u *User) ValidateEmail() error {
	if u.Email == "" {
		return ErrInvalidEmail
	}
	// Validação básica de email
	if len(u.Email) < 3 {
		return ErrInvalidEmail
	}
	hasAt := false
	hasDot := false
	for i, c := range u.Email {
		if c == '@' {
			hasAt = true
		}
		if hasAt && c == '.' && i > 0 {
			hasDot = true
		}
	}
	if !hasAt || !hasDot {
		return ErrInvalidEmail
	}
	return nil
}

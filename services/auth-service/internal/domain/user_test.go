package domain_test

import (
	"testing"

	"aurora/services/auth-service/internal/domain"
)

func TestNewUser_CamposPreenchidos(t *testing.T) {
	u, err := domain.NewUser("user@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("NewUser() erro inesperado: %v", err)
	}
	if u.ID == "" {
		t.Error("ID nao deve ser vazio")
	}
	if u.Email.String() != "user@example.com" {
		t.Errorf("Email esperado user@example.com, obteve %s", u.Email.String())
	}
	if u.PasswordHash != "hashedpassword" {
		t.Error("PasswordHash nao corresponde")
	}
	if u.CreatedAt.IsZero() {
		t.Error("CreatedAt nao deve ser zero")
	}
	if u.UpdatedAt.IsZero() {
		t.Error("UpdatedAt nao deve ser zero")
	}
}

func TestNewUser_IDUnico(t *testing.T) {
	u1, _ := domain.NewUser("a@a.com", "hash1")
	u2, _ := domain.NewUser("b@b.com", "hash2")
	if u1.ID == u2.ID {
		t.Error("IDs de usuarios distintos nao devem ser iguais")
	}
}

func TestNewUser_EmailInvalido(t *testing.T) {
	_, err := domain.NewUser("invalido", "hash")
	if err != domain.ErrInvalidEmail {
		t.Errorf("esperado ErrInvalidEmail, obteve %v", err)
	}
}

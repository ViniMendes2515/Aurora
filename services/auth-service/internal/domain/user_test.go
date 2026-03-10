package domain_test

import (
"testing"

"aurora/services/auth-service/internal/domain"
)

func TestNewUser_CamposPreenchidos(t *testing.T) {
	u := domain.NewUser("user@example.com", "hashedpassword")
	if u.ID == "" {
		t.Error("ID nao deve ser vazio")
	}
	if u.Email != "user@example.com" {
		t.Errorf("Email esperado user@example.com, obteve %s", u.Email)
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
	u1 := domain.NewUser("a@a.com", "hash1")
	u2 := domain.NewUser("b@b.com", "hash2")
	if u1.ID == u2.ID {
		t.Error("IDs de usuarios distintos nao devem ser iguais")
	}
}

func TestValidateEmail(t *testing.T) {
	testes := []struct {
		nome    string
		email   string
		comErro bool
	}{
		{"email valido", "user@example.com", false},
		{"email com subdominio", "user@mail.example.com", false},
		{"email vazio", "", true},
		{"sem arroba", "userexample.com", true},
		{"sem ponto apos arroba", "user@examplecom", true},
		{"muito curto", "a@", true},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
u := &domain.User{Email: tt.email}
			err := u.ValidateEmail()
			if (err != nil) != tt.comErro {
				t.Errorf("ValidateEmail() erro = %v, esperado erro = %v", err, tt.comErro)
			}
			if tt.comErro && err != domain.ErrInvalidEmail {
				t.Errorf("esperado ErrInvalidEmail, obteve %v", err)
			}
		})
	}
}

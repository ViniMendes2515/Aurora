package domain_test

import (
	"testing"

	"aurora/services/auth-service/internal/domain"
)

func TestNewEmail(t *testing.T) {
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
			e, err := domain.NewEmail(tt.email)
			if (err != nil) != tt.comErro {
				t.Errorf("NewEmail() erro = %v, esperado erro = %v", err, tt.comErro)
			}
			if tt.comErro && err != domain.ErrInvalidEmail {
				t.Errorf("esperado ErrInvalidEmail, obteve %v", err)
			}
			if !tt.comErro && e.String() != tt.email {
				t.Errorf("Email.String() = %q, esperado %q", e.String(), tt.email)
			}
		})
	}
}

package domain_test

import (
	"testing"

	"aurora/services/lighting-service/internal/domain"
)

func TestNewLight_EstadoInicial(t *testing.T) {
	l := domain.NewLight("id1", "Sala", "sala", "user1", "device1")
	if l.ID != "id1" {
		t.Errorf("ID esperado id1, obteve %s", l.ID)
	}
	if l.State != domain.LightStateOff {
		t.Errorf("Estado inicial deve ser off, obteve %s", l.State)
	}
	if l.CreatedAt.IsZero() {
		t.Error("CreatedAt nao deve ser zero")
	}
	if l.UpdatedAt.IsZero() {
		t.Error("UpdatedAt nao deve ser zero")
	}
}

func TestLight_TurnOn(t *testing.T) {
	l := domain.NewLight("id1", "Sala", "sala", "user1", "device1")
	l.TurnOn()
	if l.State != domain.LightStateOn {
		t.Errorf("Estado deve ser on apos TurnOn, obteve %s", l.State)
	}
	if !l.IsOn() {
		t.Error("IsOn deve retornar true apos TurnOn")
	}
}

func TestLight_TurnOff(t *testing.T) {
	l := domain.NewLight("id1", "Sala", "sala", "user1", "device1")
	l.TurnOn()
	l.TurnOff()
	if l.State != domain.LightStateOff {
		t.Errorf("Estado deve ser off apos TurnOff, obteve %s", l.State)
	}
	if l.IsOn() {
		t.Error("IsOn deve retornar false apos TurnOff")
	}
}

func TestLight_BelongsTo(t *testing.T) {
	testes := []struct {
		nome     string
		ownerID  string
		userID   string
		esperado bool
	}{
		{"dono correto", "user1", "user1", true},
		{"dono errado", "user1", "user2", false},
		{"wildcard", "*", "qualquer", true},
	}
	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			l := domain.NewLight("id1", "Sala", "sala", tt.ownerID, "device1")
			if got := l.BelongsTo(tt.userID); got != tt.esperado {
				t.Errorf("BelongsTo() = %v, esperado %v", got, tt.esperado)
			}
		})
	}
}

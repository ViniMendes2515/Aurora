package domain_test

import (
	"testing"

	"aurora/services/rules-service/internal/domain"
)

func TestNewRule_CamposPreenchidos(t *testing.T) {
	r := domain.NewRule("r1", "Regra 1", "user1",
		domain.TriggerMotion, "sensor1", 0,
		domain.ActionTurnOnLight, "light1")
	if r.ID != "r1" {
		t.Errorf("ID esperado r1, obteve %s", r.ID)
	}
	if !r.Enabled {
		t.Error("Nova regra deve ser criada habilitada")
	}
	if r.CreatedAt.IsZero() {
		t.Error("CreatedAt nao deve ser zero")
	}
}

func TestNewRule_ThresholdZeroUsaPadrao(t *testing.T) {
	r := domain.NewRule("r1", "Regra", "user1",
		domain.TriggerLight, "sensor1", 0,
		domain.ActionTurnOnLight, "light1")
	if r.TriggerThreshold != domain.DefaultLightThreshold {
		t.Errorf("Threshold zero deve usar padrao %v, obteve %v",
			domain.DefaultLightThreshold, r.TriggerThreshold)
	}
}

func TestNewRule_ThresholdPersonalizado(t *testing.T) {
	r := domain.NewRule("r2", "Regra", "user1",
		domain.TriggerLight, "sensor1", 50.0,
		domain.ActionTurnOnLight, "light1")
	if r.TriggerThreshold != 50.0 {
		t.Errorf("Threshold esperado 50.0, obteve %v", r.TriggerThreshold)
	}
}

func TestRule_LightThreshold(t *testing.T) {
	testes := []struct {
		nome      string
		threshold float64
		esperado  float64
	}{
		{"personalizado valido", 45.0, 45.0},
		{"limite 100", 100.0, 100.0},
		{"zero usa padrao", 0, domain.DefaultLightThreshold},
		{"negativo usa padrao", -10, domain.DefaultLightThreshold},
		{"acima de 100 usa padrao", 101.0, domain.DefaultLightThreshold},
	}
	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			r := &domain.Rule{TriggerThreshold: tt.threshold}
			if got := r.LightThreshold(); got != tt.esperado {
				t.Errorf("LightThreshold() = %v, esperado %v", got, tt.esperado)
			}
		})
	}
}

func TestRule_BelongsTo(t *testing.T) {
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
			r := &domain.Rule{OwnerID: tt.ownerID}
			if got := r.BelongsTo(tt.userID); got != tt.esperado {
				t.Errorf("BelongsTo() = %v, esperado %v", got, tt.esperado)
			}
		})
	}
}

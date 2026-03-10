package domain_test

import (
	"testing"

	"aurora/services/sensors-service/internal/domain"
)

func TestNewSensor_CamposPreenchidos(t *testing.T) {
	s := domain.NewSensor("s1", "Sensor Sala", domain.SensorTypeMotion, "sala", "user1")
	if s.ID != "s1" {
		t.Errorf("ID esperado s1, obteve %s", s.ID)
	}
	if s.Type != domain.SensorTypeMotion {
		t.Errorf("Type esperado motion, obteve %s", s.Type)
	}
	if !s.Active {
		t.Error("Novo sensor deve ser criado como ativo")
	}
	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt nao deve ser zero")
	}
}

func TestSensor_BelongsTo(t *testing.T) {
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
			s := domain.NewSensor("s1", "Sensor", domain.SensorTypeMotion, "sala", tt.ownerID)
			if got := s.BelongsTo(tt.userID); got != tt.esperado {
				t.Errorf("BelongsTo() = %v, esperado %v", got, tt.esperado)
			}
		})
	}
}

func TestSensor_IsMotionSensor(t *testing.T) {
	testes := []struct {
		nome        string
		tipo        domain.SensorType
		ehMovimento bool
	}{
		{"tipo motion", domain.SensorTypeMotion, true},
		{"tipo light", domain.SensorTypeLight, false},
		{"tipo temperature", domain.SensorTypeTemperature, false},
		{"tipo door", domain.SensorTypeDoor, false},
	}
	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			s := domain.NewSensor("s1", "Sensor", tt.tipo, "sala", "user1")
			if got := s.IsMotionSensor(); got != tt.ehMovimento {
				t.Errorf("IsMotionSensor() = %v, esperado %v", got, tt.ehMovimento)
			}
		})
	}
}

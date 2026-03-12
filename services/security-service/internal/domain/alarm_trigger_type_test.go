package domain_test

import (
	"testing"

	"aurora/services/security-service/internal/domain"
)

func TestNewAlarmTriggerType(t *testing.T) {
	testes := []struct {
		nome    string
		input   string
		comErro bool
	}{
		{"motion valido", "motion", false},
		{"manual valido", "manual", false},
		{"tipo desconhecido", "schedule", true},
		{"string vazia", "", true},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			trigger, err := domain.NewAlarmTriggerType(tt.input)
			if (err != nil) != tt.comErro {
				t.Errorf("NewAlarmTriggerType(%q) erro = %v, esperado erro = %v", tt.input, err, tt.comErro)
			}
			if !tt.comErro && trigger.String() != tt.input {
				t.Errorf("String() = %q, esperado %q", trigger.String(), tt.input)
			}
		})
	}
}

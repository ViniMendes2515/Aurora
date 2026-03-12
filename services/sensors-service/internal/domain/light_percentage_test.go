package domain_test

import (
	"testing"

	"aurora/services/sensors-service/internal/domain"
)

func TestNewLightPercentage(t *testing.T) {
	testes := []struct {
		nome    string
		input   float64
		comErro bool
	}{
		{"zero (escuro total)", 0, false},
		{"100 (claro total)", 100, false},
		{"valor intermediario", 42.5, false},
		{"negativo", -1, true},
		{"acima de 100", 100.1, true},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			pct, err := domain.NewLightPercentage(tt.input)
			if (err != nil) != tt.comErro {
				t.Errorf("NewLightPercentage(%v) erro = %v, esperado erro = %v", tt.input, err, tt.comErro)
			}
			if !tt.comErro && pct.Float64() != tt.input {
				t.Errorf("Float64() = %v, esperado %v", pct.Float64(), tt.input)
			}
		})
	}
}

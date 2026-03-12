package domain_test

import (
	"testing"

	"aurora/services/rules-service/internal/domain"
)

func TestNewScheduleTime(t *testing.T) {
	testes := []struct {
		nome    string
		input   string
		comErro bool
		hora    int
		minuto  int
	}{
		{"horario valido", "08:30", false, 8, 30},
		{"meia-noite", "00:00", false, 0, 0},
		{"ultimo horario", "23:59", false, 23, 59},
		{"formato sem zero a esquerda", "8:30", true, 0, 0},
		{"formato sem separador", "830", true, 0, 0},
		{"hora invalida", "25:00", true, 0, 0},
		{"minuto invalido", "08:60", true, 0, 0},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			st, err := domain.NewScheduleTime(tt.input)
			if (err != nil) != tt.comErro {
				t.Errorf("NewScheduleTime(%q) erro = %v, esperado erro = %v", tt.input, err, tt.comErro)
			}
			if !tt.comErro {
				if st.Hour() != tt.hora {
					t.Errorf("Hour() = %d, esperado %d", st.Hour(), tt.hora)
				}
				if st.Minute() != tt.minuto {
					t.Errorf("Minute() = %d, esperado %d", st.Minute(), tt.minuto)
				}
				if st.String() != tt.input {
					t.Errorf("String() = %q, esperado %q", st.String(), tt.input)
				}
			}
		})
	}
}

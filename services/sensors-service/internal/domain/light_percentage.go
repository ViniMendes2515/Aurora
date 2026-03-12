package domain

import (
	"encoding/json"
	"errors"
)

// ErrInvalidLightPercentage indica que o valor de luminosidade está fora do intervalo 0-100.
var ErrInvalidLightPercentage = errors.New("light percentage must be between 0 and 100")

// LightPercentage é um Value Object que representa luminosidade em percentual (0-100).
type LightPercentage struct {
	value float64
}

// NewLightPercentage valida e constrói um LightPercentage.
func NewLightPercentage(v float64) (LightPercentage, error) {
	if v < 0 || v > 100 {
		return LightPercentage{}, ErrInvalidLightPercentage
	}
	return LightPercentage{value: v}, nil
}

// Float64 retorna o valor numérico
func (l LightPercentage) Float64() float64 { return l.value }

// MarshalJSON serializa como número simples
func (l LightPercentage) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.value)
}

// UnmarshalJSON desserializa a partir de um número, validando o range.
func (l *LightPercentage) UnmarshalJSON(data []byte) error {
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	pct, err := NewLightPercentage(v)
	if err != nil {
		return err
	}
	*l = pct
	return nil
}

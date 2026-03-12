package domain

import (
	"time"

	"github.com/google/uuid"
)

// LightLevelEvent representa um evento de leitura de luminosidade
type LightLevelEvent struct {
	ID         string          `json:"id"`
	SensorID   string          `json:"sensor_id"`
	Value      LightPercentage `json:"value"` // porcentagem: 0 = escuro, 100 = claro
	Raw        int             `json:"raw"`   // valor bruto do ADC (0-4095)
	RecordedAt time.Time       `json:"recorded_at"`
}

// NewLightLevelEvent cria um novo evento de luminosidade. Retorna erro se value estiver fora de 0-100.
func NewLightLevelEvent(sensorID string, value float64, raw int) (*LightLevelEvent, error) {
	pct, err := NewLightPercentage(value)

	if err != nil {
		return nil, err
	}

	return &LightLevelEvent{
		ID:         uuid.New().String(),
		SensorID:   sensorID,
		Value:      pct,
		Raw:        raw,
		RecordedAt: time.Now(),
	}, nil
}

// Topic retorna o tópico NATS para publicação
func (e *LightLevelEvent) Topic() string {
	return "sensors.light.changed"
}

// LightRecord representa um registro de luminosidade armazenado
type LightRecord struct {
	ID         string
	SensorID   string
	Value      LightPercentage
	Raw        int
	RecordedAt time.Time
}

// NewLightRecord cria um novo registro de luminosidade. Retorna erro se value estiver fora de 0-100.
func NewLightRecord(sensorID string, value float64, raw int) (*LightRecord, error) {
	pct, err := NewLightPercentage(value)

	if err != nil {
		return nil, err
	}

	return &LightRecord{
		ID:         uuid.New().String(),
		SensorID:   sensorID,
		Value:      pct,
		Raw:        raw,
		RecordedAt: time.Now(),
	}, nil
}

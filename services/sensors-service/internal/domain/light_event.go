package domain

import (
	"time"

	"github.com/google/uuid"
)

// LightLevelEvent representa um evento de leitura de luminosidade
type LightLevelEvent struct {
	ID         string    `json:"id"`
	SensorID   string    `json:"sensor_id"`
	Value      float64   `json:"value"`       // porcentagem: 0 = escuro, 100 = claro
	Raw        int       `json:"raw"`         // valor bruto do ADC (0-4095)
	RecordedAt time.Time `json:"recorded_at"`
}

// NewLightLevelEvent cria um novo evento de luminosidade
func NewLightLevelEvent(sensorID string, value float64, raw int) *LightLevelEvent {
	return &LightLevelEvent{
		ID:         uuid.New().String(),
		SensorID:   sensorID,
		Value:      value,
		Raw:        raw,
		RecordedAt: time.Now(),
	}
}

// Topic retorna o tópico NATS para publicação
func (e *LightLevelEvent) Topic() string {
	return "sensors.light.changed"
}

// LightRecord representa um registro de luminosidade armazenado
type LightRecord struct {
	ID         string
	SensorID   string
	Value      float64
	Raw        int
	RecordedAt time.Time
}

// NewLightRecord cria um novo registro de luminosidade
func NewLightRecord(sensorID string, value float64, raw int) *LightRecord {
	return &LightRecord{
		ID:         uuid.New().String(),
		SensorID:   sensorID,
		Value:      value,
		Raw:        raw,
		RecordedAt: time.Now(),
	}
}

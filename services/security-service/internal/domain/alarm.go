package domain

import (
	"errors"
	"time"
)

var (
	ErrAlarmNotFound     = errors.New("alarm event not found")
	ErrDeviceUnreachable = errors.New("device unreachable")
)

// AlarmStatus representa o estado de um alarme
type AlarmStatus string

const (
	AlarmStatusTriggered = AlarmStatus("triggered")
	AlarmStatusSilenced  = AlarmStatus("silenced")
)

// AlarmEvent representa um evento de acionamento de alarme
type AlarmEvent struct {
	ID          string
	TriggerType AlarmTriggerType
	SensorID    string
	Location    string
	Status      AlarmStatus
	TriggeredAt time.Time
	SilencedAt  *time.Time
}

// NewAlarmEvent cria um novo evento de alarme. Retorna erro se triggerType for inválido.
func NewAlarmEvent(triggerType, sensorID, location string) (*AlarmEvent, error) {
	tt, err := NewAlarmTriggerType(triggerType)

	if err != nil {
		return nil, err
	}

	return &AlarmEvent{
		ID:          newID(),
		TriggerType: tt,
		SensorID:    sensorID,
		Location:    location,
		Status:      AlarmStatusTriggered,
		TriggeredAt: time.Now(),
	}, nil
}

// Silence silencia o alarme
func (a *AlarmEvent) Silence() {
	now := time.Now()
	a.Status = AlarmStatusSilenced
	a.SilencedAt = &now
}

// BuzzerClient define o contrato para comunicação com o buzzer do ESP32
type BuzzerClient interface {
	TurnOn(deviceID string) error
	TurnOff(deviceID string) error
}

// AlarmDevice representa um dispositivo físico ESP32
type AlarmDevice struct {
	ID        string
	IPAddress string
	Port      int
}

// AlarmRepository define o contrato para persistência de eventos de alarme
type AlarmRepository interface {
	Save(event *AlarmEvent) error
	FindByID(id string) (*AlarmEvent, error)
	FindRecent(limit int) ([]*AlarmEvent, error)
}

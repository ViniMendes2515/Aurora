package domain

import "time"

// AlarmTriggeredEvent é o domain event emitido quando um alarme é acionado.
// Pertence ao domínio pois representa algo que aconteceu dentro deste bounded context.
type AlarmTriggeredEvent struct {
	AlarmID     string    `json:"alarm_id"`
	UserID      string    `json:"user_id"`
	TriggerType string    `json:"trigger_type"`
	SensorID    string    `json:"sensor_id"`
	Location    string    `json:"location"`
	TriggeredAt time.Time `json:"triggered_at"`
}

func (e *AlarmTriggeredEvent) Topic() string {
	return "security.alarm.triggered"
}

// NewAlarmTriggeredEvent constrói um domain event a partir de um AlarmEvent persistido.
func NewAlarmTriggeredEvent(alarm *AlarmEvent) *AlarmTriggeredEvent {
	return &AlarmTriggeredEvent{
		AlarmID:     alarm.ID,
		UserID:      alarm.UserID,
		TriggerType: alarm.TriggerType.String(),
		SensorID:    alarm.SensorID,
		Location:    alarm.Location,
		TriggeredAt: alarm.TriggeredAt,
	}
}

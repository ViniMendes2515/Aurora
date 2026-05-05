package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type alarmTriggeredEvent struct {
	AlarmID     string    `json:"alarm_id"`
	UserID      string    `json:"user_id"`
	TriggerType string    `json:"trigger_type"`
	SensorID    string    `json:"sensor_id"`
	Location    string    `json:"location"`
	TriggeredAt time.Time `json:"triggered_at"`
}

// NATSAlarmPublisher publica eventos de alarme no NATS
type NATSAlarmPublisher struct {
	conn *nats.Conn
}

func NewNATSAlarmPublisher(natsURL string) (*NATSAlarmPublisher, error) {
	var nc *nats.Conn
	var err error
	for i := 0; i < 10; i++ {
		nc, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		log.Printf("[NATS] Aguardando NATS... tentativa %d/10", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}
	return &NATSAlarmPublisher{conn: nc}, nil
}

func (p *NATSAlarmPublisher) PublishAlarmTriggered(alarmID, userID, triggerType, sensorID, location string) {
	event := alarmTriggeredEvent{
		AlarmID:     alarmID,
		UserID:      userID,
		TriggerType: triggerType,
		SensorID:    sensorID,
		Location:    location,
		TriggeredAt: time.Now(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[NATS] Erro ao serializar security.alarm.triggered: %v", err)
		return
	}
	if err := p.conn.Publish("security.alarm.triggered", data); err != nil {
		log.Printf("[NATS] Erro ao publicar security.alarm.triggered: %v", err)
	}
}

func (p *NATSAlarmPublisher) Close() {
	p.conn.Close()
}

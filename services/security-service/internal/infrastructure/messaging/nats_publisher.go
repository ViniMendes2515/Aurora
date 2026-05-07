package messaging

import (
	"encoding/json"
	"log"

	"aurora/services/security-service/internal/domain"
)

// NATSAlarmPublisher implementa domain.EventPublisher usando NATS
type NATSAlarmPublisher struct {
	conn *NATSConnection
}

func NewNATSAlarmPublisher(conn *NATSConnection) *NATSAlarmPublisher {
	return &NATSAlarmPublisher{conn: conn}
}

func (p *NATSAlarmPublisher) PublishAlarmTriggered(event *domain.AlarmTriggeredEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err := p.conn.GetConnection().Publish(event.Topic(), data); err != nil {
		log.Printf("[NATS] Erro ao publicar %s: %v", event.Topic(), err)
		return err
	}
	return nil
}

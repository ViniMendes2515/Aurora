package messaging

import (
	"encoding/json"
	"log"

	"aurora/services/lighting-service/internal/domain"
)

// NATSLightPublisher implementa domain.EventPublisher usando NATS
type NATSLightPublisher struct {
	conn *NATSConnection
}

func NewNATSLightPublisher(conn *NATSConnection) *NATSLightPublisher {
	return &NATSLightPublisher{conn: conn}
}

func (p *NATSLightPublisher) PublishLightChanged(event *domain.LightStateChangedEvent) error {
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

package messaging

import (
	"encoding/json"
	"log"

	"aurora/services/sensors-service/internal/domain"
)

// NATSPublisher implementa domain.EventPublisher usando NATS
type NATSPublisher struct {
	conn *NATSConnection
}

// NewNATSPublisher cria uma nova instância de NATSPublisher
func NewNATSPublisher(conn *NATSConnection) *NATSPublisher {
	return &NATSPublisher{
		conn: conn,
	}
}

// PublishMotionEvent publica um evento de movimento no NATS
func (p *NATSPublisher) PublishMotionEvent(event *domain.MotionDetectedEvent) error {
	// Serializar evento para JSON
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publicar no tópico do evento
	topic := event.Topic()
	if err := p.conn.GetConnection().Publish(topic, data); err != nil {
		log.Printf("Failed to publish event to %s: %v", topic, err)
		return domain.ErrPublishFailed
	}

	log.Printf("Published motion event to %s: sensorID=%s, userID=%s",
		topic, event.SensorID, event.UserID)

	return nil
}

package messaging

import (
	"encoding/json"
	"log"
	"time"

	"aurora/services/sensors-service/internal/domain"
	"aurora/services/sensors-service/internal/infrastructure/ws"
)

// NATSPublisher implementa domain.EventPublisher usando NATS
// e também replica eventos em tempo real via WebSocket (Hub).
type NATSPublisher struct {
	conn *NATSConnection
	hub  *ws.Hub
}

// NewNATSPublisher cria uma nova instância de NATSPublisher
func NewNATSPublisher(conn *NATSConnection, hub *ws.Hub) *NATSPublisher {
	return &NATSPublisher{
		conn: conn,
		hub:  hub,
	}
}

// PublishMotionEvent publica um evento de movimento no NATS
func (p *NATSPublisher) PublishMotionEvent(event *domain.MotionDetectedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	topic := event.Topic()
	if err := p.conn.GetConnection().Publish(topic, data); err != nil {
		log.Printf("Failed to publish event to %s: %v", topic, err)
		return domain.ErrPublishFailed
	}

	log.Printf("Published motion event to %s: sensorID=%s, userID=%s",
		topic, event.SensorID, event.UserID)

	// Envia atualização em tempo real para WebSocket (histórico de movimento)
	if p.hub != nil {
		p.hub.Broadcast(ws.SensorEvent{
			Type:       ws.EventMotionDetected,
			SensorID:   event.SensorID,
			SensorType: string(domain.SensorTypeMotion),
			DetectedAt: event.DetectedAt.Format(time.RFC3339),
		})
	}

	return nil
}

// PublishLightEvent publica um evento de luminosidade no NATS
func (p *NATSPublisher) PublishLightEvent(event *domain.LightLevelEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	topic := event.Topic()
	if err := p.conn.GetConnection().Publish(topic, data); err != nil {
		log.Printf("Failed to publish light event to %s: %v", topic, err)
		return domain.ErrPublishFailed
	}

	log.Printf("Published light event to %s: sensorID=%s, value=%.1f%%",
		topic, event.SensorID, event.Value.Float64())

	// Envia atualização em tempo real para WebSocket (histórico de luminosidade)
	if p.hub != nil {
		p.hub.Broadcast(ws.SensorEvent{
			Type:       ws.EventLightUpdated,
			SensorID:   event.SensorID,
			SensorType: string(domain.SensorTypeLight),
			Value:      event.Value.Float64(),
			Raw:        event.Raw,
			RecordedAt: event.RecordedAt.Format(time.RFC3339),
		})
	}

	return nil
}

package messaging

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"

	"aurora/services/rules-service/internal/application"
)

// motionEvent e lightEvent são wire DTOs dos eventos externos recebidos do sensors-service.
type motionEvent struct {
	ID       string `json:"id"`
	SensorID string `json:"sensor_id"`
	Location string `json:"location"`
}

type lightEvent struct {
	ID       string  `json:"id"`
	SensorID string  `json:"sensor_id"`
	Value    float64 `json:"value"`
}

// NATSSubscriber gerencia as subscrições NATS do rules-service
type NATSSubscriber struct {
	conn        *NATSConnection
	rulesEngine *application.RulesEngine
}

// NewNATSSubscriber cria um novo subscriber NATS
func NewNATSSubscriber(conn *NATSConnection, rulesEngine *application.RulesEngine) *NATSSubscriber {
	return &NATSSubscriber{conn: conn, rulesEngine: rulesEngine}
}

// Subscribe registra todas as subscrições
func (s *NATSSubscriber) Subscribe() error {
	_, err := s.conn.GetConnection().Subscribe("sensors.motion.detected", func(msg *nats.Msg) {
		var event motionEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[NATS] Failed to unmarshal motion event: %v", err)
			return
		}
		log.Printf("[NATS] Evaluating motion rules for sensor: %s", event.SensorID)
		s.rulesEngine.EvaluateMotionTrigger(event.SensorID)
	})
	if err != nil {
		return err
	}

	// Luminosidade alterada — o engine aplica o limiar configurado em cada regra
	_, err = s.conn.GetConnection().Subscribe("sensors.light.changed", func(msg *nats.Msg) {
		var event lightEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[NATS] Failed to unmarshal light event: %v", err)
			return
		}
		log.Printf("[NATS] Light level %.1f%% from sensor %s — evaluating light rules", event.Value, event.SensorID)
		s.rulesEngine.EvaluateLightTrigger(event.SensorID, event.Value)
		s.rulesEngine.EvaluateLightHighTrigger(event.SensorID, event.Value)
	})
	if err != nil {
		return err
	}

	log.Println("[NATS] Subscribed to sensors.motion.detected and sensors.light.changed")
	return nil
}

package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"aurora/services/rules-service/internal/application"
)

// MotionEvent representa o evento de movimento do NATS
type MotionEvent struct {
	ID       string `json:"id"`
	SensorID string `json:"sensor_id"`
	Location string `json:"location"`
}

// LightEvent representa o evento de luminosidade do NATS
type LightEvent struct {
	ID       string  `json:"id"`
	SensorID string  `json:"sensor_id"`
	Value    float64 `json:"value"`
}

// NATSSubscriber gerencia as subscrições NATS do rules-service
type NATSSubscriber struct {
	conn        *nats.Conn
	rulesEngine *application.RulesEngine
}

// NewNATSSubscriber cria um novo subscriber NATS
func NewNATSSubscriber(natsURL string, rulesEngine *application.RulesEngine) (*NATSSubscriber, error) {
	var nc *nats.Conn
	var err error

	for i := 0; i < 10; i++ {
		nc, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		log.Printf("[NATS] Waiting for NATS... attempt %d/10", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	return &NATSSubscriber{conn: nc, rulesEngine: rulesEngine}, nil
}

// Subscribe registra todas as subscrições
func (s *NATSSubscriber) Subscribe() error {
	// Movimento detectado
	_, err := s.conn.Subscribe("sensors.motion.detected", func(msg *nats.Msg) {
		var event MotionEvent
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
	_, err = s.conn.Subscribe("sensors.light.changed", func(msg *nats.Msg) {
		var event LightEvent
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

// Close fecha a conexão NATS
func (s *NATSSubscriber) Close() {
	s.conn.Close()
}

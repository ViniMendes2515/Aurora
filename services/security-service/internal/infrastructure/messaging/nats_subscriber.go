package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"aurora/services/security-service/internal/application"
)

// MotionEvent representa o evento recebido do NATS
type MotionEvent struct {
	ID         string    `json:"id"`
	SensorID   string    `json:"sensor_id"`
	UserID     string    `json:"user_id"`
	Location   string    `json:"location"`
	DetectedAt time.Time `json:"detected_at"`
}

// NATSSubscriber gerencia as subscrições NATS do security-service
type NATSSubscriber struct {
	conn         *nats.Conn
	alarmService *application.AlarmService
	autoAlarm    bool
}

// NewNATSSubscriber cria um novo subscriber NATS
func NewNATSSubscriber(natsURL string, alarmService *application.AlarmService, autoAlarm bool) (*NATSSubscriber, error) {
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

	return &NATSSubscriber{conn: nc, alarmService: alarmService, autoAlarm: autoAlarm}, nil
}

// Subscribe registra todas as subscrições
func (s *NATSSubscriber) Subscribe() error {
	_, err := s.conn.Subscribe("sensors.motion.detected", func(msg *nats.Msg) {
		var event MotionEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[NATS] Failed to unmarshal motion event: %v", err)
			return
		}

		if !s.autoAlarm {
			log.Printf("[NATS] Motion detected at %s (sensor: %s) — auto-alarm disabled, ignoring", event.Location, event.SensorID)
			return
		}

		log.Printf("[NATS] Motion detected at %s (sensor: %s) — triggering alarm", event.Location, event.SensorID)

		if _, err := s.alarmService.TriggerAlarm("motion", event.SensorID, event.Location); err != nil {
			log.Printf("[NATS] Failed to trigger alarm: %v", err)
		}
	})

	if err != nil {
		return err
	}

	log.Println("[NATS] Subscribed to sensors.motion.detected")
	return nil
}

// Close fecha a conexão NATS
func (s *NATSSubscriber) Close() {
	s.conn.Close()
}

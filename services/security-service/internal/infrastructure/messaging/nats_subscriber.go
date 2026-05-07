package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"aurora/services/security-service/internal/application"
)

// motionEvent é o wire DTO do evento externo recebido do sensors-service.
// Permanece na infraestrutura pois é um contrato de transporte de outro bounded context.
type motionEvent struct {
	ID         string    `json:"id"`
	SensorID   string    `json:"sensor_id"`
	UserID     string    `json:"user_id"`
	Location   string    `json:"location"`
	DetectedAt time.Time `json:"detected_at"`
}

// NATSSubscriber gerencia as subscrições NATS do security-service
type NATSSubscriber struct {
	conn         *NATSConnection
	alarmService *application.AlarmService
	autoAlarm    bool
}

// NewNATSSubscriber cria um novo subscriber NATS
func NewNATSSubscriber(conn *NATSConnection, alarmService *application.AlarmService, autoAlarm bool) *NATSSubscriber {
	return &NATSSubscriber{conn: conn, alarmService: alarmService, autoAlarm: autoAlarm}
}

// Subscribe registra todas as subscrições
func (s *NATSSubscriber) Subscribe() error {
	_, err := s.conn.GetConnection().Subscribe("sensors.motion.detected", func(msg *nats.Msg) {
		var event motionEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[NATS] Failed to unmarshal motion event: %v", err)
			return
		}

		if !s.autoAlarm {
			log.Printf("[NATS] Motion detected at %s (sensor: %s) — auto-alarm disabled, ignoring", event.Location, event.SensorID)
			return
		}

		log.Printf("[NATS] Motion detected at %s (sensor: %s) — triggering alarm", event.Location, event.SensorID)

		if _, err := s.alarmService.TriggerAlarm(event.UserID, "motion", event.SensorID, event.Location); err != nil {
			log.Printf("[NATS] Failed to trigger alarm: %v", err)
		}
	})

	if err != nil {
		return err
	}

	log.Println("[NATS] Subscribed to sensors.motion.detected")
	return nil
}

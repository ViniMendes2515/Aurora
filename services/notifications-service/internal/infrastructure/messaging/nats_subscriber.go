package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"aurora/services/notifications-service/internal/application"
)

// motionEvent representa o payload publicado em sensors.motion.detected
type motionEvent struct {
	ID         string    `json:"id"`
	SensorID   string    `json:"sensor_id"`
	UserID     string    `json:"user_id"`
	Location   string    `json:"location"`
	DetectedAt time.Time `json:"detected_at"`
}

// lightEvent representa o payload publicado em sensors.light.changed
type lightEvent struct {
	ID         string    `json:"id"`
	SensorID   string    `json:"sensor_id"`
	Value      float64   `json:"value"`
	Raw        int       `json:"raw"`
	RecordedAt time.Time `json:"recorded_at"`
}

// NATSSubscriber gerencia as subscricoes NATS do notifications-service
type NATSSubscriber struct {
	conn    *nats.Conn
	service *application.NotificationService
}

// NewNATSSubscriber cria e conecta o subscriber ao servidor NATS com retry automatico
func NewNATSSubscriber(natsURL string, service *application.NotificationService) (*NATSSubscriber, error) {
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

	return &NATSSubscriber{conn: nc, service: service}, nil
}

// Subscribe registra as subscricoes nos topicos de eventos dos sensores
func (s *NATSSubscriber) Subscribe() error {
	if _, err := s.conn.Subscribe("sensors.motion.detected", s.handleMotion); err != nil {
		return err
	}
	log.Println("[NATS] Subscrito em sensors.motion.detected")

	if _, err := s.conn.Subscribe("sensors.light.changed", s.handleLight); err != nil {
		return err
	}
	log.Println("[NATS] Subscrito em sensors.light.changed")

	return nil
}

func (s *NATSSubscriber) handleMotion(msg *nats.Msg) {
	var e motionEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil {
		log.Printf("[NATS] Erro ao parsear sensors.motion.detected: %v", err)
		return
	}
	if err := s.service.HandleMotionDetected(e.SensorID, e.UserID, e.Location); err != nil {
		log.Printf("[NATS] Erro ao salvar notificacao de movimento: %v", err)
	}
}

func (s *NATSSubscriber) handleLight(msg *nats.Msg) {
	var e lightEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil {
		log.Printf("[NATS] Erro ao parsear sensors.light.changed: %v", err)
		return
	}
	if err := s.service.HandleLightLow(e.SensorID, e.Value); err != nil {
		log.Printf("[NATS] Erro ao salvar notificacao de luz baixa: %v", err)
	}
}

// Close fecha a conexao NATS de forma limpa
func (s *NATSSubscriber) Close() {
	s.conn.Close()
}

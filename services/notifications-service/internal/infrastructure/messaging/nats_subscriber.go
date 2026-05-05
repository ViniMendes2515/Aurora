package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"aurora/services/notifications-service/internal/application"
)

type motionEvent struct {
	ID         string    `json:"id"`
	SensorID   string    `json:"sensor_id"`
	UserID     string    `json:"user_id"`
	Location   string    `json:"location"`
	DetectedAt time.Time `json:"detected_at"`
}

type lightSensorEvent struct {
	ID         string    `json:"id"`
	SensorID   string    `json:"sensor_id"`
	Value      float64   `json:"value"`
	Raw        int       `json:"raw"`
	RecordedAt time.Time `json:"recorded_at"`
}

type lightStateEvent struct {
	LightID   string    `json:"light_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	State     string    `json:"state"`
	ChangedAt time.Time `json:"changed_at"`
}

type alarmTriggeredEvent struct {
	AlarmID     string    `json:"alarm_id"`
	UserID      string    `json:"user_id"`
	TriggerType string    `json:"trigger_type"`
	SensorID    string    `json:"sensor_id"`
	Location    string    `json:"location"`
	TriggeredAt time.Time `json:"triggered_at"`
}

type scheduleExecutedEvent struct {
	ScheduleID     string `json:"schedule_id"`
	ScheduleName   string `json:"schedule_name"`
	OwnerID        string `json:"owner_id"`
	ActionType     string `json:"action_type"`
	ActionDeviceID string `json:"action_device_id"`
	ExecutedAt     string `json:"executed_at"`
}

// NATSSubscriber gerencia as subscricoes NATS do notifications-service
type NATSSubscriber struct {
	conn    *nats.Conn
	service *application.NotificationService
}

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

func (s *NATSSubscriber) Subscribe() error {
	subs := []struct {
		subject string
		handler nats.MsgHandler
	}{
		{"sensors.motion.detected", s.handleMotion},
		{"sensors.light.changed", s.handleLightSensor},
		{"lighting.state.changed", s.handleLightState},
		{"security.alarm.triggered", s.handleAlarm},
		{"schedule.executed", s.handleScheduleExecuted},
	}

	for _, sub := range subs {
		if _, err := s.conn.Subscribe(sub.subject, sub.handler); err != nil {
			return err
		}
		log.Printf("[NATS] Subscrito em %s", sub.subject)
	}
	return nil
}

func (s *NATSSubscriber) handleMotion(msg *nats.Msg) {
	var e motionEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil {
		log.Printf("[NATS] Erro ao parsear sensors.motion.detected: %v", err)
		return
	}
	if err := s.service.HandleMotionDetected(e.SensorID, e.UserID, e.Location); err != nil {
		log.Printf("[NATS] Erro ao processar notificacao de movimento: %v", err)
	}
}

func (s *NATSSubscriber) handleLightSensor(msg *nats.Msg) {
	var e lightSensorEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil {
		log.Printf("[NATS] Erro ao parsear sensors.light.changed: %v", err)
		return
	}
	if err := s.service.HandleLightLow(e.SensorID, e.Value); err != nil {
		log.Printf("[NATS] Erro ao processar notificacao de luz baixa: %v", err)
	}
}

func (s *NATSSubscriber) handleLightState(msg *nats.Msg) {
	var e lightStateEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil {
		log.Printf("[NATS] Erro ao parsear lighting.state.changed: %v", err)
		return
	}
	isOn := e.State == "on"
	if err := s.service.HandleLightChanged(e.LightID, e.UserID, e.Location, isOn); err != nil {
		log.Printf("[NATS] Erro ao processar notificacao de luz: %v", err)
	}
}

func (s *NATSSubscriber) handleAlarm(msg *nats.Msg) {
	var e alarmTriggeredEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil {
		log.Printf("[NATS] Erro ao parsear security.alarm.triggered: %v", err)
		return
	}
	if err := s.service.HandleAlarmTriggered(e.SensorID, e.UserID, e.Location); err != nil {
		log.Printf("[NATS] Erro ao processar notificacao de alarme: %v", err)
	}
}

func (s *NATSSubscriber) handleScheduleExecuted(msg *nats.Msg) {
	log.Printf("[NATS] schedule.executed recebido: %s", string(msg.Data))
	var e scheduleExecutedEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil {
		log.Printf("[NATS] Erro ao parsear schedule.executed: %v", err)
		return
	}
	if err := s.service.HandleScheduleExecuted(e.ScheduleID, e.OwnerID, e.ActionType, e.ScheduleName); err != nil {
		log.Printf("[NATS] Erro ao processar notificacao de agendamento: %v", err)
	}
}

func (s *NATSSubscriber) Close() {
	s.conn.Close()
}

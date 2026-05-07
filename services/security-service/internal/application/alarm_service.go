package application

import (
	"log"
	"time"

	"aurora/services/security-service/internal/domain"
)

// AlarmService implementa os casos de uso de segurança
type AlarmService struct {
	alarmRepo      domain.AlarmRepository
	buzzerClient   domain.BuzzerClient
	buzzerDuration time.Duration
	deviceID       string
	eventPublisher domain.EventPublisher
}

// NewAlarmService cria uma nova instância de AlarmService
func NewAlarmService(alarmRepo domain.AlarmRepository, buzzerClient domain.BuzzerClient, buzzerDurationMs int, deviceID string, eventPublisher domain.EventPublisher) *AlarmService {
	return &AlarmService{
		alarmRepo:      alarmRepo,
		buzzerClient:   buzzerClient,
		buzzerDuration: time.Duration(buzzerDurationMs) * time.Millisecond,
		deviceID:       deviceID,
		eventPublisher: eventPublisher,
	}
}

// AlarmResponse representa a resposta de um alarme
type AlarmResponse struct {
	ID          string             `json:"id"`
	TriggerType string             `json:"trigger_type"`
	SensorID    string             `json:"sensor_id"`
	Location    string             `json:"location"`
	Status      domain.AlarmStatus `json:"status"`
	TriggeredAt time.Time          `json:"triggered_at"`
}

func toResponse(a *domain.AlarmEvent) *AlarmResponse {
	return &AlarmResponse{
		ID:          a.ID,
		TriggerType: a.TriggerType.String(),
		SensorID:    a.SensorID,
		Location:    a.Location,
		Status:      a.Status,
		TriggeredAt: a.TriggeredAt,
	}
}

// TriggerAlarm aciona o alarme (chamado pelo subscriber NATS ou por API)
func (s *AlarmService) TriggerAlarm(userID, triggerType, sensorID, location string) (*AlarmResponse, error) {
	event, err := domain.NewAlarmEvent(userID, triggerType, sensorID, location)
	if err != nil {
		return nil, err
	}

	if err := s.alarmRepo.Save(event); err != nil {
		return nil, err
	}

	if s.eventPublisher != nil {
		domainEvent := domain.NewAlarmTriggeredEvent(event)
		if err := s.eventPublisher.PublishAlarmTriggered(domainEvent); err != nil {
			log.Printf("[Event] Failed to publish alarm triggered: %v", err)
		}
	}

	go func() {
		log.Printf("[Alarm] Acionando buzzer por %v (trigger: %s, sensor: %s)", s.buzzerDuration, triggerType, sensorID)
		if err := s.buzzerClient.TurnOn(s.deviceID); err != nil {
			log.Printf("[Alarm] Falha ao ligar buzzer: %v", err)
			return
		}
		time.Sleep(s.buzzerDuration)
		if err := s.buzzerClient.TurnOff(s.deviceID); err != nil {
			log.Printf("[Alarm] Falha ao desligar buzzer: %v", err)
		}
	}()

	return toResponse(event), nil
}

// SilenceAlarm silencia um alarme ativo
func (s *AlarmService) SilenceAlarm(alarmID string) (*AlarmResponse, error) {
	event, err := s.alarmRepo.FindByID(alarmID)
	if err != nil {
		return nil, domain.ErrAlarmNotFound
	}

	event.Silence()
	if err := s.alarmRepo.Save(event); err != nil {
		return nil, err
	}

	if err := s.buzzerClient.TurnOff(s.deviceID); err != nil {
		log.Printf("[Alarm] Falha ao desligar buzzer: %v", err)
	}

	return toResponse(event), nil
}

// GetRecentAlarms retorna os últimos eventos de alarme
func (s *AlarmService) GetRecentAlarms(limit int) ([]*AlarmResponse, error) {
	events, err := s.alarmRepo.FindRecent(limit)
	if err != nil {
		return nil, err
	}

	var responses []*AlarmResponse
	for _, e := range events {
		responses = append(responses, toResponse(e))
	}
	return responses, nil
}

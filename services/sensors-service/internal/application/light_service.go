package application

import (
	"aurora/services/sensors-service/internal/domain"
)

// LightService implementa os casos de uso relacionados a luminosidade
type LightService struct {
	sensorRepo     domain.SensorRepository
	eventPublisher domain.EventPublisher
}

// NewLightService cria uma nova instância de LightService
func NewLightService(sensorRepo domain.SensorRepository, eventPublisher domain.EventPublisher) *LightService {
	return &LightService{
		sensorRepo:     sensorRepo,
		eventPublisher: eventPublisher,
	}
}

// RegisterLightLevelRequest representa os dados de entrada para registro de luminosidade
type RegisterLightLevelRequest struct {
	SensorID string
	Value    float64
	Raw      int
}

// RegisterLightLevelResponse representa a resposta do registro de luminosidade
type RegisterLightLevelResponse struct {
	Status   string  `json:"status"`
	EventID  string  `json:"event_id,omitempty"`
	Value    float64 `json:"value"`
}

// RegisterLightLevel executa o caso de uso de registro de luminosidade
func (s *LightService) RegisterLightLevel(req RegisterLightLevelRequest) (*RegisterLightLevelResponse, error) {
	if req.SensorID == "" {
		return nil, domain.ErrInvalidSensorID
	}

	_, err := s.sensorRepo.FindByID(req.SensorID)
	if err != nil {
		return nil, domain.ErrSensorNotFound
	}

	record, err := domain.NewLightRecord(req.SensorID, req.Value, req.Raw)
	if err != nil {
		return nil, err
	}
	if err := s.sensorRepo.SaveLightRecord(record); err != nil {
		return nil, err
	}

	event, err := domain.NewLightLevelEvent(req.SensorID, req.Value, req.Raw)
	if err != nil {
		return nil, err
	}
	if err := s.eventPublisher.PublishLightEvent(event); err != nil {
		// Falha na publicação não cancela o registro
		_ = err
	}

	return &RegisterLightLevelResponse{
		Status:  "light level registered",
		EventID: event.ID,
		Value:   req.Value,
	}, nil
}

// GetLightRecords retorna os últimos registros de luminosidade de um sensor
func (s *LightService) GetLightRecords(sensorID string, limit int) ([]*domain.LightRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.sensorRepo.GetLightRecords(sensorID, limit)
}

package application

import (
	"aurora/services/sensors-service/internal/domain"
)

// MotionService implementa os casos de uso relacionados a movimento
type MotionService struct {
	sensorRepo     domain.SensorRepository
	eventPublisher domain.EventPublisher
}

// NewMotionService cria uma nova instância de MotionService
func NewMotionService(sensorRepo domain.SensorRepository, eventPublisher domain.EventPublisher) *MotionService {
	return &MotionService{
		sensorRepo:     sensorRepo,
		eventPublisher: eventPublisher,
	}
}

// RegisterMotionRequest representa os dados de entrada para registro de movimento
type RegisterMotionRequest struct {
	SensorID string
	UserID   string
}

// RegisterMotionResponse representa a resposta do registro de movimento
type RegisterMotionResponse struct {
	Status  string `json:"status"`
	EventID string `json:"event_id,omitempty"`
}

// RegisterMotion executa o caso de uso de registro de movimento
func (s *MotionService) RegisterMotion(req RegisterMotionRequest) (*RegisterMotionResponse, error) {
	// Validar sensor ID
	if req.SensorID == "" {
		return nil, domain.ErrInvalidSensorID
	}

	// Buscar sensor no repositório
	sensor, err := s.sensorRepo.FindByID(req.SensorID)
	if err != nil {
		return nil, domain.ErrSensorNotFound
	}

	// Verificar se o sensor pertence ao usuário
	if !sensor.BelongsTo(req.UserID) {
		return nil, domain.ErrSensorAccessDenied
	}

	// Criar evento de movimento no domínio
	event := domain.NewMotionDetectedEvent(req.SensorID, req.UserID, sensor.Location)

	// Registrar ação no repositório (in-memory)
	record := domain.NewMotionRecord(req.SensorID, req.UserID)
	if err := s.sensorRepo.SaveMotionRecord(record); err != nil {
		return nil, err
	}

	// Publicar evento no NATS (infra/messaging)
	if err := s.eventPublisher.PublishMotionEvent(event); err != nil {
		// Log do erro mas não falha a operação
		// O registro já foi salvo
	}

	return &RegisterMotionResponse{
		Status:  "motion registered",
		EventID: event.ID,
	}, nil
}

// GetSensorByID retorna um sensor pelo ID (para verificações)
func (s *MotionService) GetSensorByID(sensorID, userID string) (*domain.Sensor, error) {
	sensor, err := s.sensorRepo.FindByID(sensorID)
	if err != nil {
		return nil, domain.ErrSensorNotFound
	}

	if !sensor.BelongsTo(userID) {
		return nil, domain.ErrSensorAccessDenied
	}

	return sensor, nil
}

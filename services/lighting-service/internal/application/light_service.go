package application

import (
	"aurora/services/lighting-service/internal/domain"
)

// StatePublisher é a interface que o LightService usa para notificar mudanças de estado
// (implementada pelo ws.Hub)
type StatePublisher interface {
	BroadcastState(lightID, name, location, state string)
}

// LightService implementa os casos de uso de iluminação
type LightService struct {
	lightRepo    domain.LightRepository
	deviceClient domain.DeviceClient
	publisher    StatePublisher
}

// NewLightService cria uma nova instância de LightService
func NewLightService(lightRepo domain.LightRepository, deviceClient domain.DeviceClient, publisher StatePublisher) *LightService {
	return &LightService{
		lightRepo:    lightRepo,
		deviceClient: deviceClient,
		publisher:    publisher,
	}
}

// LightResponse representa a resposta de um comando de luz
type LightResponse struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Location string           `json:"location"`
	State    domain.LightState `json:"state"`
}

func toResponse(l *domain.Light) *LightResponse {
	return &LightResponse{
		ID:       l.ID,
		Name:     l.Name,
		Location: l.Location,
		State:    l.State,
	}
}

// TurnOn acende uma luz
func (s *LightService) TurnOn(lightID, userID string) (*LightResponse, error) {
	light, err := s.lightRepo.FindByID(lightID)
	if err != nil {
		return nil, domain.ErrLightNotFound
	}

	if !light.BelongsTo(userID) {
		return nil, domain.ErrLightAccessDenied
	}

	if err := s.deviceClient.TurnOn(light.DeviceID); err != nil {
		return nil, domain.ErrDeviceUnreachable
	}

	light.TurnOn()
	if err := s.lightRepo.Save(light); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		s.publisher.BroadcastState(light.ID, light.Name, light.Location, string(light.State))
	}

	return toResponse(light), nil
}

// TurnOff apaga uma luz
func (s *LightService) TurnOff(lightID, userID string) (*LightResponse, error) {
	light, err := s.lightRepo.FindByID(lightID)
	if err != nil {
		return nil, domain.ErrLightNotFound
	}

	if !light.BelongsTo(userID) {
		return nil, domain.ErrLightAccessDenied
	}

	if err := s.deviceClient.TurnOff(light.DeviceID); err != nil {
		return nil, domain.ErrDeviceUnreachable
	}

	light.TurnOff()
	if err := s.lightRepo.Save(light); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		s.publisher.BroadcastState(light.ID, light.Name, light.Location, string(light.State))
	}

	return toResponse(light), nil
}

// GetStatus retorna o estado de uma luz
func (s *LightService) GetStatus(lightID, userID string) (*LightResponse, error) {
	light, err := s.lightRepo.FindByID(lightID)
	if err != nil {
		return nil, domain.ErrLightNotFound
	}

	if !light.BelongsTo(userID) {
		return nil, domain.ErrLightAccessDenied
	}

	return toResponse(light), nil
}

// ListLights retorna todas as luzes de um usuário
func (s *LightService) ListLights(userID string) ([]*LightResponse, error) {
	lights, err := s.lightRepo.FindByOwnerID(userID)
	if err != nil {
		return nil, err
	}

	var responses []*LightResponse
	for _, l := range lights {
		responses = append(responses, toResponse(l))
	}
	return responses, nil
}

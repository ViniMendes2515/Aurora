package application_test

import (
	"errors"
	"testing"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/domain"
)

// --- Mocks ---

// mockSensorRepoLight é um repositório de sensores falso para testes do LightService
type mockSensorRepoLight struct {
	sensor         *domain.Sensor
	findByIDErr    error
	saveLightErr   error
	lightRecords   []*domain.LightRecord
	savedLightRec  *domain.LightRecord
	// capturedLimit guarda o limit passado para GetLightRecords
	capturedLimit  int
}

func (m *mockSensorRepoLight) FindByID(id string) (*domain.Sensor, error) {
	return m.sensor, m.findByIDErr
}

func (m *mockSensorRepoLight) FindByOwnerID(ownerID string) ([]*domain.Sensor, error) {
	return nil, nil
}

func (m *mockSensorRepoLight) FindAll() ([]*domain.Sensor, error) {
	return nil, nil
}

func (m *mockSensorRepoLight) Save(sensor *domain.Sensor) error {
	return nil
}

func (m *mockSensorRepoLight) SaveMotionRecord(record *domain.MotionRecord) error {
	return nil
}

func (m *mockSensorRepoLight) GetMotionRecords(sensorID string, limit int) ([]*domain.MotionRecord, error) {
	return nil, nil
}

func (m *mockSensorRepoLight) SaveLightRecord(record *domain.LightRecord) error {
	m.savedLightRec = record
	return m.saveLightErr
}

func (m *mockSensorRepoLight) GetLightRecords(sensorID string, limit int) ([]*domain.LightRecord, error) {
	m.capturedLimit = limit
	return m.lightRecords, nil
}

// mockEventPublisherLight é um publicador de eventos falso para testes do LightService
type mockEventPublisherLight struct {
	publishLightErr    error
	publishedLightEvt  *domain.LightLevelEvent
}

func (m *mockEventPublisherLight) PublishMotionEvent(event *domain.MotionDetectedEvent) error {
	return nil
}

func (m *mockEventPublisherLight) PublishLightEvent(event *domain.LightLevelEvent) error {
	m.publishedLightEvt = event
	return m.publishLightErr
}

// --- Testes ---

// TestRegisterLightLevel_SensorIDVazio verifica que um SensorID vazio retorna ErrInvalidSensorID
func TestRegisterLightLevel_SensorIDVazio(t *testing.T) {
	repo := &mockSensorRepoLight{}
	pub := &mockEventPublisherLight{}
	svc := application.NewLightService(repo, pub)

	req := application.RegisterLightLevelRequest{
		SensorID: "",
		Value:    50.0,
		Raw:      2048,
	}

	_, err := svc.RegisterLightLevel(req)
	if !errors.Is(err, domain.ErrInvalidSensorID) {
		t.Errorf("esperava ErrInvalidSensorID, mas recebeu: %v", err)
	}
}

// TestRegisterLightLevel_SensorNaoEncontrado verifica que erro no repositório retorna ErrSensorNotFound
func TestRegisterLightLevel_SensorNaoEncontrado(t *testing.T) {
	repo := &mockSensorRepoLight{
		findByIDErr: errors.New("sensor não existe no banco"),
	}
	pub := &mockEventPublisherLight{}
	svc := application.NewLightService(repo, pub)

	req := application.RegisterLightLevelRequest{
		SensorID: "sensor-inexistente",
		Value:    50.0,
		Raw:      2048,
	}

	_, err := svc.RegisterLightLevel(req)
	if !errors.Is(err, domain.ErrSensorNotFound) {
		t.Errorf("esperava ErrSensorNotFound, mas recebeu: %v", err)
	}
}

// TestRegisterLightLevel_ValorInvalido verifica que um valor acima de 100 retorna erro de validação
func TestRegisterLightLevel_ValorInvalido(t *testing.T) {
	sensor := domain.NewSensor("sensor-luz-1", "Sensor Luz Sala", domain.SensorTypeLight, "Sala", "usuario-1")
	repo := &mockSensorRepoLight{sensor: sensor}
	pub := &mockEventPublisherLight{}
	svc := application.NewLightService(repo, pub)

	req := application.RegisterLightLevelRequest{
		SensorID: "sensor-luz-1",
		Value:    150.0, // valor inválido: acima de 100
		Raw:      4095,
	}

	_, err := svc.RegisterLightLevel(req)
	if err == nil {
		t.Fatal("esperava erro para valor fora do intervalo, mas não recebeu nenhum")
	}
	// O erro deve ser o de percentual inválido do domínio
	if !errors.Is(err, domain.ErrInvalidLightPercentage) {
		t.Errorf("esperava ErrInvalidLightPercentage, mas recebeu: %v", err)
	}
}

// TestRegisterLightLevel_Sucesso verifica o caminho feliz: registro salvo, evento publicado e valor na resposta
func TestRegisterLightLevel_Sucesso(t *testing.T) {
	sensor := domain.NewSensor("sensor-luz-2", "Sensor Luz Quarto", domain.SensorTypeLight, "Quarto", "usuario-1")
	repo := &mockSensorRepoLight{sensor: sensor}
	pub := &mockEventPublisherLight{}
	svc := application.NewLightService(repo, pub)

	req := application.RegisterLightLevelRequest{
		SensorID: "sensor-luz-2",
		Value:    75.5,
		Raw:      3090,
	}

	resp, err := svc.RegisterLightLevel(req)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
	if resp == nil {
		t.Fatal("resposta não deveria ser nil")
	}
	if resp.EventID == "" {
		t.Error("EventID não deveria estar vazio após registro bem-sucedido")
	}
	if resp.Value != 75.5 {
		t.Errorf("esperava Value 75.5, mas recebeu: %f", resp.Value)
	}
	if repo.savedLightRec == nil {
		t.Error("registro de luminosidade deveria ter sido salvo no repositório")
	}
	if pub.publishedLightEvt == nil {
		t.Error("evento de luminosidade deveria ter sido publicado")
	}
}

// TestGetLightRecords_LimitePadrao verifica que limit=0 é tratado como 20 ao consultar o repositório
func TestGetLightRecords_LimitePadrao(t *testing.T) {
	repo := &mockSensorRepoLight{
		lightRecords: []*domain.LightRecord{},
	}
	pub := &mockEventPublisherLight{}
	svc := application.NewLightService(repo, pub)

	// Passa limit 0, que deve ser normalizado para 20 pelo serviço
	_, err := svc.GetLightRecords("sensor-luz-3", 0)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
	if repo.capturedLimit != 20 {
		t.Errorf("esperava que o limit passado ao repositório fosse 20, mas foi: %d", repo.capturedLimit)
	}
}

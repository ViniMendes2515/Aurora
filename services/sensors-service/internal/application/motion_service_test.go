package application_test

import (
	"errors"
	"testing"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/domain"
)

// --- Mocks ---

// mockSensorRepoMotion é um repositório de sensores falso para testes do MotionService
type mockSensorRepoMotion struct {
	sensor          *domain.Sensor
	findByIDErr     error
	saveMotionErr   error
	savedMotionRec  *domain.MotionRecord
}

func (m *mockSensorRepoMotion) FindByID(id string) (*domain.Sensor, error) {
	return m.sensor, m.findByIDErr
}

func (m *mockSensorRepoMotion) FindByOwnerID(ownerID string) ([]*domain.Sensor, error) {
	return nil, nil
}

func (m *mockSensorRepoMotion) FindAll() ([]*domain.Sensor, error) {
	return nil, nil
}

func (m *mockSensorRepoMotion) Save(sensor *domain.Sensor) error {
	return nil
}

func (m *mockSensorRepoMotion) SaveMotionRecord(record *domain.MotionRecord) error {
	m.savedMotionRec = record
	return m.saveMotionErr
}

func (m *mockSensorRepoMotion) GetMotionRecords(sensorID string, limit int) ([]*domain.MotionRecord, error) {
	return nil, nil
}

func (m *mockSensorRepoMotion) SaveLightRecord(record *domain.LightRecord) error {
	return nil
}

func (m *mockSensorRepoMotion) GetLightRecords(sensorID string, limit int) ([]*domain.LightRecord, error) {
	return nil, nil
}

// mockEventPublisherMotion é um publicador de eventos falso para testes do MotionService
type mockEventPublisherMotion struct {
	publishMotionErr    error
	publishedMotionEvt  *domain.MotionDetectedEvent
}

func (m *mockEventPublisherMotion) PublishMotionEvent(event *domain.MotionDetectedEvent) error {
	m.publishedMotionEvt = event
	return m.publishMotionErr
}

func (m *mockEventPublisherMotion) PublishLightEvent(event *domain.LightLevelEvent) error {
	return nil
}

// --- Testes ---

// TestRegisterMotion_SensorIDVazio verifica que um SensorID vazio retorna ErrInvalidSensorID
func TestRegisterMotion_SensorIDVazio(t *testing.T) {
	repo := &mockSensorRepoMotion{}
	pub := &mockEventPublisherMotion{}
	svc := application.NewMotionService(repo, pub)

	req := application.RegisterMotionRequest{
		SensorID: "",
		UserID:   "usuario-1",
	}

	_, err := svc.RegisterMotion(req)
	if !errors.Is(err, domain.ErrInvalidSensorID) {
		t.Errorf("esperava ErrInvalidSensorID, mas recebeu: %v", err)
	}
}

// TestRegisterMotion_SensorNaoEncontrado verifica que um erro do repositório retorna ErrSensorNotFound
func TestRegisterMotion_SensorNaoEncontrado(t *testing.T) {
	repo := &mockSensorRepoMotion{
		findByIDErr: errors.New("sensor não existe no banco"),
	}
	pub := &mockEventPublisherMotion{}
	svc := application.NewMotionService(repo, pub)

	req := application.RegisterMotionRequest{
		SensorID: "sensor-inexistente",
		UserID:   "usuario-1",
	}

	_, err := svc.RegisterMotion(req)
	if !errors.Is(err, domain.ErrSensorNotFound) {
		t.Errorf("esperava ErrSensorNotFound, mas recebeu: %v", err)
	}
}

// TestRegisterMotion_AcessoNegado verifica que um sensor de outro usuário retorna ErrSensorAccessDenied
func TestRegisterMotion_AcessoNegado(t *testing.T) {
	// Sensor pertence ao "outro-usuario", não ao "usuario-1"
	sensor := domain.NewSensor("sensor-1", "Sensor Sala", domain.SensorTypeMotion, "Sala", "outro-usuario")
	repo := &mockSensorRepoMotion{sensor: sensor}
	pub := &mockEventPublisherMotion{}
	svc := application.NewMotionService(repo, pub)

	req := application.RegisterMotionRequest{
		SensorID: "sensor-1",
		UserID:   "usuario-1",
	}

	_, err := svc.RegisterMotion(req)
	if !errors.Is(err, domain.ErrSensorAccessDenied) {
		t.Errorf("esperava ErrSensorAccessDenied, mas recebeu: %v", err)
	}
}

// TestRegisterMotion_Sucesso verifica o caminho feliz: registro salvo, evento publicado e EventID preenchido
func TestRegisterMotion_Sucesso(t *testing.T) {
	sensor := domain.NewSensor("sensor-1", "Sensor Sala", domain.SensorTypeMotion, "Sala", "usuario-1")
	repo := &mockSensorRepoMotion{sensor: sensor}
	pub := &mockEventPublisherMotion{}
	svc := application.NewMotionService(repo, pub)

	req := application.RegisterMotionRequest{
		SensorID: "sensor-1",
		UserID:   "usuario-1",
	}

	resp, err := svc.RegisterMotion(req)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
	if resp == nil {
		t.Fatal("resposta não deveria ser nil")
	}
	if resp.EventID == "" {
		t.Error("EventID não deveria estar vazio após registro bem-sucedido")
	}
	if repo.savedMotionRec == nil {
		t.Error("registro de movimento deveria ter sido salvo no repositório")
	}
	if pub.publishedMotionEvt == nil {
		t.Error("evento de movimento deveria ter sido publicado")
	}
}

// TestRegisterDeviceMotion_Sucesso verifica que o dispositivo pode registrar movimento sem checagem de ownership
func TestRegisterDeviceMotion_Sucesso(t *testing.T) {
	// Sensor pertence a um usuário específico, mas o dispositivo não passa por checagem de ownership
	sensor := domain.NewSensor("sensor-2", "Sensor Corredor", domain.SensorTypeMotion, "Corredor", "usuario-99")
	repo := &mockSensorRepoMotion{sensor: sensor}
	pub := &mockEventPublisherMotion{}
	svc := application.NewMotionService(repo, pub)

	resp, err := svc.RegisterDeviceMotion("sensor-2")
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
	if resp == nil {
		t.Fatal("resposta não deveria ser nil")
	}
	if resp.EventID == "" {
		t.Error("EventID não deveria estar vazio após registro bem-sucedido")
	}
	// Verifica que o userID usado foi "device"
	if repo.savedMotionRec == nil {
		t.Fatal("registro de movimento deveria ter sido salvo no repositório")
	}
	if repo.savedMotionRec.UserID != "device" {
		t.Errorf("esperava UserID 'device', mas recebeu: %s", repo.savedMotionRec.UserID)
	}
}

// TestGetSensorByID_Sucesso verifica que o sensor é encontrado e retornado corretamente
func TestGetSensorByID_Sucesso(t *testing.T) {
	sensor := domain.NewSensor("sensor-3", "Sensor Quarto", domain.SensorTypeMotion, "Quarto", "usuario-2")
	repo := &mockSensorRepoMotion{sensor: sensor}
	pub := &mockEventPublisherMotion{}
	svc := application.NewMotionService(repo, pub)

	resultado, err := svc.GetSensorByID("sensor-3", "usuario-2")
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
	if resultado == nil {
		t.Fatal("sensor retornado não deveria ser nil")
	}
	if resultado.ID != "sensor-3" {
		t.Errorf("esperava ID 'sensor-3', mas recebeu: %s", resultado.ID)
	}
}

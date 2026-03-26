package application_test

import (
	"errors"
	"testing"

	"aurora/services/lighting-service/internal/application"
	"aurora/services/lighting-service/internal/domain"
)

// ---------------------------------------------------------------------------
// Mocks manuais
// ---------------------------------------------------------------------------

// mockLightRepository simula o repositório de luzes em memória.
// O campo `saved` captura a última luz persistida para validar o estado salvo.
type mockLightRepository struct {
	lights    map[string]*domain.Light
	ownerMap  map[string][]*domain.Light
	saved     *domain.Light
	findErr   error
	saveErr   error
}

func newMockRepo() *mockLightRepository {
	return &mockLightRepository{
		lights:   make(map[string]*domain.Light),
		ownerMap: make(map[string][]*domain.Light),
	}
}

func (m *mockLightRepository) FindByID(id string) (*domain.Light, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	l, ok := m.lights[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return l, nil
}

func (m *mockLightRepository) FindByOwnerID(ownerID string) ([]*domain.Light, error) {
	return m.ownerMap[ownerID], nil
}

func (m *mockLightRepository) FindAll() ([]*domain.Light, error) {
	var all []*domain.Light
	for _, l := range m.lights {
		all = append(all, l)
	}
	return all, nil
}

func (m *mockLightRepository) Save(light *domain.Light) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = light
	return nil
}

// ---------------------------------------------------------------------------

// mockDeviceClient simula o cliente HTTP/MQTT para o ESP32.
// Os flags `turnOnCalled` e `turnOffCalled` permitem verificar se os comandos
// físicos foram de fato disparados.
type mockDeviceClient struct {
	turnOnCalled  bool
	turnOffCalled bool
	turnOnErr     error
	turnOffErr    error
	statusState   domain.LightState
	statusErr     error
}

func (m *mockDeviceClient) TurnOn(deviceID string) error {
	m.turnOnCalled = true
	return m.turnOnErr
}

func (m *mockDeviceClient) TurnOff(deviceID string) error {
	m.turnOffCalled = true
	return m.turnOffErr
}

func (m *mockDeviceClient) GetStatus(deviceID string) (domain.LightState, error) {
	return m.statusState, m.statusErr
}

// ---------------------------------------------------------------------------

// mockStatePublisher simula o hub WebSocket.
// `broadcastCalled` e os campos de argumento verificam se o estado foi
// transmitido para os clientes conectados após cada mudança.
type mockStatePublisher struct {
	broadcastCalled bool
	lastLightID     string
	lastName        string
	lastLocation    string
	lastState       string
}

func (m *mockStatePublisher) BroadcastState(lightID, name, location, state string) {
	m.broadcastCalled = true
	m.lastLightID = lightID
	m.lastName = name
	m.lastLocation = location
	m.lastState = state
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// novaLuzDoUsuario cria uma luz de teste pertencente ao usuário "user-1".
func novaLuzDoUsuario() *domain.Light {
	return domain.NewLight("luz-1", "Sala", "Sala de Estar", "user-1", "device-1")
}

// ---------------------------------------------------------------------------
// Testes — TurnOn
// ---------------------------------------------------------------------------

// TestTurnOn_LuzNaoEncontrada garante que o serviço retorna ErrLightNotFound
// quando o repositório não consegue localizar a luz pelo ID informado.
// Isso protege contra chamadas com IDs inválidos ou removidos.
func TestTurnOn_LuzNaoEncontrada(t *testing.T) {
	repo := newMockRepo()
	repo.findErr = errors.New("not found")

	svc := application.NewLightService(repo, &mockDeviceClient{}, &mockStatePublisher{})

	_, err := svc.TurnOn("luz-inexistente", "user-1")
	if !errors.Is(err, domain.ErrLightNotFound) {
		t.Errorf("esperado ErrLightNotFound, obtido: %v", err)
	}
}

// TestTurnOn_AcessoNegado garante que um usuário não pode acionar uma luz
// que pertence a outro usuário. Essencial para isolar o controle por tenant.
func TestTurnOn_AcessoNegado(t *testing.T) {
	repo := newMockRepo()
	luz := novaLuzDoUsuario() // owner = "user-1"
	repo.lights["luz-1"] = luz

	svc := application.NewLightService(repo, &mockDeviceClient{}, &mockStatePublisher{})

	_, err := svc.TurnOn("luz-1", "user-outro")
	if !errors.Is(err, domain.ErrLightAccessDenied) {
		t.Errorf("esperado ErrLightAccessDenied, obtido: %v", err)
	}
}

// TestTurnOn_DispositivoInacessivel garante que, quando o ESP32 não responde,
// o serviço retorna ErrDeviceUnreachable sem persistir nem transmitir nada.
// Isso evita inconsistências entre estado salvo e estado físico real.
func TestTurnOn_DispositivoInacessivel(t *testing.T) {
	repo := newMockRepo()
	repo.lights["luz-1"] = novaLuzDoUsuario()

	device := &mockDeviceClient{turnOnErr: errors.New("timeout")}
	pub := &mockStatePublisher{}

	svc := application.NewLightService(repo, device, pub)

	_, err := svc.TurnOn("luz-1", "user-1")
	if !errors.Is(err, domain.ErrDeviceUnreachable) {
		t.Errorf("esperado ErrDeviceUnreachable, obtido: %v", err)
	}

	// Nenhum estado deve ter sido persistido nem transmitido
	if repo.saved != nil {
		t.Error("Save não deveria ter sido chamado quando o dispositivo é inacessível")
	}
	if pub.broadcastCalled {
		t.Error("BroadcastState não deveria ter sido chamado quando o dispositivo é inacessível")
	}
}

// TestTurnOn_Sucesso valida o caminho feliz completo:
// o dispositivo físico é acionado, o estado é persistido como "on"
// e o publisher notifica todos os clientes conectados via WebSocket.
func TestTurnOn_Sucesso(t *testing.T) {
	repo := newMockRepo()
	repo.lights["luz-1"] = novaLuzDoUsuario()

	device := &mockDeviceClient{}
	pub := &mockStatePublisher{}

	svc := application.NewLightService(repo, device, pub)

	resp, err := svc.TurnOn("luz-1", "user-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	// Verifica retorno
	if resp.State != domain.LightStateOn {
		t.Errorf("esperado estado 'on', obtido: %s", resp.State)
	}
	if resp.ID != "luz-1" {
		t.Errorf("esperado ID 'luz-1', obtido: %s", resp.ID)
	}

	// Verifica que o dispositivo físico foi acionado
	if !device.turnOnCalled {
		t.Error("TurnOn do dispositivo deveria ter sido chamado")
	}

	// Verifica que o estado foi persistido corretamente
	if repo.saved == nil {
		t.Fatal("Save deveria ter sido chamado")
	}
	if repo.saved.State != domain.LightStateOn {
		t.Errorf("estado salvo esperado 'on', obtido: %s", repo.saved.State)
	}

	// Verifica que o broadcast foi disparado com os dados corretos
	if !pub.broadcastCalled {
		t.Error("BroadcastState deveria ter sido chamado")
	}
	if pub.lastLightID != "luz-1" {
		t.Errorf("broadcast com ID errado: %s", pub.lastLightID)
	}
	if pub.lastState != "on" {
		t.Errorf("broadcast com estado errado: %s", pub.lastState)
	}
}

// ---------------------------------------------------------------------------
// Testes — TurnOff
// ---------------------------------------------------------------------------

// TestTurnOff_Sucesso valida que o caminho feliz do TurnOff persiste
// o estado "off" e notifica o publisher, espelhando o comportamento do TurnOn.
func TestTurnOff_Sucesso(t *testing.T) {
	repo := newMockRepo()
	luz := novaLuzDoUsuario()
	luz.TurnOn() // começa ligada para o teste ter sentido
	repo.lights["luz-1"] = luz

	device := &mockDeviceClient{}
	pub := &mockStatePublisher{}

	svc := application.NewLightService(repo, device, pub)

	resp, err := svc.TurnOff("luz-1", "user-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.State != domain.LightStateOff {
		t.Errorf("esperado estado 'off', obtido: %s", resp.State)
	}

	if !device.turnOffCalled {
		t.Error("TurnOff do dispositivo deveria ter sido chamado")
	}

	if repo.saved == nil {
		t.Fatal("Save deveria ter sido chamado")
	}
	if repo.saved.State != domain.LightStateOff {
		t.Errorf("estado salvo esperado 'off', obtido: %s", repo.saved.State)
	}

	if !pub.broadcastCalled {
		t.Error("BroadcastState deveria ter sido chamado")
	}
	if pub.lastState != "off" {
		t.Errorf("broadcast com estado errado: %s", pub.lastState)
	}
}

// ---------------------------------------------------------------------------
// Testes — GetStatus
// ---------------------------------------------------------------------------

// TestGetStatus_Sucesso verifica que o serviço retorna o estado atual da luz
// sem acionar nenhum comando no dispositivo físico.
// Importante para operações de leitura que não devem gerar side effects.
func TestGetStatus_Sucesso(t *testing.T) {
	repo := newMockRepo()
	luz := novaLuzDoUsuario()
	luz.TurnOn()
	repo.lights["luz-1"] = luz

	device := &mockDeviceClient{}
	pub := &mockStatePublisher{}

	svc := application.NewLightService(repo, device, pub)

	resp, err := svc.GetStatus("luz-1", "user-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.State != domain.LightStateOn {
		t.Errorf("esperado estado 'on', obtido: %s", resp.State)
	}

	// GetStatus é somente leitura — nenhum dispositivo deve ser acionado
	if device.turnOnCalled || device.turnOffCalled {
		t.Error("GetStatus não deve acionar o dispositivo físico")
	}

	// Sem persistência nem broadcast para operações de leitura
	if repo.saved != nil {
		t.Error("GetStatus não deve chamar Save")
	}
	if pub.broadcastCalled {
		t.Error("GetStatus não deve chamar BroadcastState")
	}
}

// ---------------------------------------------------------------------------
// Testes — ListLights
// ---------------------------------------------------------------------------

// TestListLights_RetornaLuzesDoUsuario garante que o serviço lista apenas as
// luzes pertencentes ao usuário solicitado, sem vazar dados de outros tenants.
func TestListLights_RetornaLuzesDoUsuario(t *testing.T) {
	repo := newMockRepo()

	luz1 := domain.NewLight("luz-1", "Sala", "Sala de Estar", "user-1", "device-1")
	luz2 := domain.NewLight("luz-2", "Quarto", "Quarto Principal", "user-1", "device-2")
	luzOutro := domain.NewLight("luz-3", "Cozinha", "Cozinha", "user-outro", "device-3")

	repo.ownerMap["user-1"] = []*domain.Light{luz1, luz2}
	repo.ownerMap["user-outro"] = []*domain.Light{luzOutro}

	svc := application.NewLightService(repo, &mockDeviceClient{}, &mockStatePublisher{})

	resps, err := svc.ListLights("user-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(resps) != 2 {
		t.Fatalf("esperado 2 luzes para user-1, obtido: %d", len(resps))
	}

	ids := map[string]bool{}
	for _, r := range resps {
		ids[r.ID] = true
	}

	if !ids["luz-1"] || !ids["luz-2"] {
		t.Error("resposta deveria conter luz-1 e luz-2")
	}
	if ids["luz-3"] {
		t.Error("luz-3 não pertence a user-1 e não deveria aparecer na listagem")
	}
}

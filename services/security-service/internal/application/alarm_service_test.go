package application_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"aurora/services/security-service/internal/application"
	"aurora/services/security-service/internal/domain"
)

// ---------------------------------------------------------------------------
// Mocks manuais
// ---------------------------------------------------------------------------

// mockAlarmRepo implementa domain.AlarmRepository em memória.
// Permite configurar o evento a ser retornado em FindByID e o erro a ser
// propagado, facilitando o controle total nas asserções dos testes.
type mockAlarmRepo struct {
	saved        []*domain.AlarmEvent
	findEvent    *domain.AlarmEvent
	findErr      error
	recentEvents []*domain.AlarmEvent
}

func (m *mockAlarmRepo) Save(event *domain.AlarmEvent) error {
	m.saved = append(m.saved, event)
	return nil
}

func (m *mockAlarmRepo) FindByID(id string) (*domain.AlarmEvent, error) {
	return m.findEvent, m.findErr
}

func (m *mockAlarmRepo) FindRecent(limit int) ([]*domain.AlarmEvent, error) {
	return m.recentEvents, nil
}

// mockBuzzerClient implementa domain.BuzzerClient.
// Usa atomic para que os flags possam ser lidos com segurança fora da goroutine
// que os modifica (necessário no teste de acionamento em background).
type mockBuzzerClient struct {
	turnOnCalled  atomic.Bool
	turnOffCalled atomic.Bool
}

func (m *mockBuzzerClient) TurnOn(deviceID string) error {
	m.turnOnCalled.Store(true)
	return nil
}

func (m *mockBuzzerClient) TurnOff(deviceID string) error {
	m.turnOffCalled.Store(true)
	return nil
}

// ---------------------------------------------------------------------------
// Testes
// ---------------------------------------------------------------------------

// TestTriggerAlarm_TipoInvalido verifica que a camada de aplicação rejeita
// gatilhos desconhecidos antes mesmo de tentar persistir qualquer coisa.
// Isso é importante para garantir que a validação do domínio é respeitada
// e que eventos inválidos nunca chegam ao repositório.
func TestTriggerAlarm_TipoInvalido(t *testing.T) {
	repo := &mockAlarmRepo{}
	buzzer := &mockBuzzerClient{}
	svc := application.NewAlarmService(repo, buzzer, 100, "device-1", nil)

	resp, err := svc.TriggerAlarm("user-1", "unknown", "sensor-1", "sala")

	if err == nil {
		t.Fatal("esperava erro para tipo de gatilho inválido, mas não recebeu nenhum")
	}
	if resp != nil {
		t.Errorf("esperava resposta nil em caso de erro, mas recebeu: %+v", resp)
	}
	if len(repo.saved) != 0 {
		t.Errorf("nenhum evento deveria ter sido salvo, mas %d foram", len(repo.saved))
	}
}

// TestTriggerAlarm_Sucesso verifica que um disparo válido persiste o evento
// e retorna uma resposta com os dados corretos e status "triggered".
// Garantir esse fluxo feliz é essencial para o comportamento central do serviço.
func TestTriggerAlarm_Sucesso(t *testing.T) {
	repo := &mockAlarmRepo{}
	buzzer := &mockBuzzerClient{}
	// Duração longa para que o buzzer não interfira neste teste
	svc := application.NewAlarmService(repo, buzzer, 10000, "device-1", nil)

	resp, err := svc.TriggerAlarm("user-1", "motion", "sensor-42", "corredor")

	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
	if resp == nil {
		t.Fatal("esperava resposta não-nil")
	}
	if resp.Status != domain.AlarmStatusTriggered {
		t.Errorf("status esperado %q, recebido %q", domain.AlarmStatusTriggered, resp.Status)
	}
	if resp.TriggerType != "motion" {
		t.Errorf("triggerType esperado %q, recebido %q", "motion", resp.TriggerType)
	}
	if resp.SensorID != "sensor-42" {
		t.Errorf("sensorID esperado %q, recebido %q", "sensor-42", resp.SensorID)
	}
	if resp.Location != "corredor" {
		t.Errorf("location esperada %q, recebida %q", "corredor", resp.Location)
	}
	if resp.ID == "" {
		t.Error("ID do evento não deve ser vazio")
	}
	if len(repo.saved) != 1 {
		t.Errorf("esperava 1 evento salvo, mas foram %d", len(repo.saved))
	}
}

// TestTriggerAlarm_BuzzerEAcionadoEmBackground verifica que o buzzer é ligado
// e depois desligado de forma assíncrona após o disparo do alarme.
// Esse teste é importante porque a goroutine não deve ser silenciosamente
// ignorada: o hardware precisa ser acionado e desligado após o tempo configurado.
func TestTriggerAlarm_BuzzerEAcionadoEmBackground(t *testing.T) {
	repo := &mockAlarmRepo{}
	buzzer := &mockBuzzerClient{}
	// Duração curta para que a goroutine complete dentro do tempo de espera do teste
	svc := application.NewAlarmService(repo, buzzer, 10, "device-1", nil)

	_, err := svc.TriggerAlarm("user-1", "manual", "sensor-1", "entrada")
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}

	// Aguarda tempo suficiente para a goroutine completar TurnOn + sleep(10ms) + TurnOff
	time.Sleep(50 * time.Millisecond)

	if !buzzer.turnOnCalled.Load() {
		t.Error("esperava que TurnOn tivesse sido chamado, mas não foi")
	}
	if !buzzer.turnOffCalled.Load() {
		t.Error("esperava que TurnOff tivesse sido chamado, mas não foi")
	}
}

// TestSilenceAlarm_NaoEncontrado verifica que tentar silenciar um alarme
// inexistente retorna ErrAlarmNotFound.
// É fundamental que a API exponha o erro correto para que o cliente possa
// distinguir "não encontrado" de outros erros internos.
func TestSilenceAlarm_NaoEncontrado(t *testing.T) {
	repo := &mockAlarmRepo{
		findErr: errors.New("qualquer erro de busca"),
	}
	buzzer := &mockBuzzerClient{}
	svc := application.NewAlarmService(repo, buzzer, 100, "device-1", nil)

	resp, err := svc.SilenceAlarm("id-inexistente")

	if !errors.Is(err, domain.ErrAlarmNotFound) {
		t.Errorf("esperava ErrAlarmNotFound, recebeu: %v", err)
	}
	if resp != nil {
		t.Errorf("esperava resposta nil, recebeu: %+v", resp)
	}
}

// TestSilenceAlarm_Sucesso verifica que silenciar um alarme existente persiste
// o estado atualizado, desliga o buzzer e retorna status "silenced".
// Esse fluxo é crítico para o controle operacional do sistema de segurança.
func TestSilenceAlarm_Sucesso(t *testing.T) {
	existing, err := domain.NewAlarmEvent("user-1", "motion", "sensor-10", "garagem")
	if err != nil {
		t.Fatalf("falha ao criar evento de domínio para o teste: %v", err)
	}

	repo := &mockAlarmRepo{
		findEvent: existing,
	}
	buzzer := &mockBuzzerClient{}
	svc := application.NewAlarmService(repo, buzzer, 100, "device-1", nil)

	resp, err := svc.SilenceAlarm(existing.ID)

	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
	if resp == nil {
		t.Fatal("esperava resposta não-nil")
	}
	if resp.Status != domain.AlarmStatusSilenced {
		t.Errorf("status esperado %q, recebido %q", domain.AlarmStatusSilenced, resp.Status)
	}
	// Verifica que o evento foi re-salvo após a silenciação
	if len(repo.saved) != 1 {
		t.Errorf("esperava 1 save após silenciar, recebeu %d", len(repo.saved))
	}
	if !buzzer.turnOffCalled.Load() {
		t.Error("esperava que TurnOff tivesse sido chamado ao silenciar")
	}
}

// TestGetRecentAlarms_RetornaLista verifica que o serviço converte corretamente
// múltiplos eventos do repositório em respostas da camada de aplicação.
// Garante que a transformação de domínio → DTO não descarta nem duplica itens.
func TestGetRecentAlarms_RetornaLista(t *testing.T) {
	ev1, _ := domain.NewAlarmEvent("user-1", "motion", "sensor-1", "sala")
	ev2, _ := domain.NewAlarmEvent("user-1", "manual", "sensor-2", "cozinha")

	repo := &mockAlarmRepo{
		recentEvents: []*domain.AlarmEvent{ev1, ev2},
	}
	buzzer := &mockBuzzerClient{}
	svc := application.NewAlarmService(repo, buzzer, 100, "device-1", nil)

	responses, err := svc.GetRecentAlarms(10)

	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
	if len(responses) != 2 {
		t.Errorf("esperava 2 respostas, recebeu %d", len(responses))
	}
}

package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"aurora/services/schedule-service/internal/application"
	"aurora/services/schedule-service/internal/domain"
)

// =============================================================================
// MOCKS
// =============================================================================

// MockScheduleRepo implementa domain.ScheduleRepository em memória.
type MockScheduleRepo struct {
	schedules  map[string]*domain.Schedule
	saved      []*domain.Schedule
	deleted    []string
	findErr    error
	saveErr    error
	deleteErr  error
	allEnabled []*domain.Schedule
}

func newMockScheduleRepo() *MockScheduleRepo {
	return &MockScheduleRepo{schedules: make(map[string]*domain.Schedule)}
}

func (m *MockScheduleRepo) Save(_ context.Context, s *domain.Schedule) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.schedules[s.ID] = s
	m.saved = append(m.saved, s)
	return nil
}

func (m *MockScheduleRepo) FindByID(_ context.Context, id string) (*domain.Schedule, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	s, ok := m.schedules[id]
	if !ok {
		return nil, domain.ErrScheduleNotFound
	}
	return s, nil
}

func (m *MockScheduleRepo) FindByOwnerID(_ context.Context, ownerID string) ([]*domain.Schedule, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []*domain.Schedule
	for _, s := range m.schedules {
		if s.OwnerID == ownerID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockScheduleRepo) FindAllEnabled(_ context.Context) ([]*domain.Schedule, error) {
	return m.allEnabled, nil
}

func (m *MockScheduleRepo) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.schedules, id)
	m.deleted = append(m.deleted, id)
	return nil
}

// MockExecutionRepo implementa domain.ScheduleExecutionRepository em memória.
type MockExecutionRepo struct {
	saved      []*domain.ScheduleExecution
	executions map[string][]*domain.ScheduleExecution
	saveErr    error
}

func newMockExecutionRepo() *MockExecutionRepo {
	return &MockExecutionRepo{executions: make(map[string][]*domain.ScheduleExecution)}
}

func (m *MockExecutionRepo) Save(_ context.Context, e *domain.ScheduleExecution) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, e)
	m.executions[e.ScheduleID] = append(m.executions[e.ScheduleID], e)
	return nil
}

func (m *MockExecutionRepo) FindByScheduleID(_ context.Context, scheduleID string, limit int) ([]*domain.ScheduleExecution, error) {
	execs := m.executions[scheduleID]
	if limit > 0 && len(execs) > limit {
		execs = execs[:limit]
	}
	return execs, nil
}

// MockDispatcher implementa domain.ActionDispatcher.
type MockDispatcher struct {
	dispatched  []*domain.Schedule
	dispatchErr error
}

func (m *MockDispatcher) Dispatch(_ context.Context, s *domain.Schedule) error {
	m.dispatched = append(m.dispatched, s)
	return m.dispatchErr
}

// MockRegistry implementa domain.SchedulerRegistry.
type MockRegistry struct {
	registered   []*domain.Schedule
	unregistered []string
	registerErr  error
}

func (m *MockRegistry) Register(s *domain.Schedule) error {
	m.registered = append(m.registered, s)
	return m.registerErr
}

func (m *MockRegistry) Unregister(scheduleID string) {
	m.unregistered = append(m.unregistered, scheduleID)
}

// =============================================================================
// HELPERS
// =============================================================================

func newService(
	schedRepo *MockScheduleRepo,
	execRepo *MockExecutionRepo,
	dispatcher *MockDispatcher,
	registry *MockRegistry,
) *application.ScheduleService {
	return application.NewScheduleService(schedRepo, execRepo, dispatcher, registry, 5*time.Second)
}

// buildCronSchedule cria e armazena um agendamento cron válido no repo.
func buildCronSchedule(t *testing.T, repo *MockScheduleRepo, id, ownerID string, enabled bool) *domain.Schedule {
	t.Helper()
	s := &domain.Schedule{
		ID:           id,
		OwnerID:      ownerID,
		Name:         "Test Schedule",
		ScheduleType: domain.ScheduleTypeCron,
		CronExpr:     "0 * * * *",
		DaysFilter:   0,
		Timezone:     "UTC",
		Action: domain.Action{
			Target:   domain.ActionTargetHTTP,
			Type:     domain.ActionTurnOnLight,
			DeviceID: "device-1",
		},
		Enabled:   enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.schedules[id] = s
	return s
}

// buildOneShotSchedule cria e armazena um agendamento one_shot válido no repo.
func buildOneShotSchedule(t *testing.T, repo *MockScheduleRepo, id, ownerID string, enabled bool) *domain.Schedule {
	t.Helper()
	runAt := time.Now().Add(time.Hour)
	s := &domain.Schedule{
		ID:           id,
		OwnerID:      ownerID,
		Name:         "One Shot Schedule",
		ScheduleType: domain.ScheduleTypeOneShot,
		RunAt:        &runAt,
		DaysFilter:   0,
		Timezone:     "UTC",
		Action: domain.Action{
			Target:   domain.ActionTargetHTTP,
			Type:     domain.ActionTurnOnLight,
			DeviceID: "device-1",
		},
		Enabled:   enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.schedules[id] = s
	return s
}

// =============================================================================
// TESTES — CreateSchedule
// =============================================================================

// TestCreateSchedule_Sucesso verifica que um agendamento válido é salvo, registrado e retornado com ID.
func TestCreateSchedule_Sucesso(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	execRepo := newMockExecutionRepo()
	dispatcher := &MockDispatcher{}
	registry := &MockRegistry{}
	svc := newService(schedRepo, execRepo, dispatcher, registry)

	resp, err := svc.CreateSchedule(application.CreateScheduleRequest{
		Name:           "Ligar luz",
		OwnerID:        "owner-1",
		ScheduleType:   domain.ScheduleTypeCron,
		CronExpr:       "0 8 * * *",
		DaysFilter:     0,
		Timezone:       "UTC",
		ActionTarget:   "http",
		ActionType:     "turn_on_light",
		ActionDeviceID: "light-sala",
	})

	if err != nil {
		t.Fatalf("CreateSchedule() erro inesperado: %v", err)
	}
	if resp.ID == "" {
		t.Error("ID da resposta não deve ser vazio")
	}
	if len(schedRepo.saved) != 1 {
		t.Errorf("esperado 1 Save no repo, obteve %d", len(schedRepo.saved))
	}
	if len(registry.registered) != 1 {
		t.Errorf("esperado 1 Register no registry, obteve %d", len(registry.registered))
	}
}

// TestCreateSchedule_ActionInvalida verifica que um action_target inválido retorna ErrInvalidAction.
func TestCreateSchedule_ActionInvalida(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	execRepo := newMockExecutionRepo()
	dispatcher := &MockDispatcher{}
	registry := &MockRegistry{}
	svc := newService(schedRepo, execRepo, dispatcher, registry)

	_, err := svc.CreateSchedule(application.CreateScheduleRequest{
		Name:           "Ligar luz",
		OwnerID:        "owner-1",
		ScheduleType:   domain.ScheduleTypeCron,
		CronExpr:       "0 8 * * *",
		ActionTarget:   "invalido",
		ActionType:     "turn_on_light",
		ActionDeviceID: "light-sala",
	})

	if !errors.Is(err, domain.ErrInvalidAction) {
		t.Errorf("esperado ErrInvalidAction, obteve: %v", err)
	}
}

// =============================================================================
// TESTES — GetSchedule
// =============================================================================

// TestGetSchedule_AcessoNegado verifica que ownerID diferente retorna ErrScheduleAccessDenied.
func TestGetSchedule_AcessoNegado(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	buildCronSchedule(t, schedRepo, "sched-1", "owner-1", true)

	svc := newService(schedRepo, newMockExecutionRepo(), &MockDispatcher{}, &MockRegistry{})

	_, err := svc.GetSchedule("sched-1", "owner-outro")

	if !errors.Is(err, domain.ErrScheduleAccessDenied) {
		t.Errorf("esperado ErrScheduleAccessDenied, obteve: %v", err)
	}
}

// =============================================================================
// TESTES — ListSchedules
// =============================================================================

// TestListSchedules_RetornaApenasDoDono verifica que apenas os agendamentos do owner são retornados.
func TestListSchedules_RetornaApenasDoDono(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	buildCronSchedule(t, schedRepo, "sched-1", "owner-1", true)
	buildCronSchedule(t, schedRepo, "sched-2", "owner-1", false)
	buildCronSchedule(t, schedRepo, "sched-3", "owner-outro", true)

	svc := newService(schedRepo, newMockExecutionRepo(), &MockDispatcher{}, &MockRegistry{})

	responses, err := svc.ListSchedules("owner-1")

	if err != nil {
		t.Fatalf("ListSchedules() erro inesperado: %v", err)
	}
	if len(responses) != 2 {
		t.Errorf("esperado 2 agendamentos para owner-1, obteve %d", len(responses))
	}
	for _, r := range responses {
		if r.OwnerID != "owner-1" {
			t.Errorf("agendamento retornado com OwnerID errado: %s", r.OwnerID)
		}
	}
}

// =============================================================================
// TESTES — DeleteSchedule
// =============================================================================

// TestDeleteSchedule_Sucesso verifica que Unregister e Delete são chamados corretamente.
func TestDeleteSchedule_Sucesso(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	buildCronSchedule(t, schedRepo, "sched-1", "owner-1", true)

	registry := &MockRegistry{}
	svc := newService(schedRepo, newMockExecutionRepo(), &MockDispatcher{}, registry)

	err := svc.DeleteSchedule("sched-1", "owner-1")

	if err != nil {
		t.Fatalf("DeleteSchedule() erro inesperado: %v", err)
	}
	if len(registry.unregistered) != 1 || registry.unregistered[0] != "sched-1" {
		t.Errorf("esperado Unregister('sched-1'), unregistered=%v", registry.unregistered)
	}
	if len(schedRepo.deleted) != 1 || schedRepo.deleted[0] != "sched-1" {
		t.Errorf("esperado Delete('sched-1') no repo, deleted=%v", schedRepo.deleted)
	}
}

// =============================================================================
// TESTES — ToggleSchedule
// =============================================================================

// TestToggleSchedule_Desabilita verifica que ao desabilitar: Unregister é chamado e Enabled=false é salvo.
func TestToggleSchedule_Desabilita(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	buildCronSchedule(t, schedRepo, "sched-1", "owner-1", true) // começa habilitado

	registry := &MockRegistry{}
	svc := newService(schedRepo, newMockExecutionRepo(), &MockDispatcher{}, registry)

	resp, err := svc.ToggleSchedule("sched-1", "owner-1")

	if err != nil {
		t.Fatalf("ToggleSchedule() erro inesperado: %v", err)
	}
	if resp.Enabled {
		t.Error("esperado Enabled=false após toggle de schedule habilitado")
	}
	if len(registry.unregistered) != 1 {
		t.Errorf("esperado 1 Unregister, obteve %d", len(registry.unregistered))
	}
	if len(registry.registered) != 0 {
		t.Errorf("esperado 0 Register, obteve %d", len(registry.registered))
	}
	// Verifica que o estado foi persistido
	saved := schedRepo.saved[len(schedRepo.saved)-1]
	if saved.Enabled {
		t.Error("esperado que o schedule salvo tenha Enabled=false")
	}
}

// TestToggleSchedule_Habilita verifica que ao habilitar: Register é chamado e Enabled=true é salvo.
func TestToggleSchedule_Habilita(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	buildCronSchedule(t, schedRepo, "sched-1", "owner-1", false) // começa desabilitado

	registry := &MockRegistry{}
	svc := newService(schedRepo, newMockExecutionRepo(), &MockDispatcher{}, registry)

	resp, err := svc.ToggleSchedule("sched-1", "owner-1")

	if err != nil {
		t.Fatalf("ToggleSchedule() erro inesperado: %v", err)
	}
	if !resp.Enabled {
		t.Error("esperado Enabled=true após toggle de schedule desabilitado")
	}
	if len(registry.registered) != 1 {
		t.Errorf("esperado 1 Register, obteve %d", len(registry.registered))
	}
	if len(registry.unregistered) != 0 {
		t.Errorf("esperado 0 Unregister, obteve %d", len(registry.unregistered))
	}
	// Verifica que o estado foi persistido
	saved := schedRepo.saved[len(schedRepo.saved)-1]
	if !saved.Enabled {
		t.Error("esperado que o schedule salvo tenha Enabled=true")
	}
}

// =============================================================================
// TESTES — LoadAllEnabled
// =============================================================================

// TestLoadAllEnabled_Delega verifica que LoadAllEnabled delega ao repo.FindAllEnabled.
func TestLoadAllEnabled_Delega(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	s1 := buildCronSchedule(t, schedRepo, "sched-1", "owner-1", true)
	s2 := buildCronSchedule(t, schedRepo, "sched-2", "owner-2", true)
	schedRepo.allEnabled = []*domain.Schedule{s1, s2}

	svc := newService(schedRepo, newMockExecutionRepo(), &MockDispatcher{}, &MockRegistry{})

	result, err := svc.LoadAllEnabled(context.Background())

	if err != nil {
		t.Fatalf("LoadAllEnabled() erro inesperado: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("esperado 2 agendamentos habilitados, obteve %d", len(result))
	}
}

// =============================================================================
// TESTES — ExecuteSchedule
// =============================================================================

// TestExecuteSchedule_OneShotDesabilita verifica que um agendamento one_shot é desabilitado após execução
// e que o dispatcher é chamado e a execução é salva com status success.
func TestExecuteSchedule_OneShotDesabilita(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	buildOneShotSchedule(t, schedRepo, "sched-1", "owner-1", true)

	execRepo := newMockExecutionRepo()
	dispatcher := &MockDispatcher{}
	registry := &MockRegistry{}
	svc := newService(schedRepo, execRepo, dispatcher, registry)

	err := svc.ExecuteSchedule("sched-1")

	if err != nil {
		t.Fatalf("ExecuteSchedule() erro inesperado: %v", err)
	}
	// Dispatcher deve ter sido chamado
	if len(dispatcher.dispatched) != 1 {
		t.Errorf("esperado 1 Dispatch, obteve %d", len(dispatcher.dispatched))
	}
	// Execução deve ter sido salva com status success
	if len(execRepo.saved) != 1 {
		t.Fatalf("esperado 1 execução salva, obteve %d", len(execRepo.saved))
	}
	if execRepo.saved[0].Status != domain.ExecutionStatusSuccess {
		t.Errorf("esperado status success, obteve %s", execRepo.saved[0].Status)
	}
	// Schedule deve ter sido desabilitado (Unregister chamado)
	if len(registry.unregistered) != 1 {
		t.Errorf("esperado 1 Unregister para one_shot, obteve %d", len(registry.unregistered))
	}
	// Verifica que o schedule foi salvo com Enabled=false
	sched := schedRepo.schedules["sched-1"]
	if sched.Enabled {
		t.Error("esperado Enabled=false após execução de one_shot")
	}
}

// TestExecuteSchedule_DispatchChamado verifica que para um agendamento cron com DaysFilter=0
// (todos os dias ativos), o dispatcher é chamado e a execução é salva com status success.
func TestExecuteSchedule_DispatchChamado(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	buildCronSchedule(t, schedRepo, "sched-cron", "owner-1", true) // DaysFilter=0

	execRepo := newMockExecutionRepo()
	dispatcher := &MockDispatcher{}
	svc := newService(schedRepo, execRepo, dispatcher, &MockRegistry{})

	err := svc.ExecuteSchedule("sched-cron")

	if err != nil {
		t.Fatalf("ExecuteSchedule() erro inesperado: %v", err)
	}
	if len(dispatcher.dispatched) != 1 {
		t.Errorf("esperado 1 Dispatch (DaysFilter=0 = todos os dias), obteve %d", len(dispatcher.dispatched))
	}
	if len(execRepo.saved) != 1 {
		t.Fatalf("esperado 1 execução salva, obteve %d", len(execRepo.saved))
	}
	if execRepo.saved[0].Status != domain.ExecutionStatusSuccess {
		t.Errorf("esperado status success, obteve %s", execRepo.saved[0].Status)
	}
}

// TestExecuteSchedule_DiaInativo verifica que um agendamento com DaysFilter excluindo o dia atual
// tem execução salva com status skipped e o dispatcher NÃO é chamado.
func TestExecuteSchedule_DiaInativo(t *testing.T) {
	schedRepo := newMockScheduleRepo()

	// Determina o dia atual e cria um bitmask que NÃO inclui nenhum dia
	// (bitmask com apenas o dia de hoje ausente — usando um valor que exclui todos os dias possíveis)
	// Estratégia: DaysFilter com um bitmask que não inclui o dia de hoje.
	// Os dias possíveis são: Sun=1, Mon=2, Tue=4, Wed=8, Thu=16, Fri=32, Sat=64 (total=127).
	// Excluímos o dia atual calculando ~(1 << today) & 127.
	today := time.Now().UTC().Weekday()
	allDays := 127 // bits 0-6 todos setados
	daysFilterSemHoje := allDays &^ (1 << uint(today))

	s := &domain.Schedule{
		ID:           "sched-inativo",
		OwnerID:      "owner-1",
		Name:         "Schedule Inativo Hoje",
		ScheduleType: domain.ScheduleTypeCron,
		CronExpr:     "0 * * * *",
		DaysFilter:   daysFilterSemHoje,
		Timezone:     "UTC",
		Action: domain.Action{
			Target:   domain.ActionTargetHTTP,
			Type:     domain.ActionTurnOnLight,
			DeviceID: "device-1",
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	schedRepo.schedules["sched-inativo"] = s

	execRepo := newMockExecutionRepo()
	dispatcher := &MockDispatcher{}
	svc := newService(schedRepo, execRepo, dispatcher, &MockRegistry{})

	err := svc.ExecuteSchedule("sched-inativo")

	if err != nil {
		t.Fatalf("ExecuteSchedule() erro inesperado: %v", err)
	}
	// Dispatcher NÃO deve ter sido chamado
	if len(dispatcher.dispatched) != 0 {
		t.Errorf("esperado 0 Dispatch (dia inativo), obteve %d", len(dispatcher.dispatched))
	}
	// Execução deve ter sido salva com status skipped
	if len(execRepo.saved) != 1 {
		t.Fatalf("esperado 1 execução salva, obteve %d", len(execRepo.saved))
	}
	if execRepo.saved[0].Status != domain.ExecutionStatusSkipped {
		t.Errorf("esperado status skipped, obteve %s", execRepo.saved[0].Status)
	}
}

// =============================================================================
// TESTES — ListExecutions
// =============================================================================

// TestListExecutions_AcessoNegado verifica que ownerID diferente retorna ErrScheduleAccessDenied.
func TestListExecutions_AcessoNegado(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	buildCronSchedule(t, schedRepo, "sched-1", "owner-1", true)

	svc := newService(schedRepo, newMockExecutionRepo(), &MockDispatcher{}, &MockRegistry{})

	_, err := svc.ListExecutions("sched-1", "owner-outro", 10)

	if !errors.Is(err, domain.ErrScheduleAccessDenied) {
		t.Errorf("esperado ErrScheduleAccessDenied, obteve: %v", err)
	}
}

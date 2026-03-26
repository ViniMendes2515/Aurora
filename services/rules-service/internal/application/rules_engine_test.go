package application_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"aurora/services/rules-service/internal/application"
	"aurora/services/rules-service/internal/domain"
)

// =============================================================================
// MOCK — RuleRepository em memória
// =============================================================================

// mockRuleRepo implementa domain.RuleRepository
type mockRuleRepo struct {
	rules     []*domain.Rule // estado inicial das regras
	saved     []*domain.Rule // regras que foram salvas via Save()
	findErr   error          // erro para FindByTrigger e FindByOwnerID
	deleteErr error          // erro para Delete
}

func (m *mockRuleRepo) Save(rule *domain.Rule) error {
	m.saved = append(m.saved, rule)
	return nil
}

func (m *mockRuleRepo) FindByID(id string) (*domain.Rule, error) {
	for _, r := range m.rules {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, domain.ErrRuleNotFound
}

func (m *mockRuleRepo) FindByOwnerID(ownerID string) ([]*domain.Rule, error) {
	var result []*domain.Rule
	for _, r := range m.rules {
		if r.OwnerID == ownerID || r.OwnerID == "*" {
			result = append(result, r)
		}
	}
	return result, m.findErr
}

// FindByTrigger retorna regras que batem com tipo e sensorID.
func (m *mockRuleRepo) FindByTrigger(triggerType domain.TriggerType, triggerID string) ([]*domain.Rule, error) {
	var result []*domain.Rule
	for _, r := range m.rules {
		if r.TriggerType == triggerType && r.TriggerID == triggerID {
			result = append(result, r)
		}
	}
	return result, m.findErr
}

func (m *mockRuleRepo) Delete(id string) error { return m.deleteErr }

// =============================================================================
// HELPER — servidor HTTP de teste
// =============================================================================
func novoServidorHTTPContador(t *testing.T) (*httptest.Server, *int64) {
	t.Helper()

	var contagem int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&contagem, 1)
		w.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(srv.Close)
	return srv, &contagem
}

// =============================================================================
// TESTES — CreateRule
// =============================================================================

// TestCreateRule_NomeVazio verifica que criar uma regra sem nome retorna ErrInvalidRule.
func TestCreateRule_NomeVazio(t *testing.T) {
	repo := &mockRuleRepo{}
	engine := application.NewRulesEngine(repo, "http://lighting", "http://security", "key", 0)

	_, err := engine.CreateRule(application.CreateRuleRequest{
		Name:        "",
		OwnerID:     "user1",
		TriggerType: domain.TriggerMotion,
		ActionType:  domain.ActionTurnOnLight,
	})

	if err != domain.ErrInvalidRule {
		t.Errorf("esperado ErrInvalidRule, obteve: %v", err)
	}
}

// TestCreateRule_Sucesso verifica que uma regra válida é salva e retornada corretamente.
func TestCreateRule_Sucesso(t *testing.T) {
	repo := &mockRuleRepo{}
	engine := application.NewRulesEngine(repo, "http://lighting", "http://security", "key", 0)

	resp, err := engine.CreateRule(application.CreateRuleRequest{
		Name:        "Liga luz ao detectar movimento",
		OwnerID:     "user1",
		TriggerType: domain.TriggerMotion,
		TriggerID:   "sensor-pir-01",
		ActionType:  domain.ActionTurnOnLight,
		ActionID:    "light-sala",
	})

	if err != nil {
		t.Fatalf("CreateRule() erro inesperado: %v", err)
	}
	if resp.ID == "" {
		t.Error("ID da regra criada não deve ser vazio")
	}
	if resp.Name != "Liga luz ao detectar movimento" {
		t.Errorf("Nome esperado 'Liga luz ao detectar movimento', obteve: %s", resp.Name)
	}
	if !resp.Enabled {
		t.Error("Nova regra deve ser criada habilitada")
	}
	// Verifica que Save foi chamado
	if len(repo.saved) != 1 {
		t.Errorf("esperado 1 regra salva, obteve: %d", len(repo.saved))
	}
}

// =============================================================================
// TESTES — EvaluateMotionTrigger
// =============================================================================
func TestEvaluateMotionTrigger_RegraDesabilitadaNaoExecuta(t *testing.T) {
	srv, contagem := novoServidorHTTPContador(t)

	regraDesabilitada := &domain.Rule{
		ID:          "regra-1",
		Name:        "Regra desabilitada",
		TriggerType: domain.TriggerMotion,
		TriggerID:   "sensor-01",
		ActionType:  domain.ActionTurnOnLight,
		ActionID:    "light-1",
		Enabled:     false,
	}

	repo := &mockRuleRepo{rules: []*domain.Rule{regraDesabilitada}}
	engine := application.NewRulesEngine(repo, srv.URL, srv.URL, "key", 0)

	engine.EvaluateMotionTrigger("sensor-01")

	// Dá um momento para qualquer goroutine que possa ter sido disparada
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt64(contagem); got != 0 {
		t.Errorf("esperado 0 chamadas HTTP (regra desabilitada), mas %d foram feitas", got)
	}
}

// TestEvaluateMotionTrigger_RegraHabilitadaExecutaAcao verifica que uma regra
func TestEvaluateMotionTrigger_RegraHabilitadaExecutaAcao(t *testing.T) {
	srv, contagem := novoServidorHTTPContador(t)

	regraHabilitada := &domain.Rule{
		ID:          "regra-2",
		Name:        "Liga luz no movimento",
		TriggerType: domain.TriggerMotion,
		TriggerID:   "sensor-01",
		ActionType:  domain.ActionTurnOnLight,
		ActionID:    "light-sala",
		Enabled:     true,
	}

	repo := &mockRuleRepo{rules: []*domain.Rule{regraHabilitada}}
	engine := application.NewRulesEngine(repo, srv.URL, srv.URL, "key", 0)

	engine.EvaluateMotionTrigger("sensor-01")

	// Aguarda a chamada HTTP ser processada
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt64(contagem); got < 1 {
		t.Errorf("esperado pelo menos 1 chamada HTTP (ação da regra), mas %d foram feitas", got)
	}
}

// TestEvaluateMotionTrigger_AutoOffEAgendado verifica que após uma regra de "ligar luz"
func TestEvaluateMotionTrigger_AutoOffEAgendado(t *testing.T) {
	srv, contagem := novoServidorHTTPContador(t)

	regraMotion := &domain.Rule{
		ID:          "regra-3",
		Name:        "Auto-off test",
		TriggerType: domain.TriggerMotion,
		TriggerID:   "sensor-01",
		ActionType:  domain.ActionTurnOnLight,
		ActionID:    "light-sala",
		Enabled:     true,
	}

	repo := &mockRuleRepo{rules: []*domain.Rule{regraMotion}}

	timeoutCurto := 20 * time.Millisecond
	engine := application.NewRulesEngine(repo, srv.URL, srv.URL, "key", timeoutCurto)

	engine.EvaluateMotionTrigger("sensor-01")

	time.Sleep(150 * time.Millisecond)

	if got := atomic.LoadInt64(contagem); got < 2 {
		t.Errorf("esperado pelo menos 2 chamadas HTTP (on + auto-off), obteve %d", got)
	}
}

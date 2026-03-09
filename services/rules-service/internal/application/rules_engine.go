package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"aurora/services/rules-service/internal/domain"
)

// RulesEngine implementa o motor de regras de automação
type RulesEngine struct {
	ruleRepo        domain.RuleRepository
	lightingBaseURL string
	securityBaseURL string
	deviceAPIKey    string

	// Auto-desligar por inatividade de movimento
	motionOffTimeout time.Duration
	mu               sync.RWMutex
	lastMotionByLight map[string]time.Time // lightID → última vez que uma regra de motion ligou essa luz
}

// NewRulesEngine cria uma nova instância do motor de regras
func NewRulesEngine(ruleRepo domain.RuleRepository, lightingBaseURL, securityBaseURL, deviceAPIKey string, motionOffTimeout time.Duration) *RulesEngine {
	if motionOffTimeout <= 0 {
		motionOffTimeout = 5 * time.Minute
	}
	return &RulesEngine{
		ruleRepo:         ruleRepo,
		lightingBaseURL:  lightingBaseURL,
		securityBaseURL:  securityBaseURL,
		deviceAPIKey:     deviceAPIKey,
		motionOffTimeout: motionOffTimeout,
		lastMotionByLight: make(map[string]time.Time),
	}
}

// CreateRuleRequest representa os dados de entrada para criação de uma regra
type CreateRuleRequest struct {
	Name             string            `json:"name"`
	OwnerID          string            `json:"-"`
	TriggerType      domain.TriggerType `json:"trigger_type"`
	TriggerID        string            `json:"trigger_id"`
	TriggerThreshold float64           `json:"trigger_threshold"` // Para light_low: considerar luz baixa abaixo deste % (0 = usar 30)
	ActionType       domain.ActionType `json:"action_type"`
	ActionID         string            `json:"action_id"`
}

// RuleResponse representa a resposta de uma regra
type RuleResponse struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	TriggerType      domain.TriggerType `json:"trigger_type"`
	TriggerID        string            `json:"trigger_id"`
	TriggerThreshold float64           `json:"trigger_threshold"`
	ActionType       domain.ActionType `json:"action_type"`
	ActionID         string            `json:"action_id"`
	Enabled          bool              `json:"enabled"`
	CreatedAt        time.Time         `json:"created_at"`
}

func toResponse(r *domain.Rule) *RuleResponse {
	return &RuleResponse{
		ID:               r.ID,
		Name:             r.Name,
		TriggerType:      r.TriggerType,
		TriggerID:        r.TriggerID,
		TriggerThreshold: r.TriggerThreshold,
		ActionType:       r.ActionType,
		ActionID:         r.ActionID,
		Enabled:          r.Enabled,
		CreatedAt:        r.CreatedAt,
	}
}

// CreateRule cria uma nova regra de automação
func (e *RulesEngine) CreateRule(req CreateRuleRequest) (*RuleResponse, error) {
	if req.Name == "" {
		return nil, domain.ErrInvalidRule
	}
	rule := domain.NewRule(uuid.New().String(), req.Name, req.OwnerID, req.TriggerType, req.TriggerID, req.TriggerThreshold, req.ActionType, req.ActionID)
	if err := e.ruleRepo.Save(rule); err != nil {
		return nil, err
	}

	return toResponse(rule), nil
}

// UpdateRule atualiza uma regra existente
func (e *RulesEngine) UpdateRule(ruleID, userID string, req CreateRuleRequest) (*RuleResponse, error) {
	rule, err := e.ruleRepo.FindByID(ruleID)
	if err != nil {
		return nil, domain.ErrRuleNotFound
	}
	if !rule.BelongsTo(userID) {
		return nil, domain.ErrRuleAccessDenied
	}
	if req.Name == "" {
		return nil, domain.ErrInvalidRule
	}

	// Atualiza campos editáveis
	rule.Name = req.Name
	rule.TriggerType = req.TriggerType
	rule.TriggerID = req.TriggerID
	rule.TriggerThreshold = req.TriggerThreshold
	rule.ActionType = req.ActionType
	rule.ActionID = req.ActionID
	rule.UpdatedAt = time.Now()

	if err := e.ruleRepo.Save(rule); err != nil {
		return nil, err
	}
	return toResponse(rule), nil
}

// ListRules retorna as regras de um usuário
func (e *RulesEngine) ListRules(ownerID string) ([]*RuleResponse, error) {
	rules, err := e.ruleRepo.FindByOwnerID(ownerID)
	if err != nil {
		return nil, err
	}

	var responses []*RuleResponse
	for _, r := range rules {
		responses = append(responses, toResponse(r))
	}
	return responses, nil
}

// DeleteRule remove uma regra
func (e *RulesEngine) DeleteRule(ruleID, userID string) error {
	rule, err := e.ruleRepo.FindByID(ruleID)
	if err != nil {
		return domain.ErrRuleNotFound
	}

	if !rule.BelongsTo(userID) {
		return domain.ErrRuleAccessDenied
	}

	return e.ruleRepo.Delete(ruleID)
}

// EvaluateMotionTrigger avalia regras para um trigger de movimento
func (e *RulesEngine) EvaluateMotionTrigger(sensorID string) {
	rules, err := e.ruleRepo.FindByTrigger(domain.TriggerMotion, sensorID)
	if err != nil {
		log.Printf("[RulesEngine] Error finding rules for sensor %s: %v", sensorID, err)
		return
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		log.Printf("[RulesEngine] Executing rule '%s': %s → %s(%s)", rule.Name, rule.TriggerType, rule.ActionType, rule.ActionID)
		e.executeAction(rule)

		// Se a ação for ligar luz, registra último movimento e agenda desligamento automático
		if rule.ActionType == domain.ActionTurnOnLight && rule.ActionID != "" {
			e.scheduleMotionAutoOff(rule.ActionID)
		}
	}
}

// scheduleMotionAutoOff agenda o desligamento automático de uma luz após motionOffTimeout,
// se nenhum novo movimento tiver ocorrido para essa luz nesse intervalo.
func (e *RulesEngine) scheduleMotionAutoOff(lightID string) {
	now := time.Now()
	e.mu.Lock()
	e.lastMotionByLight[lightID] = now
	timeout := e.motionOffTimeout
	e.mu.Unlock()

	go func(scheduledAt time.Time) {
		time.Sleep(timeout)

		e.mu.RLock()
		last, ok := e.lastMotionByLight[lightID]
		e.mu.RUnlock()

		// Se houve novo movimento depois desse schedule, não desliga
		if !ok || last.After(scheduledAt) {
			return
		}

		// Desligar a luz após inatividade
		log.Printf("[RulesEngine] No motion for %s on light %s, turning it off", timeout, lightID)
		fakeRule := &domain.Rule{
			Name:       fmt.Sprintf("auto-off-after-motion-%s", lightID),
			ActionType: domain.ActionTurnOffLight,
			ActionID:   lightID,
			Enabled:    true,
		}
		e.executeAction(fakeRule)
	}(now)
}

// EvaluateLightTrigger avalia regras para um trigger de luminosidade baixa (value = % atual, 0 = escuro, 100 = claro)
func (e *RulesEngine) EvaluateLightTrigger(sensorID string, value float64) {
	rules, err := e.ruleRepo.FindByTrigger(domain.TriggerLight, sensorID)
	if err != nil {
		log.Printf("[RulesEngine] Error finding rules for light sensor %s: %v", sensorID, err)
		return
	}

	threshold := domain.DefaultLightThreshold
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		threshold = rule.LightThreshold()
		if value >= threshold {
			continue // Luz não está "baixa" para esta regra
		}
		log.Printf("[RulesEngine] Executing light rule '%s': luz %.1f%% < %.0f%% → %s(%s)", rule.Name, value, threshold, rule.ActionType, rule.ActionID)
		e.executeAction(rule)
	}
}

// EvaluateLightHighTrigger avalia regras para um trigger de luminosidade alta (value = % atual, 0 = escuro, 100 = claro)
func (e *RulesEngine) EvaluateLightHighTrigger(sensorID string, value float64) {
	rules, err := e.ruleRepo.FindByTrigger(domain.TriggerLightHigh, sensorID)
	if err != nil {
		log.Printf("[RulesEngine] Error finding HIGH-light rules for sensor %s: %v", sensorID, err)
		return
	}

	threshold := domain.DefaultLightThreshold
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		threshold = rule.LightThreshold()
		if value <= threshold {
			continue // Luz não está "alta" para esta regra
		}
		log.Printf("[RulesEngine] Executing HIGH-light rule '%s': luz %.1f%% > %.0f%% → %s(%s)", rule.Name, value, threshold, rule.ActionType, rule.ActionID)
		e.executeAction(rule)
	}
}

func (e *RulesEngine) executeAction(rule *domain.Rule) {
	var url string
	var body any
	switch rule.ActionType {
	case domain.ActionTurnOnLight:
		url = fmt.Sprintf("%s/lights/%s/on", e.lightingBaseURL, rule.ActionID)
	case domain.ActionTurnOffLight:
		url = fmt.Sprintf("%s/lights/%s/off", e.lightingBaseURL, rule.ActionID)
	case domain.ActionTriggerAlarm:
		url = fmt.Sprintf("%s/alarms/trigger", e.securityBaseURL)
		// Quando a ação é acionar alarme, propagamos o tipo de trigger da regra
		// para que o security-service consiga distinguir "motion" de "manual".
		// Usamos o trigger_id da regra como sensor_id quando fizer sentido.
		body = map[string]any{
			"trigger_type": string(rule.TriggerType),
			"sensor_id":    rule.TriggerID,
			"location":     rule.Name,
		}
	default:
		log.Printf("[RulesEngine] Unknown action type: %s", rule.ActionType)
		return
	}

	client := &http.Client{Timeout: 3 * time.Second}
	var bodyReader io.Reader = http.NoBody
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			log.Printf("[RulesEngine] Failed to marshal action body: %v", err)
			return
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		log.Printf("[RulesEngine] Failed to create request: %v", err)
		return
	}
	req.Header.Set("X-Device-Key", e.deviceAPIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[RulesEngine] Failed to execute action %s: %v", url, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[RulesEngine] Action executed: %s → HTTP %d", url, resp.StatusCode)
}

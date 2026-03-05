package domain

import (
	"errors"
	"time"
)

var (
	ErrRuleNotFound     = errors.New("rule not found")
	ErrInvalidRule      = errors.New("invalid rule")
	ErrRuleAccessDenied = errors.New("access denied to rule")
)

// TriggerType define o tipo de gatilho de uma regra
type TriggerType string

const (
	TriggerMotion    TriggerType = "motion"
	TriggerLight     TriggerType = "light_low"
	TriggerLightHigh TriggerType = "light_high"
	TriggerSchedule  TriggerType = "schedule"
)

// ActionType define o tipo de ação de uma regra
type ActionType string

const (
	ActionTurnOnLight  ActionType = "turn_on_light"
	ActionTurnOffLight ActionType = "turn_off_light"
	ActionTriggerAlarm ActionType = "trigger_alarm"
)

// Rule representa uma regra de automação
type Rule struct {
	ID                string
	Name              string
	OwnerID           string
	TriggerType       TriggerType
	TriggerID         string  // ID do sensor que dispara
	TriggerThreshold  float64 // Para light_low: considerar "luz baixa" abaixo deste % (0-100). 0 = usar padrão 30
	ActionType        ActionType
	ActionID          string // ID do dispositivo/luz a ser acionado (ex: light-001)
	Enabled           bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// DefaultLightThreshold é o limiar padrão de "luz baixa" quando não definido na regra (%)
const DefaultLightThreshold = 30.0

// NewRule cria uma nova regra de automação
func NewRule(id, name, ownerID string, triggerType TriggerType, triggerID string, triggerThreshold float64, actionType ActionType, actionID string) *Rule {
	now := time.Now()
	if triggerThreshold <= 0 {
		triggerThreshold = DefaultLightThreshold
	}
	return &Rule{
		ID:               id,
		Name:             name,
		OwnerID:          ownerID,
		TriggerType:      triggerType,
		TriggerID:        triggerID,
		TriggerThreshold: triggerThreshold,
		ActionType:       actionType,
		ActionID:         actionID,
		Enabled:          true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// BelongsTo verifica se a regra pertence ao usuário
func (r *Rule) BelongsTo(userID string) bool {
	return r.OwnerID == "*" || r.OwnerID == userID
}

// LightThreshold retorna o limiar de luminosidade para regras light_low (0-100%). Se não definido, usa DefaultLightThreshold.
func (r *Rule) LightThreshold() float64 {
	if r.TriggerThreshold > 0 && r.TriggerThreshold <= 100 {
		return r.TriggerThreshold
	}
	return DefaultLightThreshold
}

// RuleRepository define o contrato para persistência de regras
type RuleRepository interface {
	Save(rule *Rule) error
	FindByID(id string) (*Rule, error)
	FindByOwnerID(ownerID string) ([]*Rule, error)
	FindByTrigger(triggerType TriggerType, triggerID string) ([]*Rule, error)
	Delete(id string) error
}

package domain

import "errors"

// ErrInvalidAlarmTriggerType indica que o tipo de gatilho do alarme é desconhecido.
var ErrInvalidAlarmTriggerType = errors.New("invalid alarm trigger type: expected \"motion\" or \"manual\"")

type AlarmTriggerType struct {
	value string
}

var (
	TriggerMotion = AlarmTriggerType{value: "motion"}
	TriggerManual = AlarmTriggerType{value: "manual"}
)

// NewAlarmTriggerType valida e constrói um AlarmTriggerType a partir de uma string.
func NewAlarmTriggerType(raw string) (AlarmTriggerType, error) {
	switch raw {
	case "motion":
		return TriggerMotion, nil
	case "manual":
		return TriggerManual, nil
	default:
		return AlarmTriggerType{}, ErrInvalidAlarmTriggerType
	}
}

// String retorna o valor string do tipo de gatilho.
func (t AlarmTriggerType) String() string { return t.value }

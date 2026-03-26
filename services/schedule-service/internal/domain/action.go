package domain

// ActionTarget define o mecanismo de dispatch da ação
type ActionTarget string

const (
	ActionTargetNATS ActionTarget = "nats"
	ActionTargetHTTP ActionTarget = "http"
)

// ActionType define o tipo de ação a ser executada
type ActionType string

const (
	ActionTurnOnLight  ActionType = "turn_on_light"
	ActionTurnOffLight ActionType = "turn_off_light"
	ActionTriggerAlarm ActionType = "trigger_alarm"
	ActionSilenceAlarm ActionType = "silence_alarm"
)

// Action é o value object que descreve a ação a ser executada pelo agendamento
type Action struct {
	Target   ActionTarget
	Type     ActionType
	DeviceID string
	Payload  string
}

// NewAction cria um novo Action com validação do target.
func NewAction(target, actionType, deviceID string) (*Action, error) {
	if target != string(ActionTargetNATS) && target != string(ActionTargetHTTP) {
		return nil, ErrInvalidAction
	}
	return &Action{
		Target:   ActionTarget(target),
		Type:     ActionType(actionType),
		DeviceID: deviceID,
	}, nil
}

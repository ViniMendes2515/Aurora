package domain

// ScheduleExecutedEvent é o domain event emitido quando um agendamento é executado.
type ScheduleExecutedEvent struct {
	ScheduleID     string `json:"schedule_id"`
	ScheduleName   string `json:"schedule_name"`
	OwnerID        string `json:"owner_id"`
	ActionType     string `json:"action_type"`
	ActionDeviceID string `json:"action_device_id"`
	ExecutedAt     string `json:"executed_at"`
}

func (e *ScheduleExecutedEvent) Topic() string {
	return "schedule.executed"
}

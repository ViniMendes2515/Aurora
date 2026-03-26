package domain

import "time"

// ExecutionStatus define o status de uma execucao de agendamento
type ExecutionStatus string

const (
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusFailed  ExecutionStatus = "failed"
	ExecutionStatusSkipped ExecutionStatus = "skipped"
)

// ScheduleExecution representa o registro de uma execucao de agendamento
type ScheduleExecution struct {
	ID           string
	ScheduleID   string
	ExecutedAt   time.Time
	Status       ExecutionStatus
	ErrorMessage string
}

// NewScheduleExecution cria um novo registro de execucao de agendamento
func NewScheduleExecution(id, scheduleID string, status ExecutionStatus, errorMessage string) *ScheduleExecution {
	return &ScheduleExecution{
		ID:           id,
		ScheduleID:   scheduleID,
		ExecutedAt:   time.Now(),
		Status:       status,
		ErrorMessage: errorMessage,
	}
}

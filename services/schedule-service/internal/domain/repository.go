package domain

// ScheduleRepository define o contrato para persistencia de agendamentos
type ScheduleRepository interface {
	Save(schedule *Schedule) error
	FindByID(id string) (*Schedule, error)
	FindByOwnerID(ownerID string) ([]*Schedule, error)
	FindAllEnabled() ([]*Schedule, error)
	Delete(id string) error
}

// ScheduleExecutionRepository define o contrato para persistencia de execucoes
type ScheduleExecutionRepository interface {
	Save(execution *ScheduleExecution) error
	FindByScheduleID(scheduleID string, limit int) ([]*ScheduleExecution, error)
}

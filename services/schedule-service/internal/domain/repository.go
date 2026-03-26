package domain

import "context"

// ScheduleRepository define o contrato para persistencia de agendamentos
type ScheduleRepository interface {
	Save(ctx context.Context, schedule *Schedule) error
	FindByID(ctx context.Context, id string) (*Schedule, error)
	FindByOwnerID(ctx context.Context, ownerID string) ([]*Schedule, error)
	FindAllEnabled(ctx context.Context) ([]*Schedule, error)
	Delete(ctx context.Context, id string) error
}

// ScheduleExecutionRepository define o contrato para persistencia de execucoes
type ScheduleExecutionRepository interface {
	Save(ctx context.Context, execution *ScheduleExecution) error
	FindByScheduleID(ctx context.Context, scheduleID string, limit int) ([]*ScheduleExecution, error)
}

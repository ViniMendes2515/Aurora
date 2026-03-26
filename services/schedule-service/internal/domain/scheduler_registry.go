package domain

// SchedulerRegistry define o contrato para registro de agendamentos no scheduler
type SchedulerRegistry interface {
	Register(schedule *Schedule) error
	Unregister(scheduleID string)
}

package domain

// ActionDispatcher define o contrato para despacho de acoes de agendamento
type ActionDispatcher interface {
	Dispatch(schedule *Schedule) error
}

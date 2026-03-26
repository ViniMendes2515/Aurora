package domain

import "context"

// ActionDispatcher define o contrato para despacho de acoes de agendamento
type ActionDispatcher interface {
	Dispatch(ctx context.Context, schedule *Schedule) error
}

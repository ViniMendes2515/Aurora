package dispatcher

import (
	"context"
	"log"

	"aurora/services/schedule-service/internal/domain"
)

// CompositeDispatcher routes dispatch calls to the appropriate dispatcher
// based on the action target, with NATS-to-HTTP fallback support.
type CompositeDispatcher struct {
	natsDispatcher domain.ActionDispatcher
	httpDispatcher domain.ActionDispatcher
}

// NewCompositeDispatcher creates a new CompositeDispatcher with the given
// NATS and HTTP dispatchers.
func NewCompositeDispatcher(natsDispatcher, httpDispatcher domain.ActionDispatcher) *CompositeDispatcher {
	return &CompositeDispatcher{
		natsDispatcher: natsDispatcher,
		httpDispatcher: httpDispatcher,
	}
}

// Dispatch routes the schedule to the appropriate dispatcher based on the
// action target. For NATS targets, it falls back to HTTP if NATS fails.
func (d *CompositeDispatcher) Dispatch(ctx context.Context, schedule *domain.Schedule) error {
	switch schedule.Action.Target {
	case domain.ActionTargetNATS:
		// Temporariamente tratamos ações com target NATS como comandos HTTP.
		log.Printf("[Scheduler] Action target nats recebido para schedule %s; executando via HTTP", schedule.ID)
		return d.httpDispatcher.Dispatch(ctx, schedule)

	case domain.ActionTargetHTTP:
		return d.httpDispatcher.Dispatch(ctx, schedule)

	default:
		return domain.ErrInvalidAction
	}
}

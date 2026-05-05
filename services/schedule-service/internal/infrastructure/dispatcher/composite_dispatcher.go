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
// action target. After a successful dispatch, publishes a schedule.executed event via NATS.
func (d *CompositeDispatcher) Dispatch(ctx context.Context, schedule *domain.Schedule) error {
	var err error
	switch schedule.Action.Target {
	case domain.ActionTargetNATS:
		// Temporariamente tratamos ações com target NATS como comandos HTTP.
		log.Printf("[Scheduler] Action target nats recebido para schedule %s; executando via HTTP", schedule.ID)
		err = d.httpDispatcher.Dispatch(ctx, schedule)
	case domain.ActionTargetHTTP:
		err = d.httpDispatcher.Dispatch(ctx, schedule)
	default:
		return domain.ErrInvalidAction
	}

	if err != nil {
		return err
	}

	if pubErr := d.natsDispatcher.Dispatch(ctx, schedule); pubErr != nil {
		log.Printf("[Scheduler] Falha ao publicar schedule.executed para schedule %s: %v", schedule.ID, pubErr)
	} else {
		log.Printf("[Scheduler] schedule.executed publicado para schedule %s", schedule.ID)
	}
	return nil
}

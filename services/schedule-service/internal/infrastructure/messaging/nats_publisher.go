package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"aurora/services/schedule-service/internal/domain"
)

// NATSPublisher implementa domain.ActionDispatcher publicando eventos no NATS
type NATSPublisher struct {
	conn *NATSConnection
}

// NewNATSPublisher cria um novo NATSPublisher a partir de uma conexão existente
func NewNATSPublisher(conn *NATSConnection) *NATSPublisher {
	return &NATSPublisher{conn: conn}
}

// Dispatch constrói um domain event a partir do Schedule e publica no NATS.
// Realiza até 3 tentativas com backoff exponencial: 100ms, 500ms, 2s.
func (p *NATSPublisher) Dispatch(ctx context.Context, schedule *domain.Schedule) error {
	event := &domain.ScheduleExecutedEvent{
		ScheduleID:     schedule.ID,
		ScheduleName:   schedule.Name,
		OwnerID:        schedule.OwnerID,
		ActionType:     string(schedule.Action.Type),
		ActionDeviceID: schedule.Action.DeviceID,
		ExecutedAt:     time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	delays := []time.Duration{100 * time.Millisecond, 500 * time.Millisecond, 2 * time.Second}

	for attempt, delay := range delays {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err = p.conn.GetConnection().Publish(event.Topic(), data); err == nil {
			return nil
		}
		log.Printf("NATS publisher: publish attempt %d failed: %v", attempt+1, err)
		if attempt < len(delays)-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return domain.ErrPublishFailed
}

package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"aurora/services/schedule-service/internal/domain"
)

// NATSPublisher publica eventos de execução de agendamentos no NATS
type NATSPublisher struct {
	conn *nats.Conn
}

// NewNATSPublisher cria uma nova conexão com o NATS e retorna um NATSPublisher
func NewNATSPublisher(url string) (*NATSPublisher, error) {
	conn, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS publisher disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS publisher reconnected to %s", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("NATS publisher connected to %s", url)
	return &NATSPublisher{conn: conn}, nil
}

// scheduleExecutedEvent representa o payload publicado no tópico "schedule.executed"
type scheduleExecutedEvent struct {
	ScheduleID     string `json:"schedule_id"`
	OwnerID        string `json:"owner_id"`
	ActionType     string `json:"action_type"`
	ActionDeviceID string `json:"action_device_id"`
	ExecutedAt     string `json:"executed_at"`
}

// Dispatch publica um evento de execução de agendamento no tópico "schedule.executed".
// Realiza até 3 tentativas com backoff exponencial: 100ms, 500ms, 2s.
func (p *NATSPublisher) Dispatch(ctx context.Context, schedule *domain.Schedule) error {
	payload := scheduleExecutedEvent{
		ScheduleID:     schedule.ID,
		OwnerID:        schedule.OwnerID,
		ActionType:     string(schedule.Action.Type),
		ActionDeviceID: schedule.Action.DeviceID,
		ExecutedAt:     time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	delays := []time.Duration{100 * time.Millisecond, 500 * time.Millisecond, 2 * time.Second}

	for attempt, delay := range delays {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err = p.conn.Publish("schedule.executed", data); err == nil {
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

// Close fecha a conexão com o NATS
func (p *NATSPublisher) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}

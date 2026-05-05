package repository

import (
	"database/sql"

	"aurora/services/security-service/internal/domain"
)

// PostgresAlarmRepository implementa domain.AlarmRepository
type PostgresAlarmRepository struct {
	db *sql.DB
}

// NewPostgresAlarmRepository cria uma nova instância do repositório
func NewPostgresAlarmRepository(db *sql.DB) *PostgresAlarmRepository {
	return &PostgresAlarmRepository{db: db}
}

// Save persiste um evento de alarme
func (r *PostgresAlarmRepository) Save(event *domain.AlarmEvent) error {
	query := `
		INSERT INTO alarm_events (id, user_id, trigger_type, sensor_id, location, status, triggered_at, silenced_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			silenced_at = EXCLUDED.silenced_at
	`
	_, err := r.db.Exec(query,
		event.ID,
		event.UserID,
		event.TriggerType.String(),
		event.SensorID,
		event.Location,
		string(event.Status),
		event.TriggeredAt,
		event.SilencedAt,
	)
	return err
}

// FindByID busca um evento de alarme pelo ID
func (r *PostgresAlarmRepository) FindByID(id string) (*domain.AlarmEvent, error) {
	query := `
		SELECT id, user_id, trigger_type, sensor_id, location, status, triggered_at, silenced_at
		FROM alarm_events WHERE id = $1
	`
	var triggerTypeStr, status string

	event := &domain.AlarmEvent{}

	err := r.db.QueryRow(query, id).Scan(
		&event.ID, &event.UserID, &triggerTypeStr, &event.SensorID, &event.Location,
		&status, &event.TriggeredAt, &event.SilencedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrAlarmNotFound
	}

	if err != nil {
		return nil, err
	}

	tt, err := domain.NewAlarmTriggerType(triggerTypeStr)
	if err != nil {
		return nil, err
	}

	event.TriggerType = tt
	event.Status = domain.AlarmStatus(status)
	return event, nil
}

// FindRecent retorna os últimos eventos de alarme
func (r *PostgresAlarmRepository) FindRecent(limit int) ([]*domain.AlarmEvent, error) {
	query := `
		SELECT id, user_id, trigger_type, sensor_id, location, status, triggered_at, silenced_at
		FROM alarm_events
		ORDER BY triggered_at DESC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.AlarmEvent
	for rows.Next() {
		event := &domain.AlarmEvent{}
		var triggerTypeStr, status string

		if err := rows.Scan(
			&event.ID, &event.UserID, &triggerTypeStr, &event.SensorID, &event.Location,
			&status, &event.TriggeredAt, &event.SilencedAt,
		); err != nil {
			return nil, err
		}

		tt, err := domain.NewAlarmTriggerType(triggerTypeStr)
		if err != nil {
			return nil, err
		}

		event.TriggerType = tt
		event.Status = domain.AlarmStatus(status)
		events = append(events, event)
	}
	return events, nil
}

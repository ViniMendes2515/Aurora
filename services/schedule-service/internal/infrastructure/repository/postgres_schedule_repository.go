package repository

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"aurora/services/schedule-service/internal/domain"
)

// PostgresScheduleRepository implementa domain.ScheduleRepository
type PostgresScheduleRepository struct {
	db *sql.DB
}

// NewPostgresScheduleRepository cria uma nova instância do repositório de agendamentos
func NewPostgresScheduleRepository(db *sql.DB) *PostgresScheduleRepository {
	return &PostgresScheduleRepository{db: db}
}

// Save persiste um agendamento (insert ou update)
func (r *PostgresScheduleRepository) Save(ctx context.Context, schedule *domain.Schedule) error {
	query := `
		INSERT INTO schedules (
			id, owner_id, name, description, schedule_type, cron_expr, run_at,
			days_filter, timezone, action_target, action_type, action_device_id,
			action_payload, enabled, created_at, updated_at, last_run_at, next_run_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18
		)
		ON CONFLICT (id) DO UPDATE SET
			name             = EXCLUDED.name,
			description      = EXCLUDED.description,
			schedule_type    = EXCLUDED.schedule_type,
			cron_expr        = EXCLUDED.cron_expr,
			run_at           = EXCLUDED.run_at,
			days_filter      = EXCLUDED.days_filter,
			timezone         = EXCLUDED.timezone,
			action_target    = EXCLUDED.action_target,
			action_type      = EXCLUDED.action_type,
			action_device_id = EXCLUDED.action_device_id,
			action_payload   = EXCLUDED.action_payload,
			enabled          = EXCLUDED.enabled,
			updated_at       = EXCLUDED.updated_at,
			last_run_at      = EXCLUDED.last_run_at,
			next_run_at      = EXCLUDED.next_run_at
	`
	_, err := r.db.ExecContext(ctx, query,
		schedule.ID,
		schedule.OwnerID,
		schedule.Name,
		schedule.Description,
		string(schedule.ScheduleType),
		schedule.CronExpr,
		schedule.RunAt,
		schedule.DaysFilter,
		schedule.Timezone,
		string(schedule.Action.Target),
		string(schedule.Action.Type),
		schedule.Action.DeviceID,
		schedule.Action.Payload,
		schedule.Enabled,
		schedule.CreatedAt,
		schedule.UpdatedAt,
		schedule.LastRunAt,
		schedule.NextRunAt,
	)
	return err
}

// FindByID busca um agendamento pelo ID
func (r *PostgresScheduleRepository) FindByID(ctx context.Context, id string) (*domain.Schedule, error) {
	query := `
		SELECT
			id, owner_id, name, description, schedule_type, cron_expr, run_at,
			days_filter, timezone, action_target, action_type, action_device_id,
			action_payload, enabled, created_at, updated_at, last_run_at, next_run_at
		FROM schedules
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	schedule, err := r.scanSchedule(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrScheduleNotFound
	}
	return schedule, err
}

// FindByOwnerID busca agendamentos pelo ID do proprietário
func (r *PostgresScheduleRepository) FindByOwnerID(ctx context.Context, ownerID string) ([]*domain.Schedule, error) {
	query := `
		SELECT
			id, owner_id, name, description, schedule_type, cron_expr, run_at,
			days_filter, timezone, action_target, action_type, action_device_id,
			action_payload, enabled, created_at, updated_at, last_run_at, next_run_at
		FROM schedules
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanScheduleRows(rows)
}

// FindAllEnabled busca todos os agendamentos habilitados
func (r *PostgresScheduleRepository) FindAllEnabled(ctx context.Context) ([]*domain.Schedule, error) {
	query := `
		SELECT
			id, owner_id, name, description, schedule_type, cron_expr, run_at,
			days_filter, timezone, action_target, action_type, action_device_id,
			action_payload, enabled, created_at, updated_at, last_run_at, next_run_at
		FROM schedules
		WHERE enabled = true
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanScheduleRows(rows)
}

// Delete remove um agendamento
func (r *PostgresScheduleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM schedules WHERE id = $1", id)
	return err
}

// scanner is a common interface for *sql.Row and *sql.Rows
type scanner interface {
	Scan(dest ...interface{}) error
}

// scanSchedule lê uma linha e converte para *domain.Schedule
func (r *PostgresScheduleRepository) scanSchedule(s scanner) (*domain.Schedule, error) {
	var (
		scheduleType  string
		cronExpr      string
		runAt         sql.NullTime
		actionTarget  string
		actionType    string
		actionDevID   string
		actionPayload string
		lastRunAt     sql.NullTime
		nextRunAt     sql.NullTime
	)

	schedule := &domain.Schedule{}

	err := s.Scan(
		&schedule.ID,
		&schedule.OwnerID,
		&schedule.Name,
		&schedule.Description,
		&scheduleType,
		&cronExpr,
		&runAt,
		&schedule.DaysFilter,
		&schedule.Timezone,
		&actionTarget,
		&actionType,
		&actionDevID,
		&actionPayload,
		&schedule.Enabled,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
		&lastRunAt,
		&nextRunAt,
	)
	if err != nil {
		return nil, err
	}

	schedule.ScheduleType = domain.ScheduleType(scheduleType)
	schedule.CronExpr = cronExpr

	if runAt.Valid {
		t := runAt.Time
		schedule.RunAt = &t
	}

	schedule.Action = domain.Action{
		Target:   domain.ActionTarget(actionTarget),
		Type:     domain.ActionType(actionType),
		DeviceID: actionDevID,
		Payload:  actionPayload,
	}

	if lastRunAt.Valid {
		t := lastRunAt.Time
		schedule.LastRunAt = &t
	}
	if nextRunAt.Valid {
		t := nextRunAt.Time
		schedule.NextRunAt = &t
	}

	return schedule, nil
}

// scanScheduleRows lê todas as linhas de um *sql.Rows
func (r *PostgresScheduleRepository) scanScheduleRows(rows *sql.Rows) ([]*domain.Schedule, error) {
	var schedules []*domain.Schedule
	for rows.Next() {
		schedule, err := r.scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return schedules, nil
}

// PostgresScheduleExecutionRepository implementa domain.ScheduleExecutionRepository
type PostgresScheduleExecutionRepository struct {
	db *sql.DB
}

// NewPostgresScheduleExecutionRepository cria uma nova instância do repositório de execuções
func NewPostgresScheduleExecutionRepository(db *sql.DB) *PostgresScheduleExecutionRepository {
	return &PostgresScheduleExecutionRepository{db: db}
}

// Save persiste uma execução de agendamento
func (r *PostgresScheduleExecutionRepository) Save(ctx context.Context, execution *domain.ScheduleExecution) error {
	query := `
		INSERT INTO schedule_executions (id, schedule_id, executed_at, status, error_message)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		execution.ID,
		execution.ScheduleID,
		execution.ExecutedAt,
		string(execution.Status),
		execution.ErrorMessage,
	)
	return err
}

// FindByScheduleID busca execuções de um agendamento com limite
func (r *PostgresScheduleExecutionRepository) FindByScheduleID(ctx context.Context, scheduleID string, limit int) ([]*domain.ScheduleExecution, error) {
	query := `
		SELECT id, schedule_id, executed_at, status, error_message
		FROM schedule_executions
		WHERE schedule_id = $1
		ORDER BY executed_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, scheduleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []*domain.ScheduleExecution
	for rows.Next() {
		execution := &domain.ScheduleExecution{}
		var status string
		var executedAt time.Time
		if err := rows.Scan(
			&execution.ID,
			&execution.ScheduleID,
			&executedAt,
			&status,
			&execution.ErrorMessage,
		); err != nil {
			return nil, err
		}
		execution.ExecutedAt = executedAt
		execution.Status = domain.ExecutionStatus(status)
		executions = append(executions, execution)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return executions, nil
}

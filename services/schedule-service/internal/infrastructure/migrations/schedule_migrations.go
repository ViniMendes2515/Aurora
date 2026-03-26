package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// RunScheduleMigrations executa as migrations do schedule-service.
// Cria as tabelas schedules e schedule_executions com todos os campos necessários para a v1.
// Inclui timezone, schedule_type, days_filter, run_at, last_run_at e next_run_at desde o início
// para evitar ALTER TABLE migrations retroativas nas fases seguintes.
func RunScheduleMigrations(db *sql.DB) error {
	migration := `
		CREATE TABLE IF NOT EXISTS schedules (
			id              VARCHAR(36) PRIMARY KEY,
			owner_id        VARCHAR(255) NOT NULL,
			name            VARCHAR(255) NOT NULL,
			description     TEXT NOT NULL DEFAULT '',
			schedule_type   VARCHAR(20) NOT NULL,
			cron_expr       VARCHAR(100) NOT NULL DEFAULT '',
			run_at          TIMESTAMP,
			days_filter     INTEGER NOT NULL DEFAULT 0,
			timezone        VARCHAR(64) NOT NULL DEFAULT 'UTC',
			action_target   VARCHAR(10) NOT NULL,
			action_type     VARCHAR(50) NOT NULL,
			action_device_id VARCHAR(255) NOT NULL DEFAULT '',
			action_payload  TEXT NOT NULL DEFAULT '',
			enabled         BOOLEAN NOT NULL DEFAULT true,
			created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
			last_run_at     TIMESTAMP,
			next_run_at     TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_schedules_owner_id ON schedules(owner_id);

		CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);

		CREATE TABLE IF NOT EXISTS schedule_executions (
			id            VARCHAR(36) PRIMARY KEY,
			schedule_id   VARCHAR(36) NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
			executed_at   TIMESTAMP NOT NULL DEFAULT NOW(),
			status        VARCHAR(20) NOT NULL,
			error_message TEXT NOT NULL DEFAULT ''
		);

		CREATE INDEX IF NOT EXISTS idx_executions_schedule_id ON schedule_executions(schedule_id);
	`

	_, err := db.Exec(migration)
	if err != nil {
		return fmt.Errorf("failed to run schedule migrations: %w", err)
	}

	log.Println("Schedule service migrations completed successfully")
	return nil
}

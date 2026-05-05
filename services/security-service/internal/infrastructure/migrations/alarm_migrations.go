package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// RunAlarmMigrations executa as migrations do security-service
func RunAlarmMigrations(db *sql.DB) error {
	migration := `
		CREATE TABLE IF NOT EXISTS alarm_events (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL DEFAULT '',
			trigger_type VARCHAR(50) NOT NULL,
			sensor_id VARCHAR(255) NOT NULL DEFAULT '',
			location VARCHAR(255) NOT NULL DEFAULT '',
			status VARCHAR(20) NOT NULL DEFAULT 'triggered',
			triggered_at TIMESTAMP NOT NULL DEFAULT NOW(),
			silenced_at TIMESTAMP
		);

		ALTER TABLE alarm_events ADD COLUMN IF NOT EXISTS user_id VARCHAR(255) NOT NULL DEFAULT '';

		CREATE INDEX IF NOT EXISTS idx_alarm_events_status ON alarm_events(status);
		CREATE INDEX IF NOT EXISTS idx_alarm_events_triggered_at ON alarm_events(triggered_at);
	`

	_, err := db.Exec(migration)
	if err != nil {
		return fmt.Errorf("failed to run alarm migrations: %w", err)
	}

	log.Println("Security service migrations completed successfully")
	return nil
}

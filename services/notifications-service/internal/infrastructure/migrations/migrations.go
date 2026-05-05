package migrations

import "database/sql"

// RunNotificationMigrations cria as tabelas necessarias para o notifications-service
func RunNotificationMigrations(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS telegram_preferences (
			id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id       UUID         NOT NULL UNIQUE,
			chat_id       BIGINT       NOT NULL,
			enabled_types TEXT[]       NOT NULL DEFAULT ARRAY['motion_detected','light_on','light_off','alarm_triggered','schedule_executed'],
			active        BOOLEAN      NOT NULL DEFAULT TRUE,
			linked_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS telegram_link_tokens (
			token      VARCHAR(20)  PRIMARY KEY,
			user_id    UUID         NOT NULL,
			expires_at TIMESTAMPTZ  NOT NULL
		);

		CREATE TABLE IF NOT EXISTS notifications (
			id         UUID         PRIMARY KEY,
			user_id    VARCHAR(255) NOT NULL,
			type       VARCHAR(50)  NOT NULL,
			title      VARCHAR(255) NOT NULL,
			message    TEXT         NOT NULL,
			sensor_id  VARCHAR(255),
			location   VARCHAR(255),
			read       BOOLEAN      NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_notifications_user_id
			ON notifications(user_id);

		CREATE INDEX IF NOT EXISTS idx_notifications_user_read
			ON notifications(user_id, read);
	`)
	return err
}

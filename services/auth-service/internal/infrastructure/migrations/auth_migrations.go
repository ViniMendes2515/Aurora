package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// RunAuthMigrations executa as migrations específicas do auth-service
func RunAuthMigrations(db *sql.DB) error {
	migration := `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES (
			'00000000-0000-0000-0000-000000000001',
			'admin@aurora.local',
			'$2a$10$6WO1jUPrFGoUlEEf6v7fo.aKowknXNK2beWpbFXhf4vLATz/bbw7G',
			NOW(),
			NOW()
		)
		ON CONFLICT (id) DO NOTHING;
	`

	_, err := db.Exec(migration)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Auth service migrations completed successfully")
	return nil
}

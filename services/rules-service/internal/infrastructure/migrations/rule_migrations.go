package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// RunRuleMigrations executa as migrations do rules-service
func RunRuleMigrations(db *sql.DB) error {
	migration := `
		CREATE TABLE IF NOT EXISTS rules (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			owner_id VARCHAR(255) NOT NULL DEFAULT '*',
			trigger_type VARCHAR(50) NOT NULL,
			trigger_id VARCHAR(255) NOT NULL DEFAULT '',
			trigger_threshold DECIMAL(5,2) NOT NULL DEFAULT 30,
			action_type VARCHAR(50) NOT NULL,
			action_id VARCHAR(255) NOT NULL DEFAULT '',
			enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_rules_owner_id ON rules(owner_id);
		CREATE INDEX IF NOT EXISTS idx_rules_trigger ON rules(trigger_type, trigger_id);
	`

	_, err := db.Exec(migration)
	if err != nil {
		return fmt.Errorf("failed to run rule migrations: %w", err)
	}

	// Adiciona coluna trigger_threshold se não existir (migração incremental)
	_, _ = db.Exec(`
		ALTER TABLE rules ADD COLUMN IF NOT EXISTS trigger_threshold DECIMAL(5,2) NOT NULL DEFAULT 30;
	`)

	log.Println("Rules service migrations completed successfully")
	return nil
}

// SeedDemoRules insere regras de demonstração se não existirem
func SeedDemoRules(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM rules").Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		log.Println("Rules already seeded, skipping...")
		return nil
	}

	rules := []struct {
		id                 string
		name               string
		ownerID            string
		triggerType        string
		triggerID          string
		triggerThreshold   float64
		actionType         string
		actionID           string
	}{
		{
			"rule-001",
			"Acender luz ao detectar presença",
			"*",
			"motion",
			"esp32-pir-001",
			30,
			"turn_on_light",
			"light-001",
		},
		{
			"rule-002",
			"Acender luz quando escurecer",
			"*",
			"light_low",
			"esp32-ldr-001",
			30,
			"turn_on_light",
			"light-001",
		},
	}

	query := `
		INSERT INTO rules (id, name, owner_id, trigger_type, trigger_id, trigger_threshold, action_type, action_id, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, NOW(), NOW())
	`

	for _, r := range rules {
		if _, err := db.Exec(query, r.id, r.name, r.ownerID, r.triggerType, r.triggerID, r.triggerThreshold, r.actionType, r.actionID); err != nil {
			return fmt.Errorf("failed to seed rule %s: %w", r.id, err)
		}
	}

	log.Printf("Seeded %d demo rules", len(rules))
	return nil
}

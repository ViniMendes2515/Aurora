package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// RunLightMigrations executa as migrations do lighting-service
func RunLightMigrations(db *sql.DB) error {
	migration := `
		CREATE TABLE IF NOT EXISTS lights (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			location VARCHAR(255) NOT NULL,
			owner_id VARCHAR(255) NOT NULL DEFAULT '*',
			state VARCHAR(10) NOT NULL DEFAULT 'off',
			device_id VARCHAR(255) NOT NULL DEFAULT 'esp32-main',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_lights_owner_id ON lights(owner_id);

		CREATE TABLE IF NOT EXISTS registered_devices (
			id VARCHAR(255) PRIMARY KEY,
			ip_address VARCHAR(50) NOT NULL,
			port INTEGER NOT NULL DEFAULT 80,
			last_seen TIMESTAMP NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`

	_, err := db.Exec(migration)
	if err != nil {
		return fmt.Errorf("failed to run light migrations: %w", err)
	}

	// Atualiza device_id das luzes existentes para incluir o pin do LED
	pinUpdate := `
		UPDATE lights SET device_id = 'esp32-main:1' WHERE id = 'light-001' AND device_id = 'esp32-main';
		UPDATE lights SET device_id = 'esp32-main:2' WHERE id = 'light-002' AND device_id = 'esp32-main';
		UPDATE lights SET device_id = 'esp32-main:3' WHERE id = 'light-003' AND device_id = 'esp32-main';
	`
	if _, err := db.Exec(pinUpdate); err != nil {
		log.Printf("Warning: could not update legacy device_id pins: %v", err)
	}

	log.Println("Lighting service migrations completed successfully")
	return nil
}

// SeedDemoLights insere luzes de demonstração se não existirem
func SeedDemoLights(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM lights").Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		log.Println("Lights already seeded, skipping...")
		return nil
	}

	ownerID := "00000000-0000-0000-0000-000000000001"

	lights := []struct {
		id       string
		name     string
		location string
		ownerID  string
		deviceID string
	}{
		{"light-001", "Luz da Sala", "Sala de Estar", ownerID, "esp32-main:1"},
		{"light-002", "Luz do Corredor", "Corredor", ownerID, "esp32-main:2"},
		{"light-003", "Luz da Garagem", "Garagem", ownerID, "esp32-main:3"},
		{"light-004", "Luz da Cozinha", "Cozinha", ownerID, "esp32-main:4"},
	}

	query := `
		INSERT INTO lights (id, name, location, owner_id, state, device_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'off', $5, NOW(), NOW())
	`

	for _, l := range lights {
		if _, err := db.Exec(query, l.id, l.name, l.location, l.ownerID, l.deviceID); err != nil {
			return fmt.Errorf("failed to seed light %s: %w", l.id, err)
		}
	}

	log.Printf("Seeded %d demo lights", len(lights))
	return nil
}

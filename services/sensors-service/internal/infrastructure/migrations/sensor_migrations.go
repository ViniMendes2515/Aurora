package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// RunSensorMigrations executa as migrations específicas do sensors-service
func RunSensorMigrations(db *sql.DB) error {
	migration := `
		CREATE TABLE IF NOT EXISTS sensors (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			location VARCHAR(255) NOT NULL,
			owner_id VARCHAR(255) NOT NULL DEFAULT '*',
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_sensors_owner_id ON sensors(owner_id);
		CREATE INDEX IF NOT EXISTS idx_sensors_type ON sensors(type);

		CREATE TABLE IF NOT EXISTS motion_records (
			id VARCHAR(36) PRIMARY KEY,
			sensor_id VARCHAR(36) NOT NULL REFERENCES sensors(id),
			user_id VARCHAR(255) NOT NULL,
			detected_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_motion_records_sensor_id ON motion_records(sensor_id);
		CREATE INDEX IF NOT EXISTS idx_motion_records_detected_at ON motion_records(detected_at);

		CREATE TABLE IF NOT EXISTS light_records (
			id VARCHAR(36) PRIMARY KEY,
			sensor_id VARCHAR(36) NOT NULL REFERENCES sensors(id),
			value DECIMAL(5,2) NOT NULL,
			raw INTEGER NOT NULL,
			recorded_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_light_records_sensor_id ON light_records(sensor_id);
		CREATE INDEX IF NOT EXISTS idx_light_records_recorded_at ON light_records(recorded_at);

		CREATE TABLE IF NOT EXISTS devices (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			ip_address VARCHAR(50) NOT NULL,
			port INTEGER NOT NULL DEFAULT 80,
			last_seen TIMESTAMP NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`

	_, err := db.Exec(migration)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Sensors service migrations completed successfully")
	return nil
}

// SeedDemoSensors insere sensores de demonstração se não existirem
func SeedDemoSensors(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sensors").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Sensors already seeded, skipping...")
		return nil
	}

	ownerID := "00000000-0000-0000-0000-000000000001"

	sensors := []struct {
		id       string
		name     string
		sType    string
		location string
		ownerID  string
	}{
		// Sensores físicos do ESP32
		{"esp32-pir-001", "PIR ESP32 - Sensor de Presença", "motion", "Sala IoT", ownerID},
		{"esp32-ldr-001", "LDR ESP32 - Sensor de Luminosidade", "light", "Sala IoT", ownerID},
	}

	query := `
		INSERT INTO sensors (id, name, type, location, owner_id, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
	`

	for _, s := range sensors {
		_, err := db.Exec(query, s.id, s.name, s.sType, s.location, s.ownerID)
		if err != nil {
			return fmt.Errorf("failed to seed sensor %s: %w", s.id, err)
		}
	}

	log.Printf("Seeded %d sensors (%d demo + 2 ESP32 physical)", len(sensors), len(sensors)-2)
	return nil
}

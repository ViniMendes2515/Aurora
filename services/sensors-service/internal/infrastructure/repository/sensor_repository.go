package repository

import (
	"database/sql"

	"aurora/services/sensors-service/internal/domain"
)

// PostgresSensorRepository implementa domain.SensorRepository usando PostgreSQL
type PostgresSensorRepository struct {
	db *sql.DB
}

// NewPostgresSensorRepository cria uma nova instância do repositório PostgreSQL
func NewPostgresSensorRepository(db *sql.DB) *PostgresSensorRepository {
	return &PostgresSensorRepository{
		db: db,
	}
}

// FindByID busca um sensor pelo ID
func (r *PostgresSensorRepository) FindByID(id string) (*domain.Sensor, error) {
	query := `
		SELECT id, name, type, location, owner_id, active, created_at, updated_at
		FROM sensors
		WHERE id = $1
	`
	sensor := &domain.Sensor{}
	err := r.db.QueryRow(query, id).Scan(
		&sensor.ID,
		&sensor.Name,
		&sensor.Type,
		&sensor.Location,
		&sensor.OwnerID,
		&sensor.Active,
		&sensor.CreatedAt,
		&sensor.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrSensorNotFound
	}
	if err != nil {
		return nil, err
	}
	return sensor, nil
}

// FindByOwnerID busca sensores pelo ID do proprietário
func (r *PostgresSensorRepository) FindByOwnerID(ownerID string) ([]*domain.Sensor, error) {
	query := `
		SELECT id, name, type, location, owner_id, active, created_at, updated_at
		FROM sensors
		WHERE owner_id = $1 OR owner_id = '*'
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sensors []*domain.Sensor
	for rows.Next() {
		sensor := &domain.Sensor{}
		err := rows.Scan(
			&sensor.ID,
			&sensor.Name,
			&sensor.Type,
			&sensor.Location,
			&sensor.OwnerID,
			&sensor.Active,
			&sensor.CreatedAt,
			&sensor.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, sensor)
	}
	return sensors, nil
}

// Save persiste um sensor
func (r *PostgresSensorRepository) Save(sensor *domain.Sensor) error {
	query := `
		INSERT INTO sensors (id, name, type, location, owner_id, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			location = EXCLUDED.location,
			owner_id = EXCLUDED.owner_id,
			active = EXCLUDED.active,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query,
		sensor.ID,
		sensor.Name,
		sensor.Type,
		sensor.Location,
		sensor.OwnerID,
		sensor.Active,
		sensor.CreatedAt,
		sensor.UpdatedAt,
	)
	return err
}

// SaveMotionRecord persiste um registro de movimento
func (r *PostgresSensorRepository) SaveMotionRecord(record *domain.MotionRecord) error {
	query := `
		INSERT INTO motion_records (id, sensor_id, user_id, detected_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(query,
		record.ID,
		record.SensorID,
		record.UserID,
		record.DetectedAt,
	)
	return err
}

// FindAll retorna todos os sensores
func (r *PostgresSensorRepository) FindAll() ([]*domain.Sensor, error) {
	query := `
		SELECT id, name, type, location, owner_id, active, created_at, updated_at
		FROM sensors
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sensors []*domain.Sensor
	for rows.Next() {
		sensor := &domain.Sensor{}
		err := rows.Scan(
			&sensor.ID,
			&sensor.Name,
			&sensor.Type,
			&sensor.Location,
			&sensor.OwnerID,
			&sensor.Active,
			&sensor.CreatedAt,
			&sensor.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, sensor)
	}
	return sensors, nil
}

// GetMotionRecords retorna os registros de movimento de um sensor
func (r *PostgresSensorRepository) GetMotionRecords(sensorID string, limit int) ([]*domain.MotionRecord, error) {
	query := `
		SELECT id, sensor_id, user_id, detected_at
		FROM motion_records
		WHERE sensor_id = $1
		ORDER BY detected_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(query, sensorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.MotionRecord
	for rows.Next() {
		record := &domain.MotionRecord{}
		err := rows.Scan(
			&record.ID,
			&record.SensorID,
			&record.UserID,
			&record.DetectedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// SaveLightRecord persiste um registro de luminosidade
func (r *PostgresSensorRepository) SaveLightRecord(record *domain.LightRecord) error {
	query := `
		INSERT INTO light_records (id, sensor_id, value, raw, recorded_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query,
		record.ID,
		record.SensorID,
		record.Value,
		record.Raw,
		record.RecordedAt,
	)
	return err
}

// GetLightRecords retorna os registros de luminosidade de um sensor
func (r *PostgresSensorRepository) GetLightRecords(sensorID string, limit int) ([]*domain.LightRecord, error) {
	query := `
		SELECT id, sensor_id, value, raw, recorded_at
		FROM light_records
		WHERE sensor_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(query, sensorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.LightRecord
	for rows.Next() {
		record := &domain.LightRecord{}
		err := rows.Scan(
			&record.ID,
			&record.SensorID,
			&record.Value,
			&record.Raw,
			&record.RecordedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

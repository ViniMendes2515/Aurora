package repository

import (
	"database/sql"

	"aurora/services/lighting-service/internal/domain"
)

// PostgresLightRepository implementa domain.LightRepository
type PostgresLightRepository struct {
	db *sql.DB
}

// NewPostgresLightRepository cria uma nova instância do repositório
func NewPostgresLightRepository(db *sql.DB) *PostgresLightRepository {
	return &PostgresLightRepository{db: db}
}

// FindByID busca uma luz pelo ID
func (r *PostgresLightRepository) FindByID(id string) (*domain.Light, error) {
	query := `
		SELECT id, name, location, owner_id, state, device_id, created_at, updated_at
		FROM lights WHERE id = $1
	`
	light := &domain.Light{}
	err := r.db.QueryRow(query, id).Scan(
		&light.ID, &light.Name, &light.Location, &light.OwnerID,
		&light.State, &light.DeviceID, &light.CreatedAt, &light.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrLightNotFound
	}
	if err != nil {
		return nil, err
	}
	return light, nil
}

// FindByOwnerID busca luzes pelo ID do proprietário
func (r *PostgresLightRepository) FindByOwnerID(ownerID string) ([]*domain.Light, error) {
	query := `
		SELECT id, name, location, owner_id, state, device_id, created_at, updated_at
		FROM lights WHERE owner_id = $1 OR owner_id = '*'
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lights []*domain.Light
	for rows.Next() {
		light := &domain.Light{}
		if err := rows.Scan(
			&light.ID, &light.Name, &light.Location, &light.OwnerID,
			&light.State, &light.DeviceID, &light.CreatedAt, &light.UpdatedAt,
		); err != nil {
			return nil, err
		}
		lights = append(lights, light)
	}
	return lights, nil
}

// FindAll retorna todas as luzes
func (r *PostgresLightRepository) FindAll() ([]*domain.Light, error) {
	query := `
		SELECT id, name, location, owner_id, state, device_id, created_at, updated_at
		FROM lights ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lights []*domain.Light
	for rows.Next() {
		light := &domain.Light{}
		if err := rows.Scan(
			&light.ID, &light.Name, &light.Location, &light.OwnerID,
			&light.State, &light.DeviceID, &light.CreatedAt, &light.UpdatedAt,
		); err != nil {
			return nil, err
		}
		lights = append(lights, light)
	}
	return lights, nil
}

// Save persiste uma luz
func (r *PostgresLightRepository) Save(light *domain.Light) error {
	query := `
		INSERT INTO lights (id, name, location, owner_id, state, device_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			location = EXCLUDED.location,
			owner_id = EXCLUDED.owner_id,
			state = EXCLUDED.state,
			device_id = EXCLUDED.device_id,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query,
		light.ID, light.Name, light.Location, light.OwnerID,
		light.State, light.DeviceID, light.CreatedAt, light.UpdatedAt,
	)
	return err
}

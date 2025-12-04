package repository

import (
	"sync"

	"aurora/services/sensors-service/internal/domain"
)

// MemorySensorRepository implementa domain.SensorRepository em memória
type MemorySensorRepository struct {
	sensors       map[string]*domain.Sensor
	motionRecords map[string][]*domain.MotionRecord
	mu            sync.RWMutex
}

// NewMemorySensorRepository cria uma nova instância do repositório em memória
func NewMemorySensorRepository() *MemorySensorRepository {
	repo := &MemorySensorRepository{
		sensors:       make(map[string]*domain.Sensor),
		motionRecords: make(map[string][]*domain.MotionRecord),
	}

	// Inicializar com alguns sensores de exemplo para testes
	repo.seedData()

	return repo
}

// seedData adiciona dados de exemplo para testes
func (r *MemorySensorRepository) seedData() {
	// Sensores de demonstração - usam "*" como ownerID para permitir qualquer usuário
	// Em produção, os sensores teriam ownerIDs específicos
	sensor1 := domain.NewSensor("sensor-001", "Hall Motion Sensor", domain.SensorTypeMotion, "Hall de Entrada", "*")
	sensor2 := domain.NewSensor("sensor-002", "Living Room Motion", domain.SensorTypeMotion, "Sala de Estar", "*")
	sensor3 := domain.NewSensor("sensor-003", "Garage Motion", domain.SensorTypeMotion, "Garagem", "*")

	r.sensors["sensor-001"] = sensor1
	r.sensors["sensor-002"] = sensor2
	r.sensors["sensor-003"] = sensor3

	// Sensor com ID "1" para testes simples
	sensor4 := domain.NewSensor("1", "Test Motion Sensor", domain.SensorTypeMotion, "Test Location", "*")
	r.sensors["1"] = sensor4
}

// FindByID busca um sensor pelo ID
func (r *MemorySensorRepository) FindByID(id string) (*domain.Sensor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sensor, exists := r.sensors[id]
	if !exists {
		return nil, domain.ErrSensorNotFound
	}

	return sensor, nil
}

// FindByOwnerID busca sensores pelo ID do proprietário
func (r *MemorySensorRepository) FindByOwnerID(ownerID string) ([]*domain.Sensor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Sensor
	for _, sensor := range r.sensors {
		if sensor.OwnerID == ownerID {
			result = append(result, sensor)
		}
	}

	return result, nil
}

// Save persiste um sensor
func (r *MemorySensorRepository) Save(sensor *domain.Sensor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sensors[sensor.ID] = sensor
	return nil
}

// SaveMotionRecord persiste um registro de movimento
func (r *MemorySensorRepository) SaveMotionRecord(record *domain.MotionRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.motionRecords[record.SensorID] = append(r.motionRecords[record.SensorID], record)
	return nil
}

// GetMotionRecords retorna os registros de movimento de um sensor
func (r *MemorySensorRepository) GetMotionRecords(sensorID string, limit int) ([]*domain.MotionRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, exists := r.motionRecords[sensorID]
	if !exists {
		return []*domain.MotionRecord{}, nil
	}

	if limit > 0 && len(records) > limit {
		// Retornar os últimos 'limit' registros
		return records[len(records)-limit:], nil
	}

	return records, nil
}

// UpdateSensorOwner atualiza o proprietário de todos os sensores de teste
// Útil para testes de integração
func (r *MemorySensorRepository) UpdateSensorOwner(oldOwnerID, newOwnerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, sensor := range r.sensors {
		if sensor.OwnerID == oldOwnerID {
			sensor.OwnerID = newOwnerID
		}
	}
}

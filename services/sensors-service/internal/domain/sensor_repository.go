package domain

import "errors"

var (
	// ErrSensorNotFound indica que o sensor não foi encontrado
	ErrSensorNotFound = errors.New("sensor not found")

	// ErrSensorAccessDenied indica que o usuário não tem acesso ao sensor
	ErrSensorAccessDenied = errors.New("access denied to sensor")

	// ErrInvalidSensorID indica que o ID do sensor é inválido
	ErrInvalidSensorID = errors.New("invalid sensor ID")

	// ErrPublishFailed indica falha na publicação do evento
	ErrPublishFailed = errors.New("failed to publish event")
)

// SensorRepository define o contrato para persistência de sensores
type SensorRepository interface {
	// FindByID busca um sensor pelo ID
	FindByID(id string) (*Sensor, error)

	// FindByOwnerID busca sensores pelo ID do proprietário
	FindByOwnerID(ownerID string) ([]*Sensor, error)

	// Save persiste um sensor
	Save(sensor *Sensor) error

	// SaveMotionRecord persiste um registro de movimento
	SaveMotionRecord(record *MotionRecord) error

	// GetMotionRecords retorna os registros de movimento de um sensor
	GetMotionRecords(sensorID string, limit int) ([]*MotionRecord, error)
}

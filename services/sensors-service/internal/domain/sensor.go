package domain

import "time"

// SensorType define os tipos de sensores suportados
type SensorType string

const (
	SensorTypeMotion      SensorType = "motion"
	SensorTypeTemperature SensorType = "temperature"
	SensorTypeHumidity    SensorType = "humidity"
	SensorTypeDoor        SensorType = "door"
	SensorTypeWindow      SensorType = "window"
)

// Sensor representa a entidade de sensor no domínio
type Sensor struct {
	ID        string
	Name      string
	Type      SensorType
	Location  string
	OwnerID   string // UserID do proprietário
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSensor cria uma nova instância de Sensor
func NewSensor(id, name string, sensorType SensorType, location, ownerID string) *Sensor {
	now := time.Now()
	return &Sensor{
		ID:        id,
		Name:      name,
		Type:      sensorType,
		Location:  location,
		OwnerID:   ownerID,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// BelongsTo verifica se o sensor pertence ao usuário especificado
// O wildcard "*" permite acesso a qualquer usuário (para sensores de demonstração)
func (s *Sensor) BelongsTo(userID string) bool {
	return s.OwnerID == "*" || s.OwnerID == userID
}

// IsMotionSensor verifica se é um sensor de movimento
func (s *Sensor) IsMotionSensor() bool {
	return s.Type == SensorTypeMotion
}

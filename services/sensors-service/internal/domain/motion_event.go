package domain

import (
	"time"

	"github.com/google/uuid"
)

// MotionDetectedEvent representa um evento de detecção de movimento
type MotionDetectedEvent struct {
	ID         string    `json:"id"`
	SensorID   string    `json:"sensor_id"`
	UserID     string    `json:"user_id"`
	Location   string    `json:"location"`
	DetectedAt time.Time `json:"detected_at"`
	Intensity  float64   `json:"intensity"`
}

// NewMotionDetectedEvent cria um novo evento de movimento detectado
func NewMotionDetectedEvent(sensorID, userID, location string) *MotionDetectedEvent {
	return &MotionDetectedEvent{
		ID:         uuid.New().String(),
		SensorID:   sensorID,
		UserID:     userID,
		Location:   location,
		DetectedAt: time.Now(),
		Intensity:  1.0,
	}
}

// Topic retorna o tópico para publicação do evento
func (e *MotionDetectedEvent) Topic() string {
	return "sensors.motion.detected"
}

// MotionRecord representa um registro de movimento armazenado
type MotionRecord struct {
	ID         string
	SensorID   string
	UserID     string
	DetectedAt time.Time
}

// NewMotionRecord cria um novo registro de movimento
func NewMotionRecord(sensorID, userID string) *MotionRecord {
	return &MotionRecord{
		ID:         uuid.New().String(),
		SensorID:   sensorID,
		UserID:     userID,
		DetectedAt: time.Now(),
	}
}

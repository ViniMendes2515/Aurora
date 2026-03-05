package domain

import (
	"errors"
	"time"
)

var (
	ErrLightNotFound    = errors.New("light not found")
	ErrLightAccessDenied = errors.New("access denied to light")
	ErrInvalidLightID   = errors.New("invalid light ID")
	ErrDeviceUnreachable = errors.New("device unreachable")
)

// LightState representa o estado de uma luz
type LightState string

const (
	LightStateOn  LightState = "on"
	LightStateOff LightState = "off"
)

// Light representa um dispositivo de iluminação controlável
type Light struct {
	ID        string
	Name      string
	Location  string
	OwnerID   string
	State     LightState
	DeviceID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewLight cria uma nova instância de Light
func NewLight(id, name, location, ownerID, deviceID string) *Light {
	now := time.Now()
	return &Light{
		ID:        id,
		Name:      name,
		Location:  location,
		OwnerID:   ownerID,
		State:     LightStateOff,
		DeviceID:  deviceID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// TurnOn acende a luz
func (l *Light) TurnOn() {
	l.State = LightStateOn
	l.UpdatedAt = time.Now()
}

// TurnOff apaga a luz
func (l *Light) TurnOff() {
	l.State = LightStateOff
	l.UpdatedAt = time.Now()
}

// IsOn retorna se a luz está acesa
func (l *Light) IsOn() bool {
	return l.State == LightStateOn
}

// BelongsTo verifica se a luz pertence ao usuário
func (l *Light) BelongsTo(userID string) bool {
	return l.OwnerID == "*" || l.OwnerID == userID
}

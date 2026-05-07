package domain

import "time"

// LightStateChangedEvent é o domain event emitido quando uma luz muda de estado.
type LightStateChangedEvent struct {
	LightID   string    `json:"light_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	State     string    `json:"state"`
	ChangedAt time.Time `json:"changed_at"`
}

func (e *LightStateChangedEvent) Topic() string {
	return "lighting.state.changed"
}

// NewLightStateChangedEvent constrói um domain event de mudança de estado.
func NewLightStateChangedEvent(lightID, userID, name, location, state string) *LightStateChangedEvent {
	return &LightStateChangedEvent{
		LightID:   lightID,
		UserID:    userID,
		Name:      name,
		Location:  location,
		State:     state,
		ChangedAt: time.Now(),
	}
}
